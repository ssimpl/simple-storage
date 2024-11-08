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
	objManager objectManager
}

func NewHandler(objManager objectManager) *Handler {
	return &Handler{objManager: objManager}
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
		errMsg := "Failed to create temp file"
		slog.Error(errMsg, "err", err)
		http.Error(w, errMsg, http.StatusInternalServerError)
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

	if _, err := io.Copy(tempFile, r.Body); err != nil {
		errMsg := "Failed to copy file data"
		slog.Error(errMsg, "err", err)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			slog.Error("Failed to close request body", "err", err)
		}
	}()

	fileInfo, err := tempFile.Stat()
	if err != nil {
		errMsg := "Failed to get file info"
		slog.Error(errMsg, "err", err)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	if err := h.objManager.StoreObject(r.Context(), fileName, tempFile, fileInfo.Size()); err != nil {
		errMsg := "Failed to store object"
		slog.Error(errMsg, "err", err)
		http.Error(w, errMsg, http.StatusInternalServerError)
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

		errMsg := "Failed to retrieve object"
		slog.Error(errMsg, "err", err)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
}
