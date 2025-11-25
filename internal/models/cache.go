package models

import (
	"context"
	"fmt"
	"time"
)

type GetOrSetCacheOpts[T any] struct {
	Ctx      context.Context
	Key      string
	TTL      time.Duration
	Callback func() (T, error)
}

func GenerateTrialBalanceDetailsCountCacheKey(opts TrialBalanceDetailsFilterOptions) string {
	return fmt.Sprintf("go_accounting_count_tb_acc_list:%s:%s:%s", opts.SubCategoryCode, opts.SubCategoryCode, opts.Search)
}
