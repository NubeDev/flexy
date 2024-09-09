package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/nats-io/nats.go"
)

// Command structure to decode the incoming JSON
type Command struct {
	Command string                 `json:"command"`
	Body    map[string]interface{} `json:"body"`
}

// Service struct to handle NATS and file operations
type Service struct {
	natsConn *nats.Conn
}

// NewService initializes the NATS connection and returns the Service
func NewService(natsURL string) (*Service, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}
	return &Service{natsConn: nc}, nil
}

// ReadFile reads the contents of a file
func (s *Service) ReadFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MakeDir creates a new directory
func (s *Service) MakeDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// DeleteDir deletes an existing directory
func (s *Service) DeleteDir(path string) error {
	return os.RemoveAll(path)
}

// ZipFolder zips the contents of a folder
func (s *Service) ZipFolder(srcDir, dstZip string) error {
	zipFile, err := os.Create(dstZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath := strings.TrimPrefix(path, filepath.Dir(srcDir)+"/")
		zipFile, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		fileContent, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fileContent.Close()

		_, err = io.Copy(zipFile, fileContent)
		return err
	})
	return err
}

// UnzipFolder extracts a zip archive into a destination folder
func (s *Service) UnzipFolder(srcZip, destDir string) error {
	zipReader, err := zip.OpenReader(srcZip)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		fpath := filepath.Join(destDir, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		destFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		fileInArchive, err := file.Open()
		if err != nil {
			destFile.Close()
			return err
		}

		_, err = io.Copy(destFile, fileInArchive)
		destFile.Close()
		fileInArchive.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// HandleCommand processes incoming JSON commands
func (s *Service) HandleCommand(m *nats.Msg) {
	var cmd Command

	// Unmarshal the received JSON message
	if err := json.Unmarshal(m.Data, &cmd); err != nil {
		s.natsConn.Publish(m.Reply, []byte(fmt.Sprintf("Invalid JSON format: %v", err)))
		return
	}

	switch cmd.Command {
	case "read_file":
		path, ok := cmd.Body["path"].(string)
		if !ok || path == "" {
			s.natsConn.Publish(m.Reply, []byte("Error: 'path' is required for read_file"))
			return
		}
		content, err := s.ReadFile(path)
		if err != nil {
			s.natsConn.Publish(m.Reply, []byte(fmt.Sprintf("Error reading file: %v", err)))
		} else {
			s.natsConn.Publish(m.Reply, []byte(content))
		}
	case "make_dir":
		path, ok := cmd.Body["path"].(string)
		if !ok || path == "" {
			s.natsConn.Publish(m.Reply, []byte("Error: 'path' is required for make_dir"))
			return
		}
		err := s.MakeDir(path)
		if err != nil {
			s.natsConn.Publish(m.Reply, []byte(fmt.Sprintf("Error creating directory: %v", err)))
		} else {
			s.natsConn.Publish(m.Reply, []byte("Directory created"))
		}
	case "delete_dir":
		path, ok := cmd.Body["path"].(string)
		if !ok || path == "" {
			s.natsConn.Publish(m.Reply, []byte("Error: 'path' is required for delete_dir"))
			return
		}
		err := s.DeleteDir(path)
		if err != nil {
			s.natsConn.Publish(m.Reply, []byte(fmt.Sprintf("Error deleting directory: %v", err)))
		} else {
			s.natsConn.Publish(m.Reply, []byte("Directory deleted"))
		}
	case "zip_folder":
		srcDir, srcOk := cmd.Body["srcDir"].(string)
		dstZip, dstOk := cmd.Body["dstZip"].(string)
		if !srcOk || !dstOk || srcDir == "" || dstZip == "" {
			s.natsConn.Publish(m.Reply, []byte("Error: 'srcDir' and 'dstZip' are required for zip_folder"))
			return
		}
		err := s.ZipFolder(srcDir, dstZip)
		if err != nil {
			s.natsConn.Publish(m.Reply, []byte(fmt.Sprintf("Error zipping folder: %v", err)))
		} else {
			s.natsConn.Publish(m.Reply, []byte("Folder zipped"))
		}
	case "unzip_folder":
		srcZip, srcOk := cmd.Body["srcZip"].(string)
		destDir, dstOk := cmd.Body["destDir"].(string)
		if !srcOk || !dstOk || srcZip == "" || destDir == "" {
			s.natsConn.Publish(m.Reply, []byte("Error: 'srcZip' and 'destDir' are required for unzip_folder"))
			return
		}
		err := s.UnzipFolder(srcZip, destDir)
		if err != nil {
			s.natsConn.Publish(m.Reply, []byte(fmt.Sprintf("Error unzipping folder: %v", err)))
		} else {
			s.natsConn.Publish(m.Reply, []byte("Folder unzipped"))
		}
	default:
		s.natsConn.Publish(m.Reply, []byte("Unknown command"))
	}
}

// StartService starts the NATS subscription and listens for commands
func (s *Service) StartService() {
	s.natsConn.Subscribe("service.command", s.HandleCommand)
}
