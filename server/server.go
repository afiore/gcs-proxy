package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/afiore/gcs-proxy/store"
)

//ServeFromBuckets maps incoming requests to bucket objects defined in the supplied configuration
func ServeFromBuckets(bucketByAlias map[string]string, objStore store.ObjectStoreOps) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		for alias, bucketName := range bucketByAlias {
			if !strings.HasPrefix(r.URL.Path, "/"+alias) {
				continue
			}
			objectKey := strings.Replace(r.URL.Path, fmt.Sprintf("/%s/", alias), "", 1)

			log.Printf("Fetching key: %s from bucket %s", objectKey, bucketName)

			meta, err := objStore.GetObjectMetadata(bucketName, objectKey)

			var objNotFoundErr *store.ObjectNotFound
			if errors.As(err, &objNotFoundErr) {
				log.Printf("key %s not found in bucket %s", objectKey, bucketName)
				http.NotFound(w, r)
				return
			}
			if err != nil {
				http.Error(w, "An internal error has occured", 500)
				return
			}

			for k, v := range objectHeaders(meta) {
				w.Header().Add(k, v)
			}

			_, err = objStore.CopyObject(bucketName, objectKey, w)
			if err != nil {
				log.Fatal(err)
			} else {
				return
			}
		}
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}

}

func base(key string) string {
	parts := strings.Split(key, "/")
	return parts[len(parts)-1]
}

func replaceEmptyBase(key string, replacement string) string {
	if key != "" && base(key) == "" {
		parts := strings.Split(key, "/")
		return strings.Join(append(parts[:len(parts)-1], replacement), "/")

	}
	return key
}

func objectHeaders(o store.ObjectMetadata) map[string]string {
	return map[string]string{
		"Content-Type":   o.ContentType(),
		"Content-Length": fmt.Sprintf("%d", o.Size()),
		"Last-Modified":  o.Updated().Format(http.TimeFormat),
	}
}
