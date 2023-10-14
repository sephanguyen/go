package utils

import (
	"io"
	"os"
	"testing"
)

func TestTempFileCreator_WriteFromTempFile(t *testing.T) {

	tests := []struct {
		name         string
		dirPattern   string
		objectName   string
		wantErr      bool
		osFileMethod string
	}{
		{
			name:         "success writing on temp file",
			dirPattern:   "invoicemgmt-unit-test",
			objectName:   "test.csv",
			wantErr:      false,
			osFileMethod: "OpenFile",
		},
		{
			name:         "failed writing on temp file",
			dirPattern:   "invoicemgmt-unit-test-2",
			objectName:   "test2.csv",
			wantErr:      true,
			osFileMethod: "Open",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TempFileCreator{
				TempDirPattern: tt.dirPattern,
			}
			tempFileOne, err := tr.CreateTempFile(tt.objectName)
			if err != nil {
				t.Errorf("TempFileCreator.CreateTempFile() error = %v", err)
				return
			}

			var destinationFile *os.File

			if tt.osFileMethod == "Open" {
				destinationFile, err = os.Open(tempFileOne.ObjectPath)
			} else {
				destinationFile, err = os.OpenFile(tempFileOne.ObjectPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			}
			if err != nil {
				t.Errorf("os method %v error = %v, wantErr %v", tt.osFileMethod, err, tt.wantErr)
				return
			}

			tempFileTwo, err := tr.CreateTempFile(tt.objectName)
			if err != nil {
				t.Errorf("TempFileCreator.CreateTempFile() error = %v", err)
				return
			}

			_, err = io.Copy(destinationFile, tempFileTwo.File)
			if (err != nil) != tt.wantErr {
				t.Errorf("io.Copy err: %v, wantErr %v", err, tt.wantErr)
				return
			}

			destinationFile.Close()
			tempFileOne.Close()
			tempFileTwo.Close()
			tempFileOne.CleanUp()
			tempFileTwo.CleanUp()
		})
	}
}
