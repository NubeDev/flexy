package natlib

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Opts represents options for NATS operations.
type Opts struct {
	Timeout int // seconds
}

// Subjects represents a NATS subject with its type.
type Subjects struct {
	Type    string // "Subscribe" or "SubscribeWithRespond"
	Subject string
}

// NatLib is the common interface with NATS methods.
type NatLib interface {
	Publish(subj string, data []byte, cb nats.MsgHandler, opts *Opts) error
	Subscribe(subj string, cb nats.MsgHandler, opts *Opts) error
	SubscribeWithRespond(subj string, handler func(msg *nats.Msg) ([]byte, error), opts *Opts) error
	RequestAll(subj string, data []byte, timeout time.Duration) ([]*nats.Msg, error)
	Close() // close server

	// JetStream Object Store methods
	CreateObjectStore(storeName string, config *nats.ObjectStoreConfig) error
	NewObject(storeName, objectName, filePath string, overwriteIfExisting bool) error
	PutBytes(storeName, objectName string, data []byte, overwriteIfExisting bool) error
	GetStoreObjects(storeName string) ([]*nats.ObjectInfo, error)
	GetStores() ([]string, error)
	GetStore(name string) (nats.ObjectStore, error)
	GetObject(storeName string, objectName string) ([]byte, error)
	DeleteObject(storeName string, objectName string) error
	DropStore(storeName string) error
	DownloadObject(storeName string, objectName string, destinationPath string) error
}

var uuidName = "Global-UUID"

// natsLib implements the NatLib interface.
type natsLib struct {
	nc               *nats.Conn
	JetStreamContext nats.JetStreamContext
	subjects         []*Subjects
	globalUUID       string
}

type NewOpts struct {
	URL             string
	GlobalUUID      string
	NatsConn        *nats.Conn
	EnableJetStream bool
}

// New creates a new instance of NatLib.
func New(opts NewOpts) NatLib {
	var url = opts.URL
	if url == "" {
		url = nats.DefaultURL
	}
	var err error
	var nc *nats.Conn
	if opts.NatsConn == nil {
		nc, err = nats.Connect(url)
		if err != nil {
			panic(err)
		}
	} else {
		nc = opts.NatsConn
	}
	n := &natsLib{
		nc:       nc,
		subjects: []*Subjects{},
	}
	if opts.EnableJetStream {
		js, err := nc.JetStream()
		if err != nil {
			log.Fatal().Msgf("Error initializing JetStream: %v", err)
		}
		n.JetStreamContext = js
	}

	return n
}

// Close will close the connection to the server.
func (nl *natsLib) Close() {
	nl.nc.Close()
}

// Publish sends a request and waits for a response.
func (nl *natsLib) Publish(subj string, data []byte, cb nats.MsgHandler, opts *Opts) error {
	timeout := 5 * time.Second
	if opts != nil && opts.Timeout > 0 {
		timeout = time.Duration(opts.Timeout) * time.Second
	}
	msg, err := nl.nc.Request(subj, data, timeout)
	if err != nil {
		return err
	}
	if nl.globalUUID != "" {
		msg.Header.Set(uuidName, nl.globalUUID)
	}
	cb(msg)
	return nil
}

// Subscribe listens for messages on a subject.
func (nl *natsLib) Subscribe(subj string, cb nats.MsgHandler, opts *Opts) error {
	nl.subjects = append(nl.subjects, &Subjects{
		Type:    "subscribe",
		Subject: subj,
	})
	_, err := nl.nc.Subscribe(subj, cb)
	return err
}

// SubscribeWithRespond subscribes and responds to incoming messages.
func (nl *natsLib) SubscribeWithRespond(subj string, handler func(msg *nats.Msg) ([]byte, error), opts *Opts) error {
	nl.subjects = append(nl.subjects, &Subjects{
		Type:    "subscribeWithRespond",
		Subject: subj,
	})

	_, err := nl.nc.Subscribe(subj, func(msg *nats.Msg) {
		responseData, err := handler(msg)
		if err != nil {
			// Handle error appropriately
			return
		}
		if nl.globalUUID != "" {
			msg.Header.Set(uuidName, nl.globalUUID)
		}
		if err := msg.Respond(responseData); err != nil {
			// Handle error appropriately
		}
	})
	return err
}

