package main

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// ReadFile reads the contents of a file
func (inst *Service) ReadFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MakeDir creates a new directory
func (inst *Service) MakeDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// DeleteDir deletes an existing directory
func (inst *Service) DeleteDir(path string) error {
	return os.RemoveAll(path)
}

// ZipFolder zips the contents of a folder
func (inst *Service) ZipFolder(srcDir, dstZip string) error {
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
func (inst *Service) UnzipFolder(srcZip, destDir string) error {
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
