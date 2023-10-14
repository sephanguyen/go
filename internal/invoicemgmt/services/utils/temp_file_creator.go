package utils

import (
	"fmt"
	"os"
)

type TempFile struct {
	TempDirName string
	ObjectPath  string
	ObjectName  string
	File        *os.File
}

func (tf *TempFile) Close() error {
	return tf.File.Close()
}

func (tf *TempFile) CleanUp() error {
	return os.RemoveAll(tf.TempDirName)
}

type ITempFileCreator interface {
	CreateTempFile(objectName string) (*TempFile, error)
}

type TempFileCreator struct {
	TempDirPattern string
}

func (t *TempFileCreator) CreateTempFile(objectName string) (*TempFile, error) {
	tempDir, err := os.MkdirTemp("", t.TempDirPattern)
	if err != nil {
		return nil, err
	}

	objectPath := fmt.Sprintf("%v/%v", tempDir, objectName)
	file, err := os.Create(objectPath)
	if err != nil {
		return nil, fmt.Errorf("os.Create err: %v", err)
	}

	return &TempFile{
		TempDirName: tempDir,
		ObjectName:  objectName,
		ObjectPath:  objectPath,
		File:        file,
	}, nil
}
