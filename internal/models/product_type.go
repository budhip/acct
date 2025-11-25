package models

import (
	"time"
)

const kindProductType = "productType"

type CreateProductTypeRequest struct {
	Code       string `json:"code" validate:"required,numeric,min=3,max=5" example:"1001"`
	Name       string `json:"name" validate:"required,min=1,max=100,nospecial,noStartEndSpaces" example:"Chickin"`
	Status     string `json:"status" validate:"required,oneof=active inactive" example:"active"`
	EntityCode string `json:"entityCode" validate:"omitempty,numeric,min=3,max=3" example:"001"`
}

type ProductType struct {
	ID          int
	Code        string
	Name        string
	Description string
	Status      string
	EntityCode  string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

func (p *ProductType) ToResponse() ProductTypeOut {
	return ProductTypeOut{
		Kind:       kindProductType,
		Code:       p.Code,
		Name:       p.Name,
		Status:     p.Status,
		EntityCode: p.EntityCode,
	}
}

type ProductTypeOut struct {
	Kind       string     `json:"kind" example:"productType"`
	ID         int        `json:"-" example:"1"`
	Code       string     `json:"code" example:"1001"`
	Name       string     `json:"name" example:"Chickin"`
	Status     string     `json:"status" example:"active"`
	EntityCode string     `json:"entityCode" example:"001"`
	CreatedAt  *time.Time `json:"-" example:"2006-01-02 15:04:05"`
	UpdatedAt  *time.Time `json:"-" example:"2006-01-02 15:04:05"`
}

type (
	UpdateProductTypeRequest struct {
		Name       string `json:"name" validate:"omitempty,min=1,max=100,nospecial,noStartEndSpaces" example:"Chickin"`
		Status     string `json:"status" validate:"omitempty,oneof=active inactive" example:"active"`
		Code       string `param:"productTypeCode" json:"code" validate:"required,numeric,min=3,max=5" example:"1001"`
		EntityCode string `json:"entityCode" validate:"omitempty,numeric,min=3,max=3" example:"001"`
	}
	UpdateProductTypeResponse struct {
		Kind       string `json:"kind" example:"productType"`
		Code       string `json:"code" example:"1001"`
		Name       string `json:"name,omitempty" example:"Chickin"`
		Status     string `json:"status,omitempty" example:"active"`
		EntityCode string `json:"entityCode,omitempty" example:"001"`
	}
	UpdateProductType struct {
		Name       string
		Status     string
		Code       string
		EntityCode string
	}
)

func (p *UpdateProductType) ToResponse() *UpdateProductTypeResponse {
	return &UpdateProductTypeResponse{
		Kind:       kindProductType,
		Code:       p.Code,
		Name:       p.Name,
		Status:     p.Status,
		EntityCode: p.EntityCode,
	}
}
