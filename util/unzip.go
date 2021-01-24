package util

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UnzipFileFromMemory unzips a byte slice into a destination folder
// Source: https://github.com/artdarek/go-unzip
func UnzipFileFromMemory(src []byte, destFolder string) ([]string, error) {

	zipReader, err := zip.NewReader(bytes.NewReader(src), int64(len(src)))

	fileList := make([]string, 0)

	if err != nil {
		return fileList, fmt.Errorf("Unable to read byte stream %v", err.Error())
	}

	for _, f := range zipReader.File {
		fullFilePath, err := extractAndWriteFile(f, destFolder)
		if err != nil {
			return fileList, fmt.Errorf("Error extracting file '%v': %v", f.Name, err.Error())
		}

		fileList = append(fileList, fullFilePath)
	}

	return fileList, nil
}

func extractAndWriteFile(f *zip.File, destFolder string) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", fmt.Errorf("Error opening file '%v': %v", f.Name, err.Error())
	}
	defer func() {
		if err := rc.Close(); err != nil {
			panic(err)
		}
	}()

	localFullPath := filepath.Join(destFolder, f.Name)

	if !strings.HasPrefix(localFullPath, filepath.Clean(destFolder)+string(os.PathSeparator)) {
		return "", fmt.Errorf("%s: Illegal file path", localFullPath)
	}

	if f.FileInfo().IsDir() {
		os.MkdirAll(localFullPath, f.Mode())
	} else {
		os.MkdirAll(filepath.Dir(localFullPath), f.Mode())
		f, err := os.OpenFile(localFullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", fmt.Errorf("Unable to open file '%v' for writing", localFullPath)
		}
		defer func() {
			if err := f.Close(); err != nil {
				panic(err)
			}
		}()

		_, err = io.Copy(f, rc)
		if err != nil {
			return "", fmt.Errorf("Unable to write file '%v'", localFullPath)
		}
	}
	return localFullPath, nil
}
