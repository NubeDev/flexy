package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

// NATSRequest defines the expected JSON body for the REST request
type NATSRequest struct {
	Subject string      `json:"subject"`
	Payload interface{} `json:"payload"`
	Timeout int         `json:"timeout"` // Timeout in seconds
}

func main() {
	// Initialize NATS connection
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// Initialize Gin router
	r := gin.Default()

	// POST /nats-proxy endpoint
	r.POST("/nats-proxy", func(c *gin.Context) {
		// Parse incoming request body
		var req NATSRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Marshal the payload to JSON (string or map accepted)
		payload, err := json.Marshal(req.Payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal payload"})
			return
		}

		// Set timeout for the NATS request
		timeout := time.Duration(req.Timeout) * time.Second
		msg, err := nc.Request(req.Subject, payload, timeout)
		if err != nil {
			c.JSON(http.StatusRequestTimeout, gin.H{"error": fmt.Sprintf("NATS request failed: %v", err)})
			return
		}

		var jsonResponse interface{}
		err = json.Unmarshal(msg.Data, &jsonResponse)
		if err != nil {
			// If it's not valid JSON, return the response as a plain string
			c.JSON(http.StatusOK, gin.H{
				"response": string(msg.Data),
			})
			return
		}

		// If it's valid JSON, return the JSON object
		c.JSON(http.StatusOK, gin.H{
			"response": jsonResponse,
		})
	})

	// Start the Gin server
	r.Run(":9090")
}
