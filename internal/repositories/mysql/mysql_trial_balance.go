package mysql

import (
	"context"
	"fmt"
	"strings"

	"bitbucket.org/Amartha/go-accounting/internal/models"
)

type TrialBalanceRepository interface {
	BulkInsert(ctx context.Context, in []models.CreateTrialBalancePeriod) (err error)
	Close(ctx context.Context, in models.CloseTrialBalanceRequest) error
	GetByPeriod(ctx context.Context, period, entity_code string) (*models.TrialBalancePeriod, error)
	GetFirstPeriodByStatus(ctx context.Context, status string) (*models.TrialBalancePeriod, error)
	GetByPeriodStatus(ctx context.Context, period, status string) ([]models.TrialBalancePeriod, error)
	UpdateTrialBalanceAdjustment(ctx context.Context, in models.CloseTrialBalanceRequest) (err error)
}

type trialBalanceRepository sqlRepo

var _ TrialBalanceRepository = (*trialBalanceRepository)(nil)

func (tr *trialBalanceRepository) BulkInsert(ctx context.Context, in []models.CreateTrialBalancePeriod) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := tr.r.extractTx(ctx)

	valueStrings := []string{}
	valueArgs := []interface{}{}
	for _, req := range in {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs, req.Period)
		valueArgs = append(valueArgs, req.EntityCode)
		valueArgs = append(valueArgs, req.TBFilePath)
		valueArgs = append(valueArgs, req.Status)
		valueArgs = append(valueArgs, req.IsAdjustment)
	}

	query := fmt.Sprintf(queryBulkInsertTrialBalancePeriod, strings.Join(valueStrings, ","))
	res, err := db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		err = databaseError(err)
		return
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		err = databaseError(err)
		return
	}

	if affectedRows == 0 {
		err = databaseError(models.ErrNoRowsAffected)
		return
	}

	return
}

func (tr *trialBalanceRepository) Close(ctx context.Context, in models.CloseTrialBalanceRequest) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := tr.r.extractTx(ctx)
	result, err := db.ExecContext(ctx, queryTrialBalanceClose, models.TrialBalanceStatusClosed, in.ClosedBy, in.Period, in.EntityCode)
	if err != nil {
		return err
	}
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affectedRows == 0 {
		return models.ErrNoRowsAffected
	}

	return nil
}

func (tr *trialBalanceRepository) GetByPeriod(ctx context.Context, period, entity_code string) (*models.TrialBalancePeriod, error) {
	var err error

	defer func() {
		logSQL(ctx, err)
	}()

	db := tr.r.extractTx(ctx)

	var out models.TrialBalancePeriod
	if err = db.QueryRowContext(ctx, queryTrialBalanceGetByPeriod, period, entity_code).Scan(
		&out.ID,
		&out.Period,
		&out.TBFilePath,
		&out.Status,
		&out.ClosedBy,
		&out.IsAdjustment,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &out, nil
}

func (tr *trialBalanceRepository) GetFirstPeriodByStatus(ctx context.Context, status string) (*models.TrialBalancePeriod, error) {
	var err error

	defer func() {
		logSQL(ctx, err)
	}()

	db := tr.r.extractTx(ctx)

	var out models.TrialBalancePeriod
	if err = db.QueryRowContext(ctx, queryGetFirstPeriodByStatus, status).Scan(
		&out.ID,
		&out.Period,
		&out.TBFilePath,
		&out.Status,
		&out.ClosedBy,
		&out.IsAdjustment,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if err == models.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &out, nil
}

func (tr *trialBalanceRepository) UpdateTrialBalanceAdjustment(ctx context.Context, in models.CloseTrialBalanceRequest) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := tr.r.extractTx(ctx)
	result, err := db.ExecContext(ctx, queryUpdateTrialBalanceAdjustment, in.Period)
	if err != nil {
		return err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affectedRows == 0 {
		err = models.ErrNoRowsAffected
		return
	}

	return nil
}

func (tr *trialBalanceRepository) GetByPeriodStatus(ctx context.Context, period, status string) ([]models.TrialBalancePeriod, error) {
	var err error
	defer func() {
		logSQL(ctx, err)
	}()

	db := tr.r.extractTx(ctx)

	rows, err := db.QueryContext(ctx, queryGetByPeriodStatus, period, status)
	if err != nil {
		err = databaseError(err)
		return nil, err
	}
	defer rows.Close()

	var result []models.TrialBalancePeriod
	for rows.Next() {
		var value models.TrialBalancePeriod
		var err = rows.Scan(
			&value.Period,
			&value.EntityCode,
			&value.TBFilePath,
			&value.Status,
			&value.ClosedBy,
			&value.IsAdjustment,
			&value.CreatedAt,
			&value.UpdatedAt,
		)
		if err != nil {
			err = databaseError(err)
			return nil, err
		}
		result = append(result, value)
	}
	if rows.Err() != nil {
		err = databaseError(rows.Err())
		return nil, err
	}

	return result, nil
}
