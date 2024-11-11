package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ssimpl/simple-storage/pkg/storage"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const responseStatusOK = "OK"

const bufferSize = 64000

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Store(ctx context.Context, serverAddr string, objectID uuid.UUID, data io.Reader) error {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to create grpc client: %w", err)
	}

	stream, err := storage.NewStorageClient(conn).Upload(ctx)
	if err != nil {
		return fmt.Errorf("failed to open upload stream: %w", err)
	}

	if err := stream.SendMsg(&storage.UploadRequest{ObjectId: objectID.String()}); err != nil {
		return fmt.Errorf("failed to send upload request: %w", err)
	}

	buffer := make([]byte, bufferSize)
	for {
		n, readErr := data.Read(buffer)
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return fmt.Errorf("failed to read data: %w", readErr)
		}

		if n > 0 {
			if err := stream.Send(&storage.UploadRequest{Data: buffer[:n]}); err != nil {
				return fmt.Errorf("failed to send upload data chunk: %w", err)
			}
		}

		if readErr == io.EOF {
			break
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to close upload stream: %w", err)
	}

	if res.Status != responseStatusOK {
		return fmt.Errorf("upload failed: %s", res.Status)
	}

	return nil
}

func (c *Client) Retrieve(ctx context.Context, serverAddr string, objectID uuid.UUID, dst io.Writer) error {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to create grpc client: %w", err)
	}

	stream, err := storage.NewStorageClient(conn).Download(ctx, &storage.DownloadRequest{ObjectId: objectID.String()})
	if err != nil {
		return fmt.Errorf("failed to open download stream: %w", err)
	}

	for {
		res, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("failed to receive download data chunk: %w", err)
		}
		if _, err := dst.Write(res.Data); err != nil {
			return fmt.Errorf("failed to write data chunk: %w", err)
		}
	}

	return nil
}
