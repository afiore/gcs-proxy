package gcs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

//Object represents a GCP storage record
type Object struct {
	Key         string
	ContentType string
	Size        int64
	Updated     time.Time
	reader      *storage.Reader
}

//Copy copies the object to a writer
func (obj *Object) Copy(w io.Writer) (int64, error) {
	defer obj.reader.Close()
	return io.Copy(w, obj.reader)
}

//ObjectNotFound is the error value returned by GetObject when the supplied key is not found
type ObjectNotFound struct {
	Bucket string
	Key    string
}

func (e *ObjectNotFound) Error() string { return e.Key + " not found in bucket " + e.Bucket }

func toGCSObject(ctx context.Context, attr *storage.ObjectAttrs, obj *storage.ObjectHandle) Object {
	reader, err := obj.NewReader(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return Object{Key: attr.Name, ContentType: attr.ContentType, Size: attr.Size, Updated: attr.Updated, reader: reader}
}

//GetObject fetches an object from a bucket
func GetObject(jsonPath, bucketName, objectKey string) (Object, error) {
	var gcsObj Object
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(jsonPath))
	if err != nil {
		log.Fatal(err)
	}
	bucket := client.Bucket(bucketName)
	obj := bucket.Object(objectKey)
	attrs, err := obj.Attrs(ctx)

	if errors.Is(err, storage.ErrObjectNotExist) {
		return gcsObj, &ObjectNotFound{Bucket: bucketName, Key: objectKey}

	}
	if err != nil {
		return gcsObj, fmt.Errorf("An error occurred while accessing object %s", err.Error())
	}
	gcsObj = toGCSObject(ctx, attrs, obj)

	return gcsObj, nil
}
