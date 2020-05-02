package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

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
}

type emptyBucketsStore struct{}

func (s *emptyBucketsStore) GetObjectMetadata(bucket, key string) (store.ObjectMetadata, error) {
	return nil, &store.ObjectNotFound{Bucket: bucket, Key: key}
}
func (s *emptyBucketsStore) CopyObject(bucket, key string, w io.Writer) (int64, error) {
	return 0, &store.ObjectNotFound{Bucket: bucket, Key: key}
}

func TestBucketNotFound(t *testing.T) {
	r, err := http.NewRequest("GET", "/not/existing/key", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	store := emptyBucketsStore{}
	handler := http.HandlerFunc(ServeFromBuckets(map[string]string{}, &store))
	handler.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
	}
}

func TestObjectNotFound(t *testing.T) {
	r, err := http.NewRequest("GET", "/test-alias/some/obj/key", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	store := emptyBucketsStore{}
	handler := http.HandlerFunc(ServeFromBuckets(map[string]string{
		"test-alias": "test-bucket",
	}, &store))
	handler.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
	}
}

type dummyObject struct {
	contentType string
	body        string
}

func (o dummyObject) ContentType() string {
	return o.contentType
}
func (o dummyObject) Size() int64 {
	bytes := []byte(o.body)
	return int64(len(bytes))
}
func (o dummyObject) Updated() time.Time {
	return time.Now()
}

type dummyObjectStore struct {
	byBucket map[string]map[string]dummyObject
}

func (s *dummyObjectStore) getObject(bucket, key string) (o dummyObject, err error) {
	objectsByKey, ok := s.byBucket[bucket]
	if !ok {
		return o, fmt.Errorf("Bucket not found %s", bucket)
	}
	o, ok = objectsByKey[key]
	if !ok {
		return o, fmt.Errorf("key not found %s", key)
	}
	return o, nil
}

func (s *dummyObjectStore) GetObjectMetadata(bucket, key string) (store.ObjectMetadata, error) {
	return s.getObject(bucket, key)
}

func (s *dummyObjectStore) CopyObject(bucket, key string, w io.Writer) (int64, error) {
	o, err := s.getObject(bucket, key)
	written, err := w.Write([]byte(o.body))
	return int64(written), err
}

func TestObjectFoundInBucketAlias(t *testing.T) {
	r, err := http.NewRequest("GET", "/b1/existing/key", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	dummyContent := "some plain text content"
	objectsByBucket := map[string]map[string]dummyObject{}
	objectsByBucket["bucket1"] = make(map[string]dummyObject)
	objectsByBucket["bucket1"]["existing/key"] = dummyObject{contentType: "text/plain", body: dummyContent}

	objectsByBucket["bucket2"] = make(map[string]dummyObject)

	aliases := map[string]string{
		"b1": "bucket1",
		"x2": "bucket2",
	}

	store := dummyObjectStore{byBucket: objectsByBucket}
	handler := http.HandlerFunc(ServeFromBuckets(aliases, &store))
	handler.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("Unexpected status code %s", resp.Header.Get("Content-Type"))
	}

	var bytes []byte
	bytes, err = ioutil.ReadAll(resp.Body)
	content := string(bytes)
	if content != dummyContent {
		t.Errorf("Unexpected file content for /existing/key: %s", content)
	}

}
