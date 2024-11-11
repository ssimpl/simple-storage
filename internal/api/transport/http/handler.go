package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/ssimpl/simple-storage/internal/api/model"
)

type objectManager interface {
	StoreObject(ctx context.Context, objectName string, src io.Reader, size int64) error
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

	size, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
	if err != nil || size == 0 {
		http.Error(w, "Content-Length header is required", http.StatusBadRequest)
		return
	}

	slog.Info("Received file", "name", fileName, "size", size)

	fileData := http.MaxBytesReader(w, r.Body, h.fileSizeLimit)
	defer func() {
		if err := r.Body.Close(); err != nil {
			slog.Error("Failed to close request body", "err", err)
		}
	}()

	if err := h.objManager.StoreObject(r.Context(), fileName, fileData, size); err != nil {
		var errMaxBytes *http.MaxBytesError
		if errors.As(err, &errMaxBytes) {
			http.Error(w, "File size exceeds the limit", http.StatusRequestEntityTooLarge)
			return
		}
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
