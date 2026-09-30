package dto

import "github.com/google/uuid"

type CategoryDTO struct {
	ID                     uuid.UUID  `json:"id"`
	Name                   string     `json:"name"`
	Description            string     `json:"description"`
	MainCategory           string     `json:"main_category"`
	IsActive               bool       `json:"is_active"`
	DefaultUnitOfMeasureID *uuid.UUID `json:"default_unit_of_measure_id,omitempty"`
}

type CreateCategoryRequest struct {
	Name                   string     `json:"name"`
	Description            string     `json:"description"`
	MainCategory           string     `json:"main_category"`
	DefaultUnitOfMeasureID *uuid.UUID `json:"default_unit_of_measure_id"`
}

type UpdateCategoryRequest struct {
	Name                   *string    `json:"name"`
	Description            *string    `json:"description"`
	MainCategory           *string    `json:"main_category"`
	DefaultUnitOfMeasureID *uuid.UUID `json:"default_unit_of_measure_id"`
}

type CategoryStatusRequest struct {
	IsActive bool `json:"is_active"`
}
