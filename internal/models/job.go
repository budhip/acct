package models

import (
	"fmt"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
)

type (
	RunTrialBalanceRequest struct {
		Date         string `json:"date" validate:"required" example:"2025-01-01"`
		IsAdjustment bool   `json:"isAdjustment" example:"false"`
	}

	RunTrialBalanceFilter struct {
		Date         time.Time
		IsAdjustment bool
	}
)

func (req RunTrialBalanceRequest) ToFilterOpts() (*RunTrialBalanceFilter, error) {
	opts := &RunTrialBalanceFilter{
		IsAdjustment: req.IsAdjustment,
	}

	adjustmentDate, err := atime.ParseStringToDatetime(atime.DateFormatYYYYMMDD, req.Date)
	if err != nil {
		err = GetErrMap(ErrKeyInvalidFormatDate, fmt.Sprintf("date %s format must be YYYY-MM-DD", req.Date))
		return nil, err
	}
	opts.Date = adjustmentDate

	return opts, nil
}

type (
	AdjustmentTrialBalanceRequest struct {
		AdjustmentDate string `json:"adjustmentDate" validate:"required" example:"2025-01-01"`
		IsManual       bool   `json:"isManual" example:"false"`
	}

	AdjustmentTrialBalanceFilter struct {
		AdjustmentDate time.Time
		IsManual       bool
	}
)

func (req AdjustmentTrialBalanceRequest) ToFilterOpts() (*AdjustmentTrialBalanceFilter, error) {
	opts := &AdjustmentTrialBalanceFilter{
		IsManual: req.IsManual,
	}

	adjustmentDate, err := atime.ParseStringToDatetime(atime.DateFormatYYYYMMDD, req.AdjustmentDate)
	if err != nil {
		err = GetErrMap(ErrKeyInvalidFormatDate, fmt.Sprintf("date %s format must be YYYY-MM-DD", req.AdjustmentDate))
		return nil, err
	}
	opts.AdjustmentDate = adjustmentDate

	return opts, nil
}
