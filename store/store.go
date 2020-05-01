package store

import (
	"io"
	"time"
)

//ObjectMetadata exposes key metadata for an object
type ObjectMetadata interface {
	Key() string
	ContentType() string
	Size() int64
	Updated() time.Time
}

//ObjectStoreOps exposes basic operations on objects
type ObjectStoreOps interface {
	GetMetadata(bucket, key string) (ObjectMetadata, error)
	CopyObject(bucket, key string, w io.Writer) (int64, error)
}

//ObjectNotFound is the error value returned by GetObject when the supplied key is not found
type ObjectNotFound struct {
	Bucket string
	Key    string
}

func (e *ObjectNotFound) Error() string { return e.Key + " not found in bucket " + e.Bucket }
