package mysql

import (
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"github.com/stretchr/testify/assert"
)

func Test_getFieldValues(t *testing.T) {
	type args struct {
		i interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				i: models.CreateAccount{},
			},
			wantErr: false,
		},
		{
			name: "errpr",
			args: args{
				i: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getFieldValues(tt.args.i)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
