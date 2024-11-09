package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/ssimpl/simple-storage/internal/api/model"
)

type objectManager interface {
	StoreObject(ctx context.Context, objectName string, src io.ReaderAt, size int64) error
	RetrieveObject(ctx context.Context, objectName string, dst io.Writer) error
}

type Handler struct {
	objManager    objectManager
	fileSizeLimit int64
}

func NewHandler(objManager objectManager, fileSizeLimit int64) *Handler {
	return &Handler{objManager: objManager, fileSizeLimit: fileSizeLimit}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		h.uploadFile(w, r)
	case http.MethodGet:
		h.downloadFile(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) uploadFile(w http.ResponseWriter, r *http.Request) {
	fileName := strings.Trim(r.URL.Path, "/")
	if fileName == "" {
		http.Error(w, "File name is required", http.StatusBadRequest)
		return
	}

	filePath := fmt.Sprintf("/tmp/%s", uuid.New().String())

	tempFile, err := os.Create(filePath)
	if err != nil {
		respondWithInternalError(w, "Failed to create temp file", err)
		return
	}
	defer func() {
		if err := tempFile.Close(); err != nil {
			slog.Error("Failed to close temp file", "err", err)
		}
		if err := os.Remove(filePath); err != nil {
			slog.Error("Failed to remove temp file", "err", err)
		}
	}()

	written, err := io.Copy(tempFile, io.LimitReader(r.Body, h.fileSizeLimit))
	if err != nil {
		respondWithInternalError(w, "Failed to copy file data", err)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			slog.Error("Failed to close request body", "err", err)
		}
	}()

	if written == h.fileSizeLimit {
		buffer := make([]byte, 1)
		extraRead, err := r.Body.Read(buffer)
		if extraRead > 0 || (err != nil && err != io.EOF) {
			http.Error(w, "File size exceeds the limit", http.StatusRequestEntityTooLarge)
			return
		}
	}

	fileInfo, err := tempFile.Stat()
	if err != nil {
		respondWithInternalError(w, "Failed to get file info", err)
		return
	}

	if err := h.objManager.StoreObject(r.Context(), fileName, tempFile, fileInfo.Size()); err != nil {
		respondWithInternalError(w, "Failed to store object", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) downloadFile(w http.ResponseWriter, r *http.Request) {
	fileName := strings.Trim(r.URL.Path, "/")
	if fileName == "" {
		http.Error(w, "File name is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")

	if err := h.objManager.RetrieveObject(r.Context(), fileName, w); err != nil {
		if errors.Is(err, model.ErrObjectNotFound) {
			http.Error(w, model.ErrObjectNotFound.Error(), http.StatusBadRequest)
			return
		}

		respondWithInternalError(w, "Failed to retrieve object", err)
		return
	}
}

func respondWithInternalError(w http.ResponseWriter, message string, err error) {
	slog.Error(message, "err", err)
	http.Error(w, message, http.StatusInternalServerError)
}