// RequestAll sends a request to the subject and collects all responses
// received within the specified timeout duration.
func (nl *natsLib) RequestAll(subj string, data []byte, timeout time.Duration) ([]*nats.Msg, error) {
	// Create a unique inbox
	inbox := nats.NewInbox()
	sub, err := nl.nc.SubscribeSync(inbox)
	if err != nil {
		return nil, err
	}
	defer sub.Unsubscribe()

	// Publish the message with the reply set to the inbox
	m := &nats.Msg{
		Subject: subj,
		Reply:   inbox,
		Data:    data,
	}
	if nl.globalUUID != "" {
		m.Header.Set(uuidName, nl.globalUUID)
	}
	err = nl.nc.PublishMsg(m)
	if err != nil {
		return nil, err
	}

	// Buffer to store the responses
	var responses []*nats.Msg

	// Collect responses until timeout
	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		msg, err := sub.NextMsg(remaining)
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				// Timeout reached, exit loop
				break
			}
			return nil, err
		}
		responses = append(responses, msg)
	}

	return responses, nil
}

// CreateObjectStore creates an object store with the given name and configuration.
func (nl *natsLib) CreateObjectStore(storeName string, config *nats.ObjectStoreConfig) error {
	// Ensure the object store exists or create one
	_, err := nl.JetStreamContext.ObjectStore(storeName)
	if err != nil {
		if config == nil {
			config = &nats.ObjectStoreConfig{
				Bucket: storeName,
			}
		}
		_, err = nl.JetStreamContext.CreateObjectStore(config)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewObject creates a new object in the object store.
// If overwriteIfExisting is true, it will delete the existing object and add the new one.
func (nl *natsLib) NewObject(storeName, objectName, filePath string, overwriteIfExisting bool) error {
	store, err := nl.GetStore(storeName)
	if err != nil {
		return err
	}

	// Check if object exists
	obj, err := store.Get(objectName)
	if err == nil {
		obj.Close() // Close the object if it exists
		if overwriteIfExisting {
			// Delete the existing object
			err = store.Delete(objectName)
			if err != nil {
				log.Error().Msgf("Error deleting existing object %s: %v", objectName, err)
				return err
			}
			log.Info().Msgf("Existing object %s deleted", objectName)
		} else {
			log.Info().Msgf("Object %s already exists, not overwriting", objectName)
			return nil
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Error().Msgf("Error opening file: %v", err)
		return err
	}
	defer file.Close()

	// Upload the file to the Object Store using an io.Reader
	_, err = store.Put(&nats.ObjectMeta{Name: objectName}, file)
	if err != nil {
		log.Error().Msgf("Error uploading file to object store: %v", err)
		return err
	}
	log.Info().Msgf("Object %s added successfully", objectName)
	return nil
}

// PutBytes adds a new object (as bytes) to the object store.
// If overwriteIfExisting is true, it will delete the existing object and add the new one.
func (nl *natsLib) PutBytes(storeName, objectName string, data []byte, overwriteIfExisting bool) error {
	// Retrieve the object store
	store, err := nl.GetStore(storeName)
	if err != nil {
		return err
	}

	// Check if the object exists
	obj, err := store.Get(objectName)
	if err == nil {
		obj.Close() // Close the object if it exists
		if overwriteIfExisting {
			// Delete the existing object
			err = store.Delete(objectName)
			if err != nil {
				log.Error().Msgf("Error deleting existing object %s: %v", objectName, err)
				return err
			}
			log.Info().Msgf("Existing object %s deleted", objectName)
		} else {
			// If overwrite is not allowed, return without overwriting
			log.Info().Msgf("Object %s already exists, not overwriting", objectName)
			return nil
		}
	}

	// Add the new object as bytes
	_, err = store.PutBytes(objectName, data)
	if err != nil {
		log.Error().Msgf("Error putting object %s: %v", objectName, err)
		return err
	}

	log.Info().Msgf("Object %s added successfully", objectName)
	return nil
}

// GetStoreObjects retrieves details for all objects in the specified object store.
func (nl *natsLib) GetStoreObjects(storeName string) ([]*nats.ObjectInfo, error) {
	store, err := nl.GetStore(storeName)
	if err != nil {
		return nil, err
	}
	return store.List()
}

// GetStores returns the list of available object store names.
func (nl *natsLib) GetStores() ([]string, error) {
	storeNamesChan := nl.JetStreamContext.ObjectStoreNames()
	var stores []string
	for storeName := range storeNamesChan {
		stores = append(stores, storeName)
	}
	return stores, nil
}

// GetStore returns the ObjectStore for a specific name.
func (nl *natsLib) GetStore(name string) (nats.ObjectStore, error) {
	store, err := nl.JetStreamContext.ObjectStore(name)
	if err != nil {
		log.Error().Msgf("Error getting object store %s: %v", name, err)
		return nil, err
	}
	return store, nil
}

// GetObject retrieves an object by name from the object store.
func (nl *natsLib) GetObject(storeName string, objectName string) ([]byte, error) {
	store, err := nl.GetStore(storeName)
	if err != nil {
		return nil, err
	}

	obj, err := store.Get(objectName)
	if err != nil {
		log.Error().Msgf("Error getting object %s: %v", objectName, err)
		return nil, err
	}
	defer obj.Close()

	// Read the object data using io.ReadAll
	data, err := io.ReadAll(obj)
	if err != nil {
		log.Error().Msgf("Error reading object %s: %v", objectName, err)
		return nil, err
	}

	return data, nil
}

// DeleteObject removes an object from the object store by name.
func (nl *natsLib) DeleteObject(storeName string, objectName string) error {
	store, err := nl.GetStore(storeName)
	if err != nil {
		return err
	}
	err = store.Delete(objectName)
	if err != nil {
		log.Error().Msgf("Error deleting object %s: %v", objectName, err)
		return err
	}
	return nil
}

// DropStore deletes the entire object store by name.
func (nl *natsLib) DropStore(storeName string) error {
	err := nl.JetStreamContext.DeleteObjectStore(storeName)
	if err != nil {
		log.Error().Msgf("Error dropping object store %s: %v", storeName, err)
		return err
	}
	log.Info().Msgf("Object store %s deleted successfully", storeName)
	return nil
}

// DownloadObject downloads an object from the object store and saves it to the specified destination directory.
// The object will be saved with its original objectName in the destination directory.
func (nl *natsLib) DownloadObject(storeName string, objectName string, destinationPath string) error {
	// Get the object store
	store, err := nl.GetStore(storeName)
	if err != nil {
		return err
	}

	// Get the object
	obj, err := store.Get(objectName)
	if err != nil {
		log.Error().Msgf("Error getting object %s: %v", objectName, err)
		return err
	}
	defer obj.Close()

	// Ensure destinationPath is a directory
	destInfo, err := os.Stat(destinationPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create the directory
			err = os.MkdirAll(destinationPath, os.ModePerm)
			if err != nil {
				log.Error().Msgf("Error creating destination directory %s: %v", destinationPath, err)
				return err
			}
		} else {
			log.Error().Msgf("Error accessing destination path %s: %v", destinationPath, err)
			return err
		}
	} else if !destInfo.IsDir() {
		log.Error().Msgf("Destination path %s is not a directory", destinationPath)
		return fmt.Errorf("destination path %s is not a directory", destinationPath)
	}

	// Construct the full file path
	destinationFilePath := filepath.Join(destinationPath, objectName)

	// Open the destination file for writing
	outFile, err := os.Create(destinationFilePath)
	if err != nil {
		log.Error().Msgf("Error creating destination file %s: %v", destinationFilePath, err)
		return err
	}
	defer outFile.Close()

	// Copy the data from the object to the file
	_, err = io.Copy(outFile, obj)
	if err != nil {
		log.Error().Msgf("Error writing to destination file %s: %v", destinationFilePath, err)
		return err
	}

	log.Info().Msgf("Object %s downloaded successfully to %s", objectName, destinationFilePath)
	return nil
}

type Response struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	Payload     string `json:"payload"`
	Description string `json:"description,omitempty"`
}

type Args struct {
	Description string `json:"description"`
}

func NewResponse(messageCode int, payload string, args ...Args) *Response {
	resp := &Response{
		Code:    messageCode,
		Message: code.GetMsg(messageCode),
		Payload: payload,
	}
	if len(args) > 0 {
		resp.Description = args[0].Description
	}
	return resp
}

func (p *Response) ToJSONError() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Response) ToJSON() []byte {
	jsonData, _ := json.Marshal(p)
	return jsonData
}

func (p *Response) ToJSONString() (string, error) {
	jsonData, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
