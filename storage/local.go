package storage

import (
	"context"
	// "crypto/rand"
	// "crypto/sha256"
	// "encoding/base64"
	// "encoding/hex"
	"fmt"
	"io"
	"net/http"
	// "log"
	"os"
	"path/filepath"
	// "strings"
	"time"

	"github.com/gabrielforster/easy-storage/models"
	"github.com/gabrielforster/easy-storage/utils"
)

type LocalStorage struct {
	dir string
}

func NewLocalStorage(dir string) *LocalStorage {
	return &LocalStorage{dir: dir}
}

func (s *LocalStorage) Upload(ctx context.Context, key string, file io.Reader, filename string) error {
	// Create directory if it doesn't exist
	err := os.MkdirAll(s.dir, 0755)
	if err != nil {
		return err
	}

	// Create file path
	filePath := filepath.Join(s.dir, key)

	// Open file for writing
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write file content
	_, err = io.Copy(f, file)
	if err != nil {
		return err
	}

	return nil
}

func (s *LocalStorage) Download(ctx context.Context, key string) (*models.File, error) {
	// Create file path
	filePath := filepath.Join(s.dir, key)

	// Open file for reading
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Get file info
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	// Read file content
	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// Create file model
	file := &models.File{
		Name:         fi.Name(),
		ContentType:  http.DetectContentType(content),
		Content:      content,
		LastModified: fi.ModTime(),
	}

	return file, nil
}

func (s *LocalStorage) GetSignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	// Create signed URL
	signedURL := fmt.Sprintf("http://localhost:8080/download/%s?expires=%s&signature=%s",
		key,
		time.Now().Add(expires).Unix(),
		utils.Sign(key, os.Getenv("SECRET_KEY")),
	)

	return signedURL, nil
}
