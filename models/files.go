package models

import (
	"time"
)

type File struct {
	Name         string
	ContentType  string
	Content      []byte
	LastModified time.Time
	Metadata     map[string]string
}
