package models

import "time"

const (
	KindCOAType      = "coaType"
	COATypeAsset     = "AST"
	COATypeLiability = "LIA"
)

type CreateCOATypeRequest struct {
	Code          string `json:"code" validate:"required,alpha,min=3,max=3" example:"ASS"`
	Name          string `json:"name" validate:"required,min=1,max=50" example:"Asset"`
	NormalBalance string `json:"normalBalance" validate:"required,oneof=debit credit" example:"debit"`
	Status        string `json:"status" validate:"required,oneof=active inactive" example:"active"`
}

type CreateCOATypeIn struct {
	Code          string
	Name          string
	NormalBalance string
	Status        string
}

type COATypeOut struct {
	Kind          string `json:"kind" example:"coaType"`
	Code          string `json:"code" example:"ASS"`
	Name          string `json:"name" example:"Asset"`
	NormalBalance string `json:"normalBalance" example:"debit"`
	Status        string `json:"status" example:"active"`
}

type COAType struct {
	Code          string
	Name          string
	NormalBalance string
	Status        string
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
}

func (c *COAType) ToResponse() *COATypeOut {
	return &COATypeOut{
		Kind:          KindCOAType,
		Code:          c.Code,
		Name:          c.Name,
		NormalBalance: c.NormalBalance,
		Status:        c.Status,
	}
}

type COATypeCategory struct {
	Kind          string        `json:"kind" example:"coaType"`
	Code          string        `json:"code" example:"AST"`
	Name          string        `json:"name" example:"Asset"`
	NormalBalance string        `json:"normalBalance" example:"debit"`
	Categories    []CategoryCOA `json:"categoryCode"`
	Status        string        `json:"status"  example:"active"`
	CreatedAt     *time.Time    `json:"createdAt" example:"2006-01-02 15:04:05"`
	UpdatedAt     *time.Time    `json:"updatedAt" example:"2006-01-02 15:04:05"`
}

type CategoryCOA struct {
	Code string `json:"code" example:"222"`
	Name string `json:"name" example:"LENDER"`
}

type (
	UpdateCOATypeRequest struct {
		Code          string `param:"coaTypeCode" validate:"required,alpha,min=3,max=3" example:"ASS"`
		Name          string `json:"name" validate:"omitempty,min=1,max=50" example:"Asset"`
		NormalBalance string `json:"normalBalance" validate:"omitempty,oneof=debit credit" example:"debit"`
		Status        string `json:"status" validate:"omitempty,oneof=active inactive" example:"active"`
	}
	UpdateCOATypeResponse struct {
		Kind          string `json:"kind" example:"coaType"`
		Code          string `json:"code" example:"ASS"`
		Name          string `json:"name,omitempty" example:"Asset"`
		NormalBalance string `json:"normalBalance,omitempty" example:"debit"`
		Status        string `json:"status,omitempty" example:"active"`
	}

	UpdateCOAType struct {
		Name          string
		NormalBalance string
		Status        string
		Code          string
	}
)

func (u *UpdateCOAType) ToResponse() *UpdateCOATypeResponse {
	return &UpdateCOATypeResponse{
		Kind:          KindCOAType,
		Code:          u.Code,
		Name:          u.Name,
		NormalBalance: u.NormalBalance,
		Status:        u.Status,
	}
}
