package validation

import (
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestValidateStruct(t *testing.T) {
	type args struct {
		toValidate interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success DoCreateAccountRequest",
			args: args{
				toValidate: models.CreateEntityRequest{
					Code:        "001",
					Name:        "Lender",
					Description: "Lender Yang Baik",
					Status:      models.StatusActive,
				},
			},
			wantErr: false,
		},
		{
			name: "validate DoCreateAccountRequest",
			args: args{
				toValidate: models.CreateEntityRequest{
					Code:        "0001",
					Name:        "Lender",
					Description: "Lender Yang Baik",
					Status:      models.StatusActive,
				},
			},
			wantErr: true,
		},
		{
			name: "validate CreateSubCategoryRequest",
			args: args{
				toValidate: models.CreateSubCategoryRequest{
					Code:        "100",
					Name:        "001",
					Description: "IDR",
				},
			},
			wantErr: true,
		},
		{
			name: "validate error not register",
			args: args{
				toValidate: struct {
					Name string `json:"name" validate:"required,date"`
				}{
					Name: "12345678901234",
				},
			},
			wantErr: true,
		},
		{
			name: "validate Datetime",
			args: args{
				toValidate: struct {
					Datetime string `json:"datetime" validate:"required,datetime"`
				}{
					Datetime: "2006-01-01 15:04:",
				},
			},
			wantErr: true,
		},
		{
			name: "validate json",
			args: args{
				toValidate: struct {
					Datetime string `json:"-" validate:"required,datetime"`
				}{
					Datetime: "2006-01-01 15:04:",
				},
			},
			wantErr: true,
		},
		{
			name: "success validate alphanumericMix",
			args: args{
				toValidate: struct {
					Name string `json:"name" validate:"required,alphanumericMix"`
				}{
					Name: "/.,:;'\"“”‘’ajaiVSDsdsdsa1233-2kns dsk-()",
				},
			},
			wantErr: false,
		},
		{
			name: "error validate alphanumericMix",
			args: args{
				toValidate: struct {
					Name string `json:"name" validate:"required,alphanumericMix"`
				}{
					Name: "/.,:;'\"“”‘’aj@iVSDsdsdsa1233-2kns dsk-()",
				},
			},
			wantErr: true,
		},
		{
			name: "success validate alphanumDashUscore",
			args: args{
				toValidate: struct {
					Name string `json:"name" validate:"required,alphanumDashUscore"`
				}{
					Name: "Testing-123_testing",
				},
			},
			wantErr: false,
		},
		{
			name: "error validate alphanumDashUscore",
			args: args{
				toValidate: struct {
					Name string `json:"name" validate:"required,alphanumDashUscore"`
				}{
					Name: "Testing-1 23_testing",
				},
			},
			wantErr: true,
		},
		{
			name: "error validate maxnospace",
			args: args{
				toValidate: struct {
					Name string `json:"name" validate:"required,maxnospace=15"`
				}{
					Name: "Testing-1 23_testing",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.args.toValidate)
			if err != nil {
				t.Log(err.Error())
			}
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
