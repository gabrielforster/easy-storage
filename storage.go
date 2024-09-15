package main

import (
	"bytes"
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

func (p PathKey) FirstPathSegment () string {
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

func (s *Storage) writeStream(key string, r io.Reader) error {
	pathKey := s.Options.PathTransformFunc(key)
  pathWithRoot := fmt.Sprintf("%s/%s", s.Options.Root, pathKey.PathName)

	if err := os.MkdirAll(pathWithRoot, os.ModePerm); err != nil {
		return err
	}

	pathAndFilename := pathKey.FullPath()
  pathAndFilenameWithRoot := fmt.Sprintf(
    "%s/%s",
    s.Options.Root,
    pathAndFilename,
  )

	f, err := os.Create(pathAndFilenameWithRoot)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("written (%d) bytes to disk: %s\n", n, pathAndFilename)

	return nil
}

func (s *Storage) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)

	return buf, err
}

func (s *Storage) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.Options.PathTransformFunc(key)
  pathWithRoot := fmt.Sprintf("%s/%s", s.Options.Root, pathKey.FullPath())
	return os.Open(pathWithRoot)
}

// todo improve this to be more assertive
func (s *Storage) Has(key string) bool {
	pathKey := s.Options.PathTransformFunc(key)
  pathWithRoot := fmt.Sprintf("%s/%s", s.Options.Root, pathKey.FullPath())

	_, err := os.Stat(pathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Storage) Delete(key string) error {
	pathKey := s.Options.PathTransformFunc(key)

	defer func() {
		log.Printf("deleted (%s) from disk)", pathKey.Filename)
	}()

  // todo change this
  // (this will delete more than 1 file if the first segment
  // of their paths are the same...)
  pathToRemove := pathKey.FirstPathSegment()

  pathWithRoot := fmt.Sprintf("%s/%s", s.Options.Root, pathToRemove)
	if err := os.RemoveAll(pathWithRoot); err != nil {
		return err
	}

	return nil
}
