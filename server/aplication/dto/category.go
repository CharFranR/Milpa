package dto

import "github.com/google/uuid"

type CategoryDTO struct {
	ID                     uuid.UUID  `json:"id"`
	Name                   string     `json:"name"`
	Description            string     `json:"description"`
	MainCategory           string     `json:"main_category"`
	IsActive               bool       `json:"is_active"`
	DefaultUnitOfMeasureID *uuid.UUID `json:"default_unit_of_measure_id,omitempty"`
	DefaultExpiryDays      *int       `json:"default_expiry_days,omitempty"`
}

type CreateCategoryRequest struct {
	Name                   string     `json:"name"`
	Description            string     `json:"description"`
	MainCategory           string     `json:"main_category"`
	DefaultUnitOfMeasureID *uuid.UUID `json:"default_unit_of_measure_id"`
	DefaultExpiryDays      *int       `json:"default_expiry_days"`
}

type UpdateCategoryRequest struct {
	Name                   *string    `json:"name"`
	Description            *string    `json:"description"`
	MainCategory           *string    `json:"main_category"`
	DefaultUnitOfMeasureID *uuid.UUID `json:"default_unit_of_measure_id"`
	DefaultExpiryDays      *int       `json:"default_expiry_days"`
}

type CategoryStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type UnitOfMeasureDTO struct {
	ID       uuid.UUID `json:"id"`
	Code     string    `json:"code"`
	Name     string    `json:"name"`
	IsActive bool      `json:"is_active"`
}

type CreateUnitOfMeasureRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type UpdateUnitOfMeasureRequest struct {
	Code *string `json:"code"`
	Name *string `json:"name"`
}

type UnitOfMeasureStatusRequest struct {
	IsActive bool `json:"is_active"`
}
