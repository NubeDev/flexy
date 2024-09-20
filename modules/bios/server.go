package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func (s *Service) SetupGin() *gin.Engine {
	r := gin.Default()
	// Add the POST proxy route
	r.POST("/api/upload/:uuid", s.natsProxyFileUpload)
	r.POST("/api/proxy/*topic", s.natsProxyHandler)

	return r
}

func (s *Service) extractTopicAndTimeout(c *gin.Context) (string, time.Duration) {
	// Extract the NATS topic from the URL
	topic := strings.Replace(c.Param("topic")[1:], "/", ".", -1)

	// Read the timeout from the header, default to 5 seconds if not provided
	timeoutHeader := c.GetHeader("X-Timeout")
	timeout := 5 * time.Second // Default timeout
	if timeoutHeader != "" {
		parsedTimeout, err := time.ParseDuration(timeoutHeader)
		if err == nil {
			timeout = parsedTimeout
		}
	}

	return topic, timeout
}

// File upload handler for Gin to NATS proxy
func (s *Service) natsProxyFileUpload(c *gin.Context) {

	timeoutHeader := c.GetHeader("X-Timeout")
	timeout := 5 * time.Second // Default timeout
	if timeoutHeader != "" {
		parsedTimeout, err := time.ParseDuration(timeoutHeader)
		if err == nil {
			timeout = parsedTimeout
		}
	}

	// Get the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
		return
	}

	// Open the uploaded file
	uploadedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to open the uploaded file"})
		return
	}
	defer uploadedFile.Close()

	// Read the file contents
	fileBytes, err := ioutil.ReadAll(uploadedFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the file content"})
		return
	}

	// Base64 encode the file data
	encodedData := base64.StdEncoding.EncodeToString(fileBytes)

	storeName := c.GetHeader("StoreName")
	if storeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "StoreName header is required"})
		return
	}

	objectName := c.GetHeader("ObjectName")
	if objectName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ObjectName header is required"})
		return
	}
	var storeRequest StoreRequest
	// Set the action to "add.object" and the base64-encoded data
	storeRequest.StoreName = storeName
	storeRequest.ObjectName = objectName
	storeRequest.Action = "add.object"
	storeRequest.Data = encodedData

	// Marshal the StoreRequest into JSON
	storeRequestData, err := json.Marshal(storeRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request data"})
		return
	}

	// Create a NATS message with the store request as data
	msg := &nats.Msg{
		Subject: fmt.Sprintf("%s.post.system.store.add.object", c.Param("uuid")),
		Data:    storeRequestData, // JSON-encoded StoreRequest
		Header:  nats.Header{},
	}

	// Optionally, set NATS headers with additional file metadata
	msg.Header.Set("FileName", file.Filename)
	msg.Header.Set("FileSize", fmt.Sprintf("%d", file.Size))

	// Send the request to NATS with a timeout
	response, err := s.natsConn.RequestMsg(msg, timeout)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "NATS request timeout or error", "details": err.Error()})
		return
	}

	// Return the response from NATS as the Gin response
	c.Data(http.StatusOK, "application/json", response.Data)
}

// Handler for the Gin-to-NATS proxy
func (s *Service) natsProxyHandler(c *gin.Context) {
	// Use the helper function to extract topic and timeout
	topic, timeout := s.extractTopicAndTimeout(c)

	// Read the POST request body
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create a NATS message with headers
	msg := &nats.Msg{
		Subject: topic,
		Data:    body,
		Header:  nats.Header{},
	}

	// Set the ReturnFormat header in the NATS message
	debug := c.GetHeader("Debug")
	if debug == "true" {
		msg.Header.Set("Debug", "debug")
	}

	// Send the message over NATS with a timeout and wait for a response
	response, err := s.natsConn.RequestMsg(msg, timeout)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "NATS request timeout or error", "details": err.Error()})
		return
	}

	// Return the response from NATS as the Gin response
	c.Data(http.StatusOK, "application/json", response.Data)
}
