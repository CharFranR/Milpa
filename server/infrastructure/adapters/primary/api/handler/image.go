package handler

import (
	port "milpa/domain/port/secondary"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ImageHandler struct {
	imgRepo port.ImageStore
}

func NewImageHandler(imgRepo port.ImageStore) *ImageHandler {
	return &ImageHandler{imgRepo}
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
