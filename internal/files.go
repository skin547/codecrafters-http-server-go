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
	return &FileStorage{
		rootDir: rootDir,
	}
}

func (f *FileStorage) read(fileName string) ([]byte, error) {
	filePath := fmt.Sprintf("%s/%s", f.rootDir, fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// res.statusCode = 404
		// res.statusMsg = "Not Found"
		// break
		return nil, fmt.Errorf("file not found: %s", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %s", err)
	}
	defer file.Close()

	data := make([]byte, 1024)
	size, err := file.Read(data)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %s", err)
	}
	return data[:size], nil
}

func (f *FileStorage) write(fileName string, data []byte) error {
	filePath := fmt.Sprintf("%s/%s", f.rootDir, fileName)
	return os.WriteFile(filePath, []byte(data), 0644)
}
