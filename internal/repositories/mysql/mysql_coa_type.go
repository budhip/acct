package mysql

import (
	"context"
	"errors"

	"bitbucket.org/Amartha/go-accounting/internal/models"
)

type COATypeRepository interface {
	Create(ctx context.Context, in *models.CreateCOATypeIn) (err error)
	CheckCOATypeByCode(ctx context.Context, code string) (err error)
	GetCOATypeByCode(ctx context.Context, code string) (out *models.COAType, err error)
	GetAll(ctx context.Context) ([]models.COAType, error)
	Update(ctx context.Context, in models.UpdateCOAType) (err error)
}

type coaTypeRepository sqlRepo

var _ COATypeRepository = (*coaTypeRepository)(nil)

func (ctr *coaTypeRepository) Create(ctx context.Context, in *models.CreateCOATypeIn) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()
	db := ctr.r.extractTx(ctx)

	args, err := getFieldValues(*in)
	if err != nil {
		return
	}

	res, err := db.ExecContext(ctx, queryCOATypeCreate, args...)
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

func (ctr *coaTypeRepository) CheckCOATypeByCode(ctx context.Context, code string) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()
	db := ctr.r.extractTx(ctx)

	var coaTypeCode string
	err = db.QueryRowContext(ctx, queryCOATypeIsExistByCode, code, models.StatusActive).Scan(
		&coaTypeCode,
	)

	return
}

func (ctr *coaTypeRepository) GetCOATypeByCode(ctx context.Context, code string) (out *models.COAType, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ctr.r.extractTx(ctx)

	var coaType models.COAType
	err = db.QueryRowContext(ctx, queryGetCOATypeByCode, code).Scan(
		&coaType.Code,
		&coaType.Name,
		&coaType.NormalBalance,
		&coaType.Status,
		&coaType.CreatedAt,
		&coaType.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, models.ErrNoRows) {
			return nil, nil
		}
		err = databaseError(err)
		return nil, err
	}

	return &coaType, nil
}

func (ctr *coaTypeRepository) GetAll(ctx context.Context) ([]models.COAType, error) {
	var err error
	defer func() {
		logSQL(ctx, err)
	}()

	db := ctr.r.extractTx(ctx)

	// Execute the query with QueryContext
	rows, err := db.QueryContext(ctx, queryGetAllCOAType, models.StatusActive)
	if err != nil {
		err = databaseError(err)
		return nil, err
	}
	defer rows.Close()

	// Iterate over the result set and process the data
	var result []models.COAType
	for rows.Next() {
		var COAType models.COAType
		if err := rows.Scan(
			&COAType.Code,
			&COAType.Name,
			&COAType.NormalBalance,
			&COAType.Status,
			&COAType.CreatedAt,
			&COAType.UpdatedAt,
		); err != nil {
			err = databaseError(err)
			return nil, err
		}
		result = append(result, COAType)
	}
	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		err = databaseError(rows.Err())
		return nil, err
	}

	return result, nil
}

func (ctr *coaTypeRepository) Update(ctx context.Context, in models.UpdateCOAType) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ctr.r.extractTx(ctx)
	args, err := getFieldValues(in)
	if err != nil {
		return
	}

	_, err = db.ExecContext(ctx, queryCOATypeUpdate, args...)
	return
}
