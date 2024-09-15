package main

import (
	"archive/zip"
	"fmt"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/nats-io/nats.go"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func (s *Service) handleReadFile(m *nats.Msg) {
	cmd, err := s.DecodeCommand(m)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	path, err := s.GetCommandValue(cmd, "path")
	if err != nil {
		s.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	content, err := s.ReadFile(path)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error reading file: %v", err))
	} else {
		s.publish(m.Reply, content, code.SUCCESS)
	}
}

func (s *Service) handleMakeDir(m *nats.Msg) {
	cmd, err := s.DecodeCommand(m)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	path, err := s.GetCommandValue(cmd, "path")
	if err != nil {
		s.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	err = s.MakeDir(path)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error creating directory: %v", err))
	} else {
		s.publish(m.Reply, "Directory created", code.SUCCESS)
	}
}

func (s *Service) handleDeleteDir(m *nats.Msg) {
	cmd, err := s.DecodeCommand(m)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	path, err := s.GetCommandValue(cmd, "path")
	if err != nil {
		s.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	err = s.DeleteDir(path)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error deleting directory: %v", err))
	} else {
		s.publish(m.Reply, "Directory deleted", code.SUCCESS)
	}
}

func (s *Service) handleZipFolder(m *nats.Msg) {
	cmd, err := s.DecodeCommand(m)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	srcDir, err := s.GetCommandValue(cmd, "srcDir")
	if err != nil {
		s.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	dstZip, err := s.GetCommandValue(cmd, "dstZip")
	if err != nil {
		s.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	err = s.ZipFolder(srcDir, dstZip)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error zipping folder: %v", err))
	} else {
		s.publish(m.Reply, "Folder zipped", code.SUCCESS)
	}
}

func (s *Service) handleUnzipFolder(m *nats.Msg) {
	cmd, err := s.DecodeCommand(m)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	srcZip, err := s.GetCommandValue(cmd, "srcZip")
	if err != nil {
		s.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	destDir, err := s.GetCommandValue(cmd, "destDir")
	if err != nil {
		s.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	err = s.UnzipFolder(srcZip, destDir)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error unzipping folder: %v", err))
	} else {
		s.publish(m.Reply, "Folder unzipped", code.SUCCESS)
	}
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
