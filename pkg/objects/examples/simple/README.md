# Simple Object Storage Example

This example demonstrates basic usage of the `pkg/objects` storage system with the disk provider.

## Features Demonstrated

- Creating a disk storage provider
- Uploading files
- Downloading files
- Generating presigned URLs
- Checking file existence
- Deleting files

## Running the Example

Use the provided Taskfile to build and execute the binary with the required
`examples` build tag:

```bash
cd pkg/objects/examples/simple
task run
```

Alternatively, run it manually:

```bash
go run -tags examples .
```

## Expected Output

```
=== Simple Object Storage Example ===

1. Uploading file...
   ✓ Uploaded: hello.txt (size: 37 bytes)

2. Downloading file...
   ✓ Downloaded: hello.txt (37 bytes)
   Content: Hello, World! This is a test file.

3. Getting presigned URL...
   ✓ URL: http://localhost:8080/files/hello.txt

4. Checking file existence...
   ✓ File exists: true

5. Deleting file...
   ✓ File deleted successfully
   ✓ File exists after deletion: false

=== Example completed successfully ===
```

## Clean Up

The example automatically cleans up the temporary storage directory on exit.
