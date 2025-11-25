package models

import (
	"time"
)

var KindSubCategory = "subCategory"

type SubCategory struct {
	ID              int
	CategoryCode    string
	Code            string
	Name            string
	Description     string
	AccountType     string
	ProductTypeCode string
	ProductTypeName string
	Currency        string
	Status          string
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
}

type GetAllSubCategoryParam struct {
	CategoryCode string
}
type CreateSubCategory struct {
	Code            string
	Name            string
	Description     string // Optional
	CategoryCode    string
	AccountType     string
	ProductTypeCode string // Optional
	Currency        string
	Status          string
}

type SubCategoryOut struct {
	Kind            string     `json:"kind" example:"subCategory"`
	ID              int        `json:"-" example:"1"`
	Code            string     `json:"code" example:"10000"`
	Name            string     `json:"name" example:"RETAIL"`
	Description     string     `json:"description" example:"Lender Retail"`
	CategoryCode    string     `json:"categoryCode" example:"222"`
	AccountType     string     `json:"accountType" example:"LENDER_RETAIL"`
	ProductTypeCode string     `json:"productTypeCode" example:"1001"`
	ProductTypeName string     `json:"productTypeName" example:"Group Loan"`
	Currency        string     `json:"currency" example:"IDR"`
	Status          string     `json:"status" example:"active"`
	CreatedAt       *time.Time `json:"createdAt" example:"2006-01-02 15:04:05"`
	UpdatedAt       *time.Time `json:"updatedAt" example:"2006-01-02 15:04:05"`
}

func (c *SubCategory) ToResponse() *SubCategoryOut {
	return &SubCategoryOut{
		Kind:            KindSubCategory,
		Code:            c.Code,
		Name:            c.Name,
		Description:     c.Description,
		CategoryCode:    c.CategoryCode,
		AccountType:     c.AccountType,
		ProductTypeCode: c.ProductTypeCode,
		ProductTypeName: c.ProductTypeName,
		Currency:        c.Currency,
		Status:          c.Status,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}

type (
	CreateSubCategoryRequest struct {
		Code            string `json:"code" validate:"required,numeric,min=5,max=5" example:"11405"`
		Name            string `json:"name" validate:"required,min=1,max=100,noStartEndSpaces,alphanumericMix" example:"Borrower Outstanding - Chickin"`
		Description     string `json:"description" validate:"max=100" example:"Borrower Outstanding - Chickin"`
		CategoryCode    string `json:"categoryCode" validate:"required,numeric,min=3,max=3" example:"114"`
		AccountType     string `json:"accountType" validate:"omitempty" example:"LOAN_ACCOUNT_CHICKIN"`
		ProductTypeCode string `json:"productTypeCode" validate:"omitempty,numeric,min=3,max=5" example:"1001"`
		Currency        string `json:"currency" validate:"omitempty,alpha,min=3,max=3" example:"IDR"`
		Status          string `json:"status" validate:"required,oneof=active inactive" example:"active"`
	}
	CreateSubCategoryResponse struct {
		Kind            string `json:"kind" example:"subCategory"`
		Code            string `json:"code" example:"11405"`
		Name            string `json:"name" example:"Borrower Outstanding - Chickin"`
		Description     string `json:"description,omitempty" example:"Borrower Outstanding - Chickin"`
		CategoryCode    string `json:"categoryCode" example:"114"`
		AccountType     string `json:"accountType,omitempty" example:"LOAN_ACCOUNT_CHICKIN"`
		ProductTypeCode string `json:"productTypeCode,omitempty" example:"1001"`
		Currency        string `json:"currency,omitempty" example:"IDR"`
		Status          string `json:"status" example:"active"`
	}
)

func (c *SubCategory) ToCreateResponse() CreateSubCategoryResponse {
	return CreateSubCategoryResponse{
		Kind:            KindSubCategory,
		Code:            c.Code,
		Name:            c.Name,
		Description:     c.Description,
		CategoryCode:    c.CategoryCode,
		AccountType:     c.AccountType,
		ProductTypeCode: c.ProductTypeCode,
		Currency:        c.Currency,
		Status:          c.Status,
	}
}

type (
	UpdateSubCategoryRequest struct {
		Name            string  `json:"name" validate:"omitempty,min=1,max=100,noStartEndSpaces,alphanumericMix" example:"RETAIL"`
		Description     string  `json:"description" validate:"omitempty,max=100" example:"Lender Retail"`
		Status          string  `json:"status" validate:"omitempty,oneof=active inactive" example:"active"`
		ProductTypeCode *string `json:"productTypeCode" validate:"omitempty,max=5" example:"1001"`
		Currency        *string `json:"currency" validate:"omitempty,max=3" example:"IDR"`
		Code            string  `json:"code" param:"subCategoryCode" validate:"numeric,min=5,max=5" example:"10000"`
	}
	UpdateSubCategoryResponse struct {
		Kind            string  `json:"kind" example:"subCategory"`
		Code            string  `json:"code,omitempty" example:"10000"`
		Name            string  `json:"name,omitempty" example:"RETAIL"`
		Description     string  `json:"description,omitempty" example:"Retail"`
		Status          string  `json:"status,omitempty" example:"active"`
		ProductTypeCode *string `json:"productTypeCode,omitempty" example:"1001"`
		Currency        *string `json:"currency,omitempty" example:"IDR"`
	}
	UpdateSubCategory struct {
		Name            string
		Description     string
		Status          string
		ProductTypeCode *string
		Currency        *string
		Code            string
	}
)

func (e *UpdateSubCategory) ToResponse() *UpdateSubCategoryResponse {
	return &UpdateSubCategoryResponse{
		Kind:            KindSubCategory,
		Code:            e.Code,
		Name:            e.Name,
		Description:     e.Description,
		Status:          e.Status,
		ProductTypeCode: e.ProductTypeCode,
		Currency:        e.Currency,
	}
}
