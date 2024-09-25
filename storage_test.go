package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "mylittlefile"
	pathKey := CASPathTransformFunc(key)
	expectedOriginal := "eac8041a14265834aa05c194edb61d01c6e71211"
	expectedPathname := "eac80/41a14/26583/4aa05/c194e/db61d/01c6e/71211"

	if pathKey.PathName != expectedPathname {
		t.Errorf("have %s want %s", pathKey.PathName, expectedPathname)
	}

	if pathKey.Filename != expectedOriginal {
		t.Errorf("have %s want %s", pathKey.Filename, expectedOriginal)
	}
}

func newStorage() *Storage {
	options := StorageOptions{
		PathTransformFunc: CASPathTransformFunc,
	}
	storage := NewStorage(options)

	return storage
}

func teardown(t *testing.T, s *Storage) {
	if err := s.Close(); err != nil {
		t.Error(err)
	}
}

func TestStorage(t *testing.T) {
	storage := newStorage()
  defer teardown(t, storage)

	key := "testkey"
  id := generateID()
	data := []byte("a big file here")
	if _, err := storage.writeStream(id, key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if ok := storage.Has(id, key); !ok {
		t.Errorf("expect (%s) to exist", key)
	}

	_, r, err := storage.Read(id, key)
	if err != nil {
		t.Error(err)
	}

	b, _ := ioutil.ReadAll(r)
	if string(b) != string(data) {
		t.Errorf("want %s have %s", data, b)
	}

	fmt.Println(string(b))

	if err := storage.Delete(id, key); err != nil {
		t.Error(err)
	}

	if ok := storage.Has(id, key); ok {
		t.Errorf("expect (%s) to not exist", key)
	}
}
