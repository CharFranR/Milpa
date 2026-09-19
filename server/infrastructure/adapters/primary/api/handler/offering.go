package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/validate"

	"strconv"
)

type OfferingHandler struct {
	uc    primary.OfferingUseCase
	image port.ImageStore
}

func NewOfferingHandler(uc primary.OfferingUseCase, img port.ImageStore) *OfferingHandler {
	return &OfferingHandler{uc: uc, image: img}
}

func (h *OfferingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOfferingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Request([]validate.Rule{
		{Field: "user_id", Value: req.UserID},
		{Field: "name", Value: req.Name},
	}); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.uc.CreateOffering(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusCreated, result)
}

func (h *OfferingHandler) Create_v2(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 10 << 20

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := uuid.Parse(r.FormValue("user_id"))

	if err != nil {
		respondError(w, http.StatusBadRequest, "user_id not a valid number")
		return
	}

	idType, err := strconv.Atoi(r.FormValue("type"))

	if err != nil {
		respondError(w, http.StatusBadRequest, "type is not valid")
		return
	}

	OfferingType := domain.OfferingType(idType)

	price, err := strconv.ParseFloat(r.FormValue("price"), 64)

	if err != nil {
		respondError(w, http.StatusBadRequest, "ain't a correct price number")
		return
	}

	req := dto.CreateOfferingRequest{
		UserID:      userID,
		Type:        OfferingType,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Price:       price,
	}

	if file, header, err := r.FormFile("image_url"); err == nil {
		defer file.Close()

		data, err := io.ReadAll(file)

		if err != nil {
			respondError(w, http.StatusBadRequest, "na, ur image sucks")
			return
		}

		imagePath, err := h.image.Upload(r.Context(), data, header.Filename)
		if err != nil {
			respondError(w, http.StatusBadRequest, "failed to upload image")
			return
		}

		req.ImageURL = imagePath

	}

	result, err := h.uc.CreateOffering(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusCreated, result)

}

func (h *OfferingHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid offering id")
		return
	}

	result, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *OfferingHandler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.URL.Query().Get("user_id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	result, err := h.uc.GetByUserID(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

func (h *OfferingHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid offering id")
		return
	}

	var req dto.UpdateOfferingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.uc.UpdateOffering(r.Context(), id, req); err != nil {
		handleError(w, err)
		return
	}

	respond(w, http.StatusOK, nil)
}

func (h *OfferingHandler) DeleteOffering(w http.ResponseWriter, r *http.Request) {

	id, err := uuid.Parse(chi.URLParam(r, "id"))

	if err != nil {
		respondError(w, http.StatusBadRequest, "Not a valid id")
	}

	if err := h.uc.DeleteOffering(r.Context(), id); err != nil {
		handleError(w, err)
	}

	respond(w, http.StatusOK, nil)

}
