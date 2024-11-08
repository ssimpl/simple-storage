package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ssimpl/simple-storage/internal/storage/model"
)

type ObjectStorage struct {
	storagePath string
}

func NewObjectStorage(storagePath string) *ObjectStorage {
	return &ObjectStorage{
		storagePath: storagePath,
	}
}

func (s *ObjectStorage) StoreObject(_ context.Context, objectName string, src io.Reader) error {
	filePath := s.getFilePath(objectName)

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directories for %s: %w", filePath, err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	_, err = io.Copy(file, src)
	if err != nil {
		return fmt.Errorf("failed to write data to file %s: %w", filePath, err)
	}

	return nil
}

func (s *ObjectStorage) RetrieveObject(_ context.Context, objectName string, dst io.Writer) error {
	filePath := s.getFilePath(objectName)

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file %s not found: %w: %w", filePath, err, model.ErrObjectNotFound)
		}
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return fmt.Errorf("failed to read data from file %s: %w", filePath, err)
	}

	return nil
}

func (s *ObjectStorage) getFilePath(objectName string) string {
	hash := sha256.New()
	hash.Write([]byte(objectName))
	hashString := hex.EncodeToString(hash.Sum(nil))

	subDir1 := hashString[:2]
	subDir2 := hashString[2:4]

	return filepath.Join(s.storagePath, subDir1, subDir2, objectName)
}
