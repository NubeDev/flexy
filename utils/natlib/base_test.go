package natlib

import (
	"fmt"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	nl := New(NewOpts{})

	// Subscribing with respond
	responses, err := nl.RequestAll("test", []byte("ping"), 2*time.Second)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Received %d responses:\n", len(responses))
	for i, msg := range responses {
		fmt.Printf("Response %d from %s: %s\n", i+1, msg.Subject, string(msg.Data))
	}

}
