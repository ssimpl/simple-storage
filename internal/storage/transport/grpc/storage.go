package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/ssimpl/simple-storage/internal/storage/model"
	"github.com/ssimpl/simple-storage/pkg/storage"

	"google.golang.org/grpc"
)

const statusOK = "OK"

const bufferSize = 64000

type objectStorage interface {
	StoreObject(ctx context.Context, objectName string, src io.Reader) error
	RetrieveObject(ctx context.Context, objectName string, dst io.Writer) error
}

type StorageServer struct {
	storage.UnimplementedStorageServer

	storage objectStorage
}

func NewStorageServer(storage objectStorage) *StorageServer {
	return &StorageServer{
		storage: storage,
	}
}

func (s *StorageServer) Upload(stream grpc.ClientStreamingServer[storage.UploadRequest, storage.UploadResponse]) error {
	req, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive initial upload request: %w", err)
	}

	objectName := req.GetObjectId()
	if objectName == "" {
		return model.ErrObjectNameRequired
	}

	pipeReader, pipeWriter := io.Pipe()
	go func() {
		defer pipeWriter.Close()
		for {
			req, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				slog.Error("failed to receive upload chunk", "err", err)
				return
			}
			if _, err := pipeWriter.Write(req.GetData()); err != nil {
				slog.Error("failed to write chunk data to pipe", "err", err)
				return
			}
		}
	}()

	if err := s.storage.StoreObject(stream.Context(), objectName, pipeReader); err != nil {
		return fmt.Errorf("failed to store object: %w", err)
	}

	return stream.SendAndClose(&storage.UploadResponse{
		Status: statusOK,
	})
}

func (s *StorageServer) Download(
	req *storage.DownloadRequest, stream grpc.ServerStreamingServer[storage.DownloadResponse],
) error {
	objectName := req.GetObjectId()
	if objectName == "" {
		return model.ErrObjectNameRequired
	}

	pipeReader, pipeWriter := io.Pipe()
	go func() {
		defer pipeWriter.Close()
		if err := s.storage.RetrieveObject(stream.Context(), objectName, pipeWriter); err != nil {
			slog.Error("failed to retrieve object", "err", err)
			pipeWriter.CloseWithError(err)
		}
	}()

	buffer := make([]byte, bufferSize)
	for {
		n, err := pipeReader.Read(buffer)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read chunk data from pipe: %w", err)
		}

		err = stream.Send(&storage.DownloadResponse{
			Data: buffer[:n],
		})
		if err != nil {
			return fmt.Errorf("failed to send chunk data: %w", err)
		}
	}

	return nil
}
