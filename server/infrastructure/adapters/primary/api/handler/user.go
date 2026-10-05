package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/validate"
)

type UserHandler struct {
	uc    primary.UserUseCase
	image port.ImageStore
}

func NewUserHandler(uc primary.UserUseCase, img port.ImageStore) *UserHandler {
	return &UserHandler{uc: uc, image: img}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Request([]validate.Rule{
		{Field: "email", Value: req.Email},
		{Field: "password", Value: req.Password},
		{Field: "confirm_password", Value: req.ConfirmPassword},
		{Field: "first_name", Value: req.FirstName},
		{Field: "last_name", Value: req.LastName},
	}); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.uc.Register(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusCreated, result)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Request([]validate.Rule{
		{Field: "email", Value: req.Email},
		{Field: "password", Value: req.Password},
	}); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.uc.Login(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	result, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.uc.UpdateProfile(r.Context(), id, req); err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}

func (h *UserHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 5 << 20

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		respondError(w, http.StatusBadRequest, "photo is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusBadRequest, "failed to read photo")
		return
	}

	path, err := h.image.Upload(r.Context(), data, photoFilename(id, header.Filename))
	if err != nil {
		respondError(w, http.StatusBadRequest, "failed to store photo")
		return
	}

	if err := h.uc.UpdateProfile(r.Context(), id, dto.UpdateUserRequest{PhotoURL: &path}); err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}

func photoFilename(id uuid.UUID, filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		ext = ".jpg"
	}
	return fmt.Sprintf("user-%s-%d%s", id, time.Now().UnixNano(), ext)
}
