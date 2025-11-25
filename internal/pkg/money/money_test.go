package money

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/stretchr/testify/assert"
)

func TestFormatAmountToIDR(t *testing.T) {
	type args struct {
		amount decimal.Decimal
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "excpected",
			args: args{
				amount: decimal.NewFromFloat(100000.25),
			},
			want: "100.000,25",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatAmountToIDR(tt.args.amount); got != tt.want {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestFormatAmountToIDRFromDecimal(t *testing.T) {
	type args struct {
		amount decimal.Decimal
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "excpected",
			args: args{
				amount: decimal.NewFromFloat(2013040520360),
			},
			want: "20.130.405.203,60",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatAmountToIDRFromDecimal(tt.args.amount)
			if got != tt.want {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestFormatAmountToBigInt(t *testing.T) {
	type args struct {
		amount   decimal.Decimal
		currency int
	}
	tests := []struct {
		name string
		args args
		want decimal.Decimal
	}{
		{
			name: "excpected",
			args: args{
				amount:   decimal.NewFromFloat(100000.25),
				currency: 2,
			},
			want: decimal.NewFromFloat(10000025),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatAmountToBigInt(tt.args.amount, tt.args.currency)
			assert.Equal(t, tt.want.String(), got.String())
		})
	}
}

func TestFormatBigIntToAmount(t *testing.T) {
	type args struct {
		amount   decimal.Decimal
		currency int
	}
	tests := []struct {
		name string
		args args
		want decimal.Decimal
	}{
		{
			name: "excpected",
			args: args{
				amount:   decimal.NewFromFloat(10000025),
				currency: 2,
			},
			want: decimal.NewFromFloat(100000.25),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatBigIntToAmount(tt.args.amount, tt.args.currency)
			assert.Equal(t, tt.want.String(), got.String())
		})
	}
}
