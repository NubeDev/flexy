package natsrouter

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"testing"
)

func TestGetObject(t *testing.T) {
	// Setup NATS connection
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("Error connecting to NATS: %v", err)
	}
	defer nc.Close()

	// Create a NatsRouter
	router := New(nc)

	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		t.Fatalf("Error creating JetStream context: %v", err)
	}
	router.js = js

	// Define test object store and object name
	storeName := "mystore"
	objectName := "testobject"
	testData := []byte("Hello, NATS Object Store!")

	err = router.CreateObjectStore(storeName, nil)
	if err != nil {
		t.Fatalf("Error creating object store: %v", err)
		return
	}

	err = router.PutBytes(storeName, objectName, testData, false)
	if err != nil {
		t.Fatalf("Error PutBytes: %v", err)
		return
	}

	objects, err := router.GetStoreObjects(storeName)
	if err != nil {
		return
	}

	for _, object := range objects {
		fmt.Println(object.Name, object.Size)
	}

	var d = false
	if d {
		err = router.DeleteObject(storeName, objectName)
		if err != nil {
			t.Fatalf("Error PutBytes: %v", err)
			return
		}
	}

}
