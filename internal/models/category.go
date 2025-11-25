package models

import "time"

var (
	KindCategory    = "category"
	CategoryCode131 = "131"
)

type Category struct {
	ID          int
	Code        string
	Name        string
	Description string
	CoaTypeCode string
	Status      string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

func (c *Category) ConvertToCategoryOut() *CategoryOut {
	return &CategoryOut{
		Kind:        KindCategory,
		Code:        c.Code,
		Name:        c.Name,
		Description: c.Description,
		CoaTypeCode: c.CoaTypeCode,
		Status:      c.Status,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

type DoCreateCategoryRequest struct {
	Code        string `json:"code" validate:"required,min=3,max=3,numeric" example:"222"`
	Name        string `json:"name" validate:"required,min=1,max=100,noStartEndSpaces,alphanumericMix" example:"RETAIL"`
	Description string `json:"description" validate:"max=100" example:"Lender Retail"`
	CoaTypeCode string `json:"coaTypeCode" validate:"required,min=3,max=3" example:"AST"`
	Status      string `json:"status" validate:"required,oneof=active inactive" example:"active"`
}

type CreateCategoryIn struct {
	Code        string
	Name        string
	Description string
	CoaTypeCode string
	Status      string
}

type CategoryOut struct {
	Kind        string     `json:"kind" example:"category"`
	Code        string     `json:"code" example:"222"`
	Name        string     `json:"name" example:"LENDER"`
	Description string     `json:"description" example:"Lender Account"`
	CoaTypeCode string     `json:"coaTypeCode" example:"AST"`
	Status      string     `json:"status" example:"active"`
	CreatedAt   *time.Time `json:"createdAt,omitempty" example:"2006-01-02 15:04:05"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty" example:"2006-01-02 15:04:05"`
}

type (
	DoUpdateCategoryRequest struct {
		Name        string `json:"name" validate:"omitempty,min=1,max=100,noStartEndSpaces,alphanumericMix" example:"RETAIL"`
		Description string `json:"description" validate:"omitempty,max=100" example:"Lender Retail"`
		CoaTypeCode string `json:"coaTypeCode" validate:"omitempty,min=3,max=3" example:"AST"`
		Code        string `param:"categoryCode" json:"code" validate:"required,min=3,max=3,numeric" example:"222"`
	}
	DoUpdateCategoryResponse struct {
		Kind        string `json:"kind" example:"category"`
		Code        string `json:"code" example:"222"`
		Name        string `json:"name,omitempty"  example:"RETAIL"`
		Description string `json:"description,omitempty"  example:"Lender Retail"`
		CoaTypeCode string `json:"coaTypeCode,omitempty"  example:"AST"`
	}
	UpdateCategoryIn struct {
		Name        string
		Description string
		CoaTypeCode string
		Code        string
	}
)

func (c *UpdateCategoryIn) ToResponse() DoUpdateCategoryResponse {
	return DoUpdateCategoryResponse{
		Kind:        KindCategory,
		Code:        c.Code,
		Name:        c.Name,
		Description: c.Description,
		CoaTypeCode: c.CoaTypeCode,
	}
}
