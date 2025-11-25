package models

import (
	"time"
)

var KindEntity = "entity"

type Entity struct {
	ID          int
	Code        string
	Name        string
	Description string
	Status      string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

func (e *Entity) ToResponse() *EntityOut {
	return &EntityOut{
		Kind:        KindEntity,
		Code:        e.Code,
		Name:        e.Name,
		Description: e.Description,
		Status:      e.Status,
	}
}

type CreateEntityIn struct {
	Code        string
	Name        string
	Description string
	Status      string
}

type CreateEntityRequest struct {
	Code        string `json:"code" validate:"required,numeric,min=3,max=3" example:"001"`
	Name        string `json:"name" validate:"required,min=1,max=50,nospecial,noStartEndSpaces" example:"AMF"`
	Description string `json:"description" validate:"max=100" example:"PT. Amartha Mikro Fintek"`
	Status      string `json:"status" validate:"required,oneof=active inactive" example:"active"`
}

type EntityOut struct {
	Kind        string     `json:"kind" example:"entity"`
	ID          int        `json:"-" example:"1"`
	Code        string     `json:"code" example:"001"`
	Name        string     `json:"name" example:"AMF"`
	Description string     `json:"description" example:"PT. Amartha Mikro Fintek"`
	Status      string     `json:"status" example:"active"`
	CreatedAt   *time.Time `json:"-" example:"2006-01-02 15:04:05"`
	UpdatedAt   *time.Time `json:"-" example:"2006-01-02 15:04:05"`
}

type (
	UpdateEntityRequest struct {
		Name        string `json:"name" validate:"omitempty,max=50,nospecial,noStartEndSpaces" example:"AMF"`
		Description string `json:"description" validate:"omitempty,max=100" example:"PT. Amartha Mikro Fintek"`
		Status      string `json:"status" validate:"omitempty,oneof=active inactive" example:"active"`
		Code        string `param:"entityCode" validate:"required,numeric,min=3,max=3" example:"001"`
	}
	UpdateEntityResponse struct {
		Kind        string `json:"kind" example:"entity"`
		Code        string `json:"code" example:"001"`
		Name        string `json:"name,omitempty" example:"AMF"`
		Description string `json:"description,omitempty" example:"PT. Amartha Mikro Fintek"`
		Status      string `json:"status,omitempty" example:"active"`
	}

	UpdateEntity struct {
		Name        string
		Description string
		Status      string
		Code        string
	}

	GetEntityRequest struct {
		EntityCode string `query:"entityCode" json:"entityCode" validate:"required_without_all=Name,omitempty" example:"001"`
		Name       string `query:"name" json:"name" validate:"required_without_all=EntityCode,omitempty" example:"AWF"`
	}

	GetEntityResponse struct {
		Kind        string `json:"kind" example:"entity"`
		Code        string `json:"code" example:"001"`
		Name        string `json:"name,omitempty" example:"AMF"`
		Description string `json:"description,omitempty" example:"PT. Amartha Mikro Fintek"`
		Status      string `json:"status,omitempty" example:"active"`
	}

	GetEntity struct {
		Name   string
		Code   string
		Status string
	}
)

func (e *UpdateEntity) ToResponse() *UpdateEntityResponse {
	return &UpdateEntityResponse{
		Kind:        KindEntity,
		Code:        e.Code,
		Name:        e.Name,
		Description: e.Description,
		Status:      e.Status,
	}
}

func (e *Entity) ToResponseGetEntity() *GetEntityResponse {
	return &GetEntityResponse{
		Kind:        KindEntity,
		Code:        e.Code,
		Name:        e.Name,
		Description: e.Description,
		Status:      e.Status,
	}
}
