package money

import (
	"strings"

	"github.com/Rhymond/go-money"
	"github.com/shopspring/decimal"
)

func FormatAmountToIDR(amount decimal.Decimal) string {
	new, _ := amount.Float64()
	idr := money.NewFromFloat(new, money.IDR)
	result := strings.ReplaceAll(idr.Display(), "Rp", "")
	return result
}

// FormatAmountToIDRFromDecimal formats a Decimal From Big Int Value to a string in IDR format.
// The formatter is set to use commas for thousands and periods for decimal
// separators, and does not include a currency symbol.
func FormatAmountToIDRFromDecimal(amount decimal.Decimal) string {
	formatter := money.NewFormatter(2, ",", ".", "", "1$")
	return formatter.Format(amount.IntPart())
}

func FormatAmountToIDRFromDecimalGCS(amount decimal.Decimal) string {
	formatter := money.NewFormatter(2, ".", ",", "", "1$")
	return formatter.Format(amount.IntPart())
}

func FormatAmountToBigInt(amount decimal.Decimal, currency int) decimal.Decimal {
	return amount.Mul(
		decimal.NewFromInt(10).Pow(
			decimal.NewFromInt(
				int64(currency),
			),
		),
	)
}

func FormatBigIntToAmount(amount decimal.Decimal, currency int) decimal.Decimal {
	return amount.Div(
		decimal.NewFromInt(10).Pow(
			decimal.NewFromInt(
				int64(currency),
			),
		),
	)
}
