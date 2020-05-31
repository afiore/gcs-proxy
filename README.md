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

# supply your Google service account file and deploy the app through the provided Helm chart
gcp_sa=$(cat /path/to/my/gcp_sa.json|base64 -w 0)
helm install gcs-proxy charts/gcs-proxy/ --set gcp_sa_base64=$gcp_sa --set gcs_proxy.oauth.client_id=$CLIENT_ID  --set gcs_proxy.oauth.client_secret=$CLIENT_SECRET --set gcs_proxy.oauth.session_secret=$SESSION_SECRET 
```

## Configuration

The program expects a few mandatory configuration parameters to be supplied a `.toml` file.
Please refer to `config.toml` for a sample of the available configurable parameters.
