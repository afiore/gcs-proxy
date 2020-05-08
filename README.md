# GCS store Proxy

A minimal web app to serve static assets stored in Google Clould storage buckets via HTTP.

## Rationale

I ended up writing this as I couldn't find a viable way to expose GCS objects via HTTP while at the same time
enforcing authentication/authorisation. Hence I have resorted to writing a simple app to proxy HTTP requests to

## Building and running

```bash
# Build the executable
go build -o bin/gcs-proxy main/main.go

# running supplying the configuration file as it's sole argument
./bin/gcs-proxy config.toml

# Build docker image
docker build -t afiore/gcs-proxy:latest .

# Run containerized gcs-proxy making sure you mount a volume with the .toml file e.g.
docker run --rm --volume $(pwd):/tmp afiore/gcs-proxy:latest /tmp/config.toml
```

## Configuration

The program expects a few mandatory configuration parameters to be supplied a `.toml` file.
This is the expected file structure:

```
[Gcs]
# Full path to the GCP Service account credentials
ServiceAccountFilePath = "/etc/gcs-proxy-sa.json"

# A mapping from request path fragment to bucket name
Buckets = { "bucket1" = "loadballancer-test-bucket" }

[Web]
# The port the app will be listening to
Port = 9999
```
