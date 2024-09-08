package natsrouter

import (
	"github.com/nats-io/nats.go"
	"io"
	"log"
	"os"
	"path/filepath"
)

// CreateObjectStore will create an object store.
func (r *NatsRouter) CreateObjectStore(storeName string, config *nats.ObjectStoreConfig) error {
	// Ensure the object store exists or create one
	_, err := r.js.ObjectStore(storeName)
	if err != nil {
		if config == nil {
			config = &nats.ObjectStoreConfig{
				Bucket: storeName,
			}
		}
		_, err = r.js.CreateObjectStore(config)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewObject creates a new object in the object store
// If overwriteIfExisting is true, it will delete the existing object and add the new one
func (r *NatsRouter) NewObject(storeName, objectName, filePath string, overwriteIfExisting bool) error {
	store, err := r.GetStore(storeName)
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
				log.Printf("Error deleting existing object %s: %v", objectName, err)
				return err
			}
			log.Printf("Existing object %s deleted", objectName)
		} else {
			log.Printf("Object %s already exists, not overwriting", objectName)
			return nil
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return err
	}
	defer file.Close()

	// Get the file name
	fileName := filepath.Base(filePath)

	// Upload the file to the Object Store using an io.Reader
	_, err = store.Put(&nats.ObjectMeta{Name: fileName}, file)
	if err != nil {
		log.Printf("Error uploading file to object store: %v", err)
		return err
	}
	log.Printf("Object %s added successfully", objectName)
	return nil
}

// PutBytes adds a new object (as bytes) to the object store
// If overwriteIfExisting is true, it will delete the existing object and add the new one
func (r *NatsRouter) PutBytes(storeName, objectName string, data []byte, overwriteIfExisting bool) error {
	// Retrieve the object store
	store, err := r.GetStore(storeName)
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
				log.Printf("Error deleting existing object %s: %v", objectName, err)
				return err
			}
			log.Printf("Existing object %s deleted", objectName)
		} else {
			// If overwrite is not allowed, return without overwriting
			log.Printf("Object %s already exists, not overwriting", objectName)
			return nil
		}
	}

	// Add the new object as bytes
	_, err = store.PutBytes(objectName, data)
	if err != nil {
		log.Printf("Error putting object %s: %v", objectName, err)
		return err
	}

	log.Printf("Object %s added successfully", objectName)
	return nil
}

// GetStoreObjects retrieves details for all objects in the specified object store
func (r *NatsRouter) GetStoreObjects(storeName string) ([]*nats.ObjectInfo, error) {
	store, err := r.GetStore(storeName)
	if err != nil {
		return nil, err
	}
	return store.List()
}

// GetStores returns the list of available object store names
func (r *NatsRouter) GetStores() []string {
	storeNamesChan := r.js.ObjectStoreNames()

	var stores []string
	for store := range storeNamesChan {
		stores = append(stores, store)
	}

	return stores
}

// GetStore returns the ObjectStore for a specific name
func (r *NatsRouter) GetStore(name string) (nats.ObjectStore, error) {
	store, err := r.js.ObjectStore(name)
	if err != nil {
		log.Printf("Error getting object store %s: %v", name, err)
		return nil, err
	}
	return store, nil
}

// GetObject retrieves an object by name from the object store
func (r *NatsRouter) GetObject(storeName string, objectName string) ([]byte, error) {
	store, err := r.GetStore(storeName)
	if err != nil {
		return nil, err
	}

	obj, err := store.Get(objectName)
	if err != nil {
		log.Printf("Error getting object %s: %v", objectName, err)
		return nil, err
	}
	defer obj.Close()
	// Read the object data using io.ReadAll
	data, err := io.ReadAll(obj)
	if err != nil {
		log.Printf("Error reading object %s: %v", objectName, err)
		return nil, err
	}

	return data, nil
}

// DeleteObject removes an object from the object store by name
func (r *NatsRouter) DeleteObject(storeName string, objectName string) error {
	store, err := r.GetStore(storeName)
	if err != nil {
		return err
	}
	err = store.Delete(objectName)
	if err != nil {
		log.Printf("Error deleting object %s: %v", objectName, err)
		return err
	}

	return nil
}

// DropStore deletes the entire object store by name
func (r *NatsRouter) DropStore(storeName string) error {
	err := r.js.DeleteObjectStore(storeName)
	if err != nil {
		log.Printf("Error dropping object store %s: %v", storeName, err)
		return err
	}
	log.Printf("Object store %s deleted successfully", storeName)
	return nil
}
