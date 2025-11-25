package mysql

import (
	"context"
	"database/sql"

	"bitbucket.org/Amartha/go-accounting/internal/models"
)

type SubCategoryRepository interface {
	CheckSubCategoryByCodeAndCategoryCode(ctx context.Context, code, categoryCode string) (*models.SubCategory, error)
	GetByCode(ctx context.Context, code string) (*models.SubCategory, error)
	GetByAccountType(ctx context.Context, accountType string) (*models.SubCategory, error)
	Create(ctx context.Context, in *models.CreateSubCategory) (err error)
	GetAll(ctx context.Context, param models.GetAllSubCategoryParam) (*[]models.SubCategory, error)
	Update(ctx context.Context, in models.UpdateSubCategory) (err error)
	GetLatestSubCategCode(ctx context.Context, categCode string) (code string, err error)
}

type subCategoryRepository sqlRepo

var _ SubCategoryRepository = (*subCategoryRepository)(nil)

func (scr *subCategoryRepository) CheckSubCategoryByCodeAndCategoryCode(ctx context.Context, code, categoryCode string) (*models.SubCategory, error) {
	var err error
	defer func() {
		logSQL(ctx, err)
	}()

	db := scr.r.extractTx(ctx)

	var subCategory models.SubCategory
	err = db.QueryRowContext(ctx, querySubCategoryIsExistByCode, code, categoryCode, models.StatusActive).Scan(
		&subCategory.ID,
		&subCategory.Code,
		&subCategory.Name,
		&subCategory.Description,
		&subCategory.CategoryCode,
		&subCategory.AccountType,
		&subCategory.ProductTypeCode,
		&subCategory.Currency,
		&subCategory.Status,
		&subCategory.CreatedAt,
		&subCategory.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &subCategory, nil
}

// GetByCode retrieves a SubCategory by its code.
func (scr *subCategoryRepository) GetByCode(ctx context.Context, code string) (*models.SubCategory, error) {
	var err error
	defer func() {
		logSQL(ctx, err)
	}()

	db := scr.r.extractTx(ctx)

	var subCategory models.SubCategory
	err = db.QueryRowContext(ctx, querySubCategoryGetByCode, code).Scan(
		&subCategory.ID,
		&subCategory.Code,
		&subCategory.Name,
		&subCategory.Description,
		&subCategory.CategoryCode,
		&subCategory.AccountType,
		&subCategory.ProductTypeCode,
		&subCategory.Currency,
		&subCategory.Status,
		&subCategory.CreatedAt,
		&subCategory.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		err = databaseError(err)
		return nil, err
	}

	return &subCategory, nil
}

// Create implements SubCategoryRepository.
func (scr *subCategoryRepository) Create(ctx context.Context, in *models.CreateSubCategory) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := scr.r.extractTx(ctx)

	args, err := getFieldValues(*in)
	if err != nil {
		return
	}

	res, err := db.ExecContext(ctx, querySubCategoryCreate, args...)
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

// GetAll retrieves all SubCategory
func (scr *subCategoryRepository) GetAll(ctx context.Context, param models.GetAllSubCategoryParam) (*[]models.SubCategory, error) {
	var err error
	defer func() {
		logSQL(ctx, err)
	}()

	db := scr.r.extractTx(ctx)

	values := []interface{}{}
	values = append(values, models.StatusActive)
	tempQuery := queryGetAllSubCategory
	if param.CategoryCode != "" {
		tempQuery += " AND acs.category_code = ?"
		values = append(values, param.CategoryCode)
	}
	tempQuery += " ORDER BY acs.code ASC;"

	rows, err := db.QueryContext(ctx, tempQuery, values...)
	if err != nil {
		err = databaseError(err)
		return nil, err
	}
	defer rows.Close()

	var result []models.SubCategory
	for rows.Next() {
		var value models.SubCategory
		var err = rows.Scan(
			&value.ID,
			&value.CategoryCode,
			&value.Code,
			&value.Name,
			&value.Description,
			&value.AccountType,
			&value.ProductTypeCode,
			&value.ProductTypeName,
			&value.Currency,
			&value.Status,
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

	return &result, nil
}

func (scr *subCategoryRepository) GetByAccountType(ctx context.Context, accountType string) (*models.SubCategory, error) {
	var err error
	defer func() {
		logSQL(ctx, err)
	}()

	db := scr.r.extractTx(ctx)

	var subCategory models.SubCategory
	err = db.QueryRowContext(ctx, querySubCategoryGetByAccountType, accountType, models.StatusActive).Scan(
		&subCategory.ID,
		&subCategory.Code,
		&subCategory.Name,
		&subCategory.Description,
		&subCategory.AccountType,
		&subCategory.ProductTypeCode,
		&subCategory.CategoryCode,
		&subCategory.Currency,
		&subCategory.Status,
		&subCategory.CreatedAt,
		&subCategory.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &subCategory, nil
}

func (scr *subCategoryRepository) Update(ctx context.Context, in models.UpdateSubCategory) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	query := querySubCategoryUpdate
	args := []interface{}{}
	args = append(args, in.Name, in.Description, in.Status)
	if in.ProductTypeCode != nil {
		query += " default_product_type_code = ?, "
		args = append(args, *in.ProductTypeCode)
	}
	if in.Currency != nil {
		query += " default_currency = ?, "
		args = append(args, *in.Currency)
	}
	query += querySubCategoryWhere
	args = append(args, in.Code)

	db := scr.r.extractTx(ctx)
	_, err = db.ExecContext(ctx, query, args...)

	return
}

func (scr *subCategoryRepository) GetLatestSubCategCode(ctx context.Context, categCode string) (code string, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := scr.r.extractTx(ctx)

	err = db.QueryRowContext(ctx, queryGetLatestSubCategoryCode, categCode).Scan(
		&code,
	)
	if err != nil {
		err = databaseError(err)
		return code, err
	}

	return code, nil
}
