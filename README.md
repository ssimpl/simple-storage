# Simple Storage Service

Distributed file storage service. Files are uploaded to a primary server, split into several fragments, and distributed across multiple storage servers. It dynamically balances storage and supports adding new storage servers.

## FAST start and check

```bash
make up
make client
```

## How to start/stop

### Start

Start the Storage service in docker:

```bash
make up
```

### Stop

Stop the Storage service:

```bash
make down
```

## Testing client

### Run client

This command will **upload** a file (`./testdata/funny_cats.mp4`) to the storage service, **download** it back, **compare** the files and write **logs**.

```bash
make client
```

To run the client with a different file:

```bash
go run ./cmd/client/... --file=./README.md
```

## API Endpoints

### Upload

Uploads a file to the primary server.

```
PUT /<file_name>

Body: file data
```

#### CURL example

```bash
curl -X PUT --data-binary @./testdata/funny_cats.mp4 http://localhost:8080/file_from_curl.mp4
```

### Download

Downloads a file from the primary server.

```
GET /<file_name>
```

#### CURL example

```bash
curl --output ./funny_cats.mp4 http://localhost:8080/file_from_curl.mp4
```

## Useful commands

### Run tests

```bash
make test
```

### Run linter

```bash
make lint
```

### Generate code

```bash
make generate
```
