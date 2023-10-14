package fileio

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"
)

type FileUtils struct{}

func (r *FileUtils) GetFileNamesOnDir(filename string) ([]string, error) {
	files, error := ioutil.ReadDir(filename)
	if error != nil {
		return nil, error
	}
	filenames := []string{}
	for _, file := range files {
		if !file.IsDir() {
			filenames = append(filenames, file.Name())
		}
	}
	return filenames, nil
}

func (r *FileUtils) WriteStringFile(filename string, content string) error {
	return r.WriteFile(filename, []byte(content))
}

func (r *FileUtils) WriteFile(fielname string, content []byte) error {
	return os.WriteFile(fielname, content, 0o600)
}

func (r *FileUtils) GetFileContent(filepath string) ([]byte, error) {
	return ioutil.ReadFile(filepath)
}

func (r *FileUtils) AppendStrToFile(filePath string, content string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.WriteString(content); err != nil {
		return err
	}

	return nil
}

func GetFileNamesOnDir(dirname string) ([]string, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadDir: %w", err)
	}
	filenames := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			filenames = append(filenames, file.Name())
		}
	}
	return filenames, nil
}

// GetAbsolutePathFromRepoRoot will return Absolute path from
// repository root path. E.g: when call GetAbsolutePathFromRepoRoot("/internal/golibs/io/utils.go"),
// will receive "/home/[user-name]/[workspace]/backend/internal/golibs/io/utils.go"
func GetAbsolutePathFromRepoRoot(relative string) (string, error) {
	if len(relative) != 0 && !strings.HasPrefix(relative, "/") {
		relative = "/" + relative
	}
	if len(execwrapper.RootDirectory()) != 0 {
		return execwrapper.RootDirectory() + relative, nil
	}

	path, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("os.Getwd: %w", err)
	}

	for {
		list, err := GetFileNamesOnDir(path)
		if err != nil {
			return "", err
		}
		// checking in current folder is repo root
		match := 0
		for _, name := range list {
			// TODO: for make sure, need create a trusted file to mark repo root path
			if name == "go.mod" || name == "go.sum" {
				match++
			}
		}
		if match == 2 {
			return path + relative, nil
		}

		path += "/.."
		path, err = filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("filepath.Abs: %w", err)
		}
	}
}

func WriteFileByIOReader(reader io.Reader, dst string) error {
	buf := &bytes.Buffer{}
	_, err := buf.ReadFrom(reader)
	if err != nil {
		return fmt.Errorf("could not read data: %w", err)
	}

	wf, err := NewFileByPath(dst)
	if err != nil {
		return fmt.Errorf("NewFileByPath: %w", err)
	}
	_, err = wf.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("could not create file %s: %w", dst, err)
	}
	if err = wf.Close(); err != nil {
		return fmt.Errorf("could not close file %s: %w", dst, err)
	}
	return nil
}

func NewFileByPath(filePath string) (*os.File, error) {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o777)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// VisitFileLine will visit though each line of file,
// proc will call to process data of each line, if it returns false, exit.
func VisitFileLine(path string, proc func(line []byte) bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if !proc(scanner.Bytes()) {
			break
		}
	}
	if err = scanner.Err(); err != nil {
		return err
	}

	return nil
}

func ReadFileByRepoPath(repoPath string) ([]byte, error) {
	path, err := GetAbsolutePathFromRepoRoot(repoPath)
	if err != nil {
		return nil, fmt.Errorf("could not get absolute path of %s: %v", repoPath, err)
	}
	res, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %v", path, err)
	}

	return res, nil
}

func (r *FileUtils) GetFoldersOnDir(folder string) ([]string, error) {
	files, error := ioutil.ReadDir(folder)
	if error != nil {
		return nil, error
	}
	folderNames := []string{}
	for _, file := range files {
		if file.IsDir() {
			folderNames = append(folderNames, file.Name())
		}
	}
	return folderNames, nil
}

func (r *FileUtils) Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
