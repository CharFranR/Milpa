package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	dto "milpa/aplication/dto"
	port "milpa/domain/port/secondary"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ImageHandler struct {
	imgRepo port.ImageStore
}

func NewImageHandler(imgRepo port.ImageStore) *ImageHandler {
	return &ImageHandler{imgRepo}
}

func (h *ImageHandler) Upload(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 10 << 20

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		respondError(w, http.StatusBadRequest, "image is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusBadRequest, "failed to read image")
		return
	}

	name := uploadedImageFilename(header.Filename)
	path, err := h.imgRepo.Upload(r.Context(), data, name)
	if err != nil {
		log.Printf("image store upload failed for %q: %v", name, err)
		respondError(w, http.StatusBadRequest, "failed to store image")
		return
	}

	respond(w, http.StatusCreated, dto.UploadedImageDTO{Path: path})
}

func uploadedImageFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		ext = ".jpg"
	}
	return fmt.Sprintf("img-%s%s", uuid.NewString(), ext)
}

func (h *ImageHandler) Get(w http.ResponseWriter, r *http.Request) {

	filename := chi.URLParam(r, "filename")

	imageData, err := h.imgRepo.Load(r.Context(), filename)

	if err != nil {
		respondError(w, http.StatusNotFound, "could not find image")
		return
	}

	w.Header().Set("Content-Type", imageData.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(imageData.Content)

}
