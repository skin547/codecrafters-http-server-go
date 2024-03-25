package internal

import (
	"fmt"
	"os"
)

type Storage interface {
	read(fileName string) ([]byte, error)
	write(fileName string, data []byte) error
}

type FileStorage struct {
	rootDir string
}

func NewFileStorage(rootDir string) *FileStorage {
	// if directory not exist, create one
	_, err := os.Stat(rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(rootDir, 0755)
			if err != nil {
				panic(fmt.Sprintf("Error creating directory: %s", err.Error()))
			}
		} else {
			panic(fmt.Sprintf("Error checking directory: %s", err.Error()))
		}
	}
	return &FileStorage{
		rootDir: rootDir,
	}
}

func (f *FileStorage) read(fileName string) ([]byte, error) {
	filePath := fmt.Sprintf("%s/%s", f.rootDir, fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, &NotFoundError{FileName: fileName}
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, &InternalServerError{Reason: err.Error()}
	}
	defer file.Close()

	data := make([]byte, 1024)
	size, err := file.Read(data)
	if err != nil {
		return nil, &InternalServerError{Reason: err.Error()}
	}
	return data[:size], nil
}

func (f *FileStorage) write(fileName string, data []byte) error {
	filePath := fmt.Sprintf("%s/%s", f.rootDir, fileName)
	return os.WriteFile(filePath, []byte(data), 0644)
}
