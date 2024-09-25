package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const DefaultRootFolderName = "root_folder"

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashString := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashString) / blockSize
	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashString[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		Filename: hashString,
	}
}

type PathTransformFunc func(string) PathKey

var DumbPathTransformFunc PathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		Filename: key,
	}
}

type PathKey struct {
	PathName string
	Filename string
}

func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}

func (p PathKey) FirstPathSegment() string {
	firstSegment := strings.Split(p.FullPath(), "/")[0]

	// if len(firstSegment) == 0, something is wrong on the system
	if len(firstSegment) == 0 {
		return ""
	}

	return firstSegment
}

type StorageOptions struct {
	Root              string
	PathTransformFunc PathTransformFunc
}

type Storage struct {
	Options StorageOptions
}

func NewStorage(options StorageOptions) *Storage {
	if options.PathTransformFunc == nil {
		options.PathTransformFunc = DumbPathTransformFunc
	}

	if len(options.Root) == 0 {
		options.Root = DefaultRootFolderName
	}

	return &Storage{
		Options: options,
	}
}

func (s *Storage) Write(id string, key string, r io.Reader) (int64, error) {
	return s.writeStream(id, key, r)
}

func (s *Storage) WriteDecrypt(encKey []byte, id string, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(id, key)
	if err != nil {
		return 0, err
	}
	n, err := copyDecrypt(encKey, r, f)
	return int64(n), err
}

func (s *Storage) openFileForWriting(id string, key string) (*os.File, error) {
	pathKey := s.Options.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Options.Root, id, pathKey.PathName)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return nil, err
	}

	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Options.Root, id, pathKey.FullPath())

	return os.Create(fullPathWithRoot)
}

func (s *Storage) writeStream(id string, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(id, key)
	if err != nil {
		return 0, err
	}
	return io.Copy(f, r)
}

func (s *Storage) Has(id, key string) bool {
	pathKey := s.Options.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Options.Root, id, pathKey.FullPath())

	_, err := os.Stat(fullPathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Storage) Delete(id string, key string) error {
	pathKey := s.Options.PathTransformFunc(key)

	defer func() {
		log.Printf("deleted [%s] from disk", pathKey.Filename)
	}()

	firstPathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Options.Root, id, pathKey.FirstPathSegment())

	return os.RemoveAll(firstPathNameWithRoot)
}

func (s *Storage) Read(id string, key string) (int64, io.Reader, error) {
	return s.readStream(id, key)
}

func (s *Storage) readStream(id string, key string) (int64, io.ReadCloser, error) {
	pathKey := s.Options.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Options.Root, id, pathKey.FullPath())

	file, err := os.Open(fullPathWithRoot)
	if err != nil {
		return 0, nil, err
	}

	fi, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}

	return fi.Size(), file, nil
}

func (s *Storage) Close() error {
	return os.RemoveAll(s.Options.Root)
}
