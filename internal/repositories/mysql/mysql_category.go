package mysql

import (
	"context"
	"database/sql"

	"bitbucket.org/Amartha/go-accounting/internal/models"
)

type CategoryRepository interface {
	CheckCategoryByCode(ctx context.Context, code string) (err error)
	Create(ctx context.Context, in *models.CreateCategoryIn) (err error)
	GetByCode(ctx context.Context, code string) (*models.Category, error)
	List(ctx context.Context) ([]models.Category, error)
	GetByCoaCode(ctx context.Context, coaCode string) (*[]models.CategoryCOA, error)
	Update(ctx context.Context, in models.UpdateCategoryIn) (err error)
}

type categoryRepository sqlRepo

var _ CategoryRepository = (*categoryRepository)(nil)

func (cr *categoryRepository) CheckCategoryByCode(ctx context.Context, code string) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := cr.r.extractTx(ctx)

	var categoryCode string
	err = db.QueryRowContext(ctx, queryCategoryIsExistByCode, code, models.StatusActive).Scan(
		&categoryCode,
	)
	if err != nil {
		return
	}

	return
}

// Create implements CategoryRepository.
func (r *categoryRepository) Create(ctx context.Context, in *models.CreateCategoryIn) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := r.r.extractTx(ctx)

	args, err := getFieldValues(*in)
	if err != nil {
		return
	}

	res, err := db.ExecContext(ctx, queryCategoryCreate, args...)
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

// GetByCode implements CategoryRepository.
func (r *categoryRepository) GetByCode(ctx context.Context, code string) (*models.Category, error) {
	var err error
	defer func() {
		logSQL(ctx, err)
	}()
	db := r.r.extractTx(ctx)
	var category models.Category
	err = db.QueryRowContext(ctx, queryCategoryGetByCode, code).Scan(
		&category.ID,
		&category.Code,
		&category.Name,
		&category.Description,
		&category.CoaTypeCode,
		&category.Status,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		err = databaseError(err)
		return nil, err
	}
	return &category, nil
}

// List implements CategoryRepository.
func (r *categoryRepository) List(ctx context.Context) ([]models.Category, error) {
	var err error
	defer func() {
		logSQL(ctx, err)
	}()

	db := r.r.extractTx(ctx)

	// Execute the query with QueryContext
	rows, err := db.QueryContext(ctx, queryCategoryList, models.StatusActive)
	if err != nil {
		err = databaseError(err)
		return nil, err
	}
	defer rows.Close()

	// Iterate over the result set and process the data
	var result []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(
			&category.ID,
			&category.Code,
			&category.Name,
			&category.Description,
			&category.CoaTypeCode,
			&category.Status,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, category)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		err = databaseError(rows.Err())
		return nil, err
	}

	return result, nil
}

func (r *categoryRepository) GetByCoaCode(ctx context.Context, coaCode string) (*[]models.CategoryCOA, error) {
	var err error
	defer func() {
		logSQL(ctx, err)
	}()

	db := r.r.extractTx(ctx)

	rows, err := db.QueryContext(ctx, queryCategoryListByCoa, coaCode)
	if err != nil {
		err = databaseError(err)
		return nil, err
	}
	defer rows.Close()

	var result []models.CategoryCOA
	for rows.Next() {
		var category models.CategoryCOA
		if err := rows.Scan(
			&category.Code,
			&category.Name,
		); err != nil {
			err = databaseError(err)
			return nil, err
		}
		result = append(result, category)
	}
	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		err = databaseError(rows.Err())
		return nil, err
	}

	return &result, nil
}

func (r *categoryRepository) Update(ctx context.Context, in models.UpdateCategoryIn) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()
	db := r.r.extractTx(ctx)
	args, err := getFieldValues(in)
	if err != nil {
		return
	}
	_, err = db.ExecContext(ctx, queryCeCategoryUpdate, args...)
	return
}
