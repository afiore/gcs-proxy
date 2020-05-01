package gcs

import (
	"context"
	"errors"
	"io"
	"log"
	"time"

	"cloud.google.com/go/storage"
	"github.com/afiore/gcs-proxy/store"
	"google.golang.org/api/option"
)

type gcpStore struct {
	saFilePath string
	ctx        context.Context
}

//StoreOps implements basic storage operations on GCP storage
func StoreOps(saFilePath string) store.ObjectStoreOps {
	ctx := context.Background()
	return &gcpStore{saFilePath: saFilePath, ctx: ctx}
}

//Object represents a GCP storage record
type object struct {
	key    string
	attrs  *storage.ObjectAttrs
	reader *storage.Reader
}

func (o object) Key() string {
	return o.key
}
func (o object) ContentType() string {
	return o.attrs.ContentType
}
func (o object) Size() int64 {
	return o.attrs.Size
}
func (o object) Updated() time.Time {
	return o.attrs.Updated
}

func (s *gcpStore) getObject(bucketName, objectKey string) (o *object, err error) {
	var attrs *storage.ObjectAttrs
	client, err := storage.NewClient(s.ctx, option.WithCredentialsFile(s.saFilePath))
	if err != nil {
		log.Fatal(err)
	}
	bucket := client.Bucket(bucketName)
	obj := bucket.Object(objectKey)
	attrs, err = obj.Attrs(s.ctx)

	if errors.Is(err, storage.ErrObjectNotExist) {
		return o, &store.ObjectNotFound{Bucket: bucketName, Key: objectKey}
	}
	if err != nil {
		return o, err
	}
	r, err := obj.NewReader(s.ctx)
	if err != nil {
		return o, err
	}
	return &object{key: objectKey, attrs: attrs, reader: r}, nil
}

func (s *gcpStore) GetMetadata(bucketName, objectKey string) (store.ObjectMetadata, error) {
	o, err := s.getObject(bucketName, objectKey)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (s *gcpStore) CopyObject(bucketName, objectKey string, w io.Writer) (written int64, err error) {
	o, err := s.getObject(bucketName, objectKey)
	defer o.reader.Close()
	if err != nil {
		return 0, err
	}
	return io.Copy(w, o.reader)
}
