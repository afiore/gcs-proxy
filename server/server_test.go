package server

import (
	"io"
	"testing"

	"github.com/afiore/gcs-proxy/store"
)

func TestReplaceEmptyBase(t *testing.T) {
	got := replaceEmptyBase("", "index.html")
	if got != "" {
		t.Errorf("replaceEmptyBase('') = %s; want ''", got)
	}

	got = replaceEmptyBase("/foo/bar/baz", "index.html")
	if got != "/foo/bar/baz" {
		t.Errorf("replaceEmptyBase('/foo/bar/baz') = %s; want /foo/bar/baz", got)
	}

	got = replaceEmptyBase("/foo/bar", "index.html")
	if got != "/foo/bar" {
		t.Errorf("replaceEmptyBase('/foo/bar') = %s; want /foo/index.html", got)
	}
}

type emptyBucketsStore struct{}

func (s *emptyBucketsStore) GetObjectMetadata(bucket, key string) (store.ObjectMetadata, error) {
	return nil, &store.ObjectNotFound{Bucket: bucket, Key: key}
}
func (s *emptyBucketsStore) CopyObject(bucket, key string, w io.Writer) (int64, error) {
	return 0, &store.ObjectNotFound{Bucket: bucket, Key: key}
}

func TestObjectNotFound(t *testing.T) {

}
