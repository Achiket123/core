# Simple S3 (MinIO) Example

This example shows how to use `pkg/objects/storage` with an S3-compatible
endpoint (such as MinIO). It performs an upload, download, and cleanup cycle
while reusing the existing `ObjectService` API.

## Requirements

- MinIO or any S3-compatible endpoint
- Go 1.21+

You can spin up MinIO using Docker:

```bash
docker run -it --rm \
  -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  quay.io/minio/minio server /data --console-address :9001
```

## Run the example

```bash
cd pkg/objects/examples/simple-s3
task run
```

Pass any flags via environment variables, for example:

```bash
MINIO_ENDPOINT=http://127.0.0.1:9000 \
MINIO_ACCESS_KEY=minioadmin \
MINIO_SECRET_KEY=minioadmin \
task run
```

To run manually:

```bash
go run -tags examples . \
  -endpoint http://127.0.0.1:9000 \
  -access-key minioadmin \
  -secret-key minioadmin \
  -bucket core-simple-s3
```

You can also specify values via environment variables:

- `MINIO_ENDPOINT`
- `MINIO_ACCESS_KEY`
- `MINIO_SECRET_KEY`
- `MINIO_REGION`
- `MINIO_BUCKET`

The example uploads `assets/sample-data.txt`, downloads it to
`output/downloaded-sample.txt`, and then removes the remote object.
