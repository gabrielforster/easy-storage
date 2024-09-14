package storage

import (
	"bufio"
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

	"strings"
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

func (s *LocalStorage) Upload(ctx context.Context, key string, file io.Reader, filename string, metadata map[string]string) error {
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

	// Write metadata header
	var metadataHeader string
	metadataHeader += "Metadata: INIT\n"
	for k, v := range metadata {
		metadataHeader += fmt.Sprintf("%s: %s\n", k, v)
	}
	metadataHeader += "Metadata: END\n"

	_, err = f.WriteString(metadataHeader)
	if err != nil {
		return err
	}

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

	metadata, err := readMetadata(f)
	if err != nil {
		fmt.Printf("error on read metadata: %+v \n", err)
		return nil, err
	}

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
		Metadata:     metadata,
	}

	return file, nil
}

func (s *LocalStorage) GetSignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	// Create signed URL
	signedURL := fmt.Sprintf("htp://localhost:8080/download/%s?expires=%d&signature=%s",
		key,
		time.Now().Add(expires).Unix(),
		utils.Sign(key, os.Getenv("SECRET_KEY")),
	)

	return signedURL, nil
}

func readMetadata(f *os.File) (map[string]string, error) {
	var metadataHeader string
	var inMetadata bool

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "Metadata: INIT" {
			inMetadata = true
			continue
		}

		if line == "Metadata: END" {
			inMetadata = false
			break
		}

		if inMetadata {
			metadataHeader += line + "\n"
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("error on scanner: %+v\n", err)
		return nil, err
	}

	metadata := make(map[string]string)
	lines := strings.Split(metadataHeader, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		metadata[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	return metadata, nil
}
