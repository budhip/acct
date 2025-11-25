package godbledger

import (
	"github.com/darcys22/godbledger/godbledger/core"
)

type (
	Transaction *core.Transaction
	Account     *core.Account
	User        *core.User
	Split       *core.Split
	Currency    *core.Currency
)

var (
	UserSystem = &core.User{
		Id:   "system",
		Name: "System",
	}
	CurrencyIDR = &core.Currency{
		Name:     "IDR",
		Decimals: 2,
	}
	CurrencyUSD = &core.Currency{
		Name:     "USD",
		Decimals: 2,
	}
)
