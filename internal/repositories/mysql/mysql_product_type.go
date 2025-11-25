package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"bitbucket.org/Amartha/go-accounting/internal/models"
)

type ProductTypeRepository interface {
	Create(ctx context.Context, in *models.CreateProductTypeRequest) (err error)
	GetByCode(ctx context.Context, code string) (*models.ProductType, error)
	CheckProductTypeIsExist(ctx context.Context, code string) error
	List(ctx context.Context) ([]models.ProductType, error)
	Update(ctx context.Context, in models.UpdateProductType) (err error)
	GetLatestProductCode(ctx context.Context) (string, error)
}

type productTypeRepository sqlRepo

var _ ProductTypeRepository = (*productTypeRepository)(nil)

func (scr *productTypeRepository) Create(ctx context.Context, in *models.CreateProductTypeRequest) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := scr.r.extractTx(ctx)

	args, err := getFieldValues(*in)
	if err != nil {
		return
	}

	res, err := db.ExecContext(ctx, queryProductTypeCreate, args...)
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

func (r *productTypeRepository) GetByCode(ctx context.Context, code string) (*models.ProductType, error) {
	var err error
	defer func() {
		logSQL(ctx, err)
	}()

	db := r.r.extractTx(ctx)

	var productType models.ProductType
	if err = db.QueryRowContext(ctx, queryProductTypeGetByCode, code, models.StatusActive).Scan(
		&productType.ID,
		&productType.Code,
		&productType.Name,
		&productType.Status,
		&productType.EntityCode,
		&productType.CreatedAt,
		&productType.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		err = databaseError(err)
		return nil, err
	}

	return &productType, nil
}

func (r *productTypeRepository) List(ctx context.Context) ([]models.ProductType, error) {
	var err error
	defer func() {
		logSQL(ctx, err)
	}()

	db := r.r.extractTx(ctx)

	rows, err := db.QueryContext(ctx, queryProductTypeList, models.StatusActive)
	if err != nil {
		err = databaseError(err)
		return nil, err
	}
	defer rows.Close()

	var result []models.ProductType
	for rows.Next() {
		var productType models.ProductType
		if err := rows.Scan(
			&productType.ID,
			&productType.Code,
			&productType.Name,
			&productType.Status,
			&productType.EntityCode,
			&productType.CreatedAt,
			&productType.UpdatedAt,
		); err != nil {
			err = databaseError(err)
			return nil, err
		}
		result = append(result, productType)
	}

	if err := rows.Err(); err != nil {
		err = databaseError(rows.Err())
		return nil, err
	}

	return result, nil
}

func (r *productTypeRepository) CheckProductTypeIsExist(ctx context.Context, code string) error {
	var err error
	defer func() {
		logSQL(ctx, err)
	}()

	db := r.r.extractTx(ctx)

	var id string
	err = db.QueryRowContext(ctx, queryCheckProductTypeIsExist, code).Scan(
		&id,
	)

	return err
}

func (ar *productTypeRepository) Update(ctx context.Context, in models.UpdateProductType) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	query, args, err := queryProductTypeUpdate(in)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return
	}

	_, err = db.ExecContext(ctx, query, args...)

	return
}

func (r *productTypeRepository) GetLatestProductCode(ctx context.Context) (code string, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := r.r.extractTx(ctx)

	err = db.QueryRowContext(ctx, queryGetLatestProductCode).Scan(
		&code,
	)
	if err != nil {
		err = databaseError(err)
		return code, err
	}

	return code, nil
}
