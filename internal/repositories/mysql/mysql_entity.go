package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"bitbucket.org/Amartha/go-accounting/internal/models"
)

type EntityRepository interface {
	Create(ctx context.Context, in *models.CreateEntityIn) (err error)
	CheckEntityByCode(ctx context.Context, code, status string) (err error)
	GetByCode(ctx context.Context, code string) (*models.Entity, error)
	GetByParams(ctx context.Context, in models.GetEntity) (*models.Entity, error)
	List(ctx context.Context) (*[]models.Entity, error)
	Update(ctx context.Context, in models.UpdateEntity) (err error)
}

type entityRepository sqlRepo

var _ EntityRepository = (*entityRepository)(nil)

// Create implements EntityRepository.
func (er *entityRepository) Create(ctx context.Context, in *models.CreateEntityIn) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := er.r.extractTx(ctx)

	args, err := getFieldValues(*in)
	if err != nil {
		return
	}

	res, err := db.ExecContext(ctx, queryEntityCreate, args...)
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

// CheckEntityByCode check if an entity exists, only return error.
func (er *entityRepository) CheckEntityByCode(ctx context.Context, code, status string) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := er.r.extractTx(ctx)

	var entityCode string
	if status != "" {
		err = db.QueryRowContext(ctx, queryEntityIsExistByCode+queryEntityAndStatus, code, status).Scan(
			&entityCode,
		)
	} else {
		err = db.QueryRowContext(ctx, queryEntityIsExistByCode, code).Scan(
			&entityCode,
		)
	}

	return
}

// GetByCode retrieves an Entity by its code.
func (er *entityRepository) GetByCode(ctx context.Context, code string) (*models.Entity, error) {
	var err error
	defer func() {
		logSQL(ctx, err)
	}()

	db := er.r.extractTx(ctx)

	var entity models.Entity
	err = db.QueryRowContext(ctx, queryEntityGetByCode, code, models.StatusActive).Scan(
		&entity.ID,
		&entity.Code,
		&entity.Name,
		&entity.Description,
		&entity.Status,
		&entity.CreatedAt,
		&entity.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		err = databaseError(err)
		return nil, err
	}

	return &entity, nil
}

// GetByCode retrieves an Entity by its name.
func (er *entityRepository) GetByParams(ctx context.Context, in models.GetEntity) (*models.Entity, error) {
	var err error
	defer func() {
		logSQL(ctx, err)
	}()

	db := er.r.extractTx(ctx)

	var entity models.Entity
	query, args, err := builQueryGetEntityByParam(in)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return nil, err
	}

	err = db.QueryRowContext(ctx, query, args...).Scan(
		&entity.ID,
		&entity.Code,
		&entity.Name,
		&entity.Description,
		&entity.Status,
		&entity.CreatedAt,
		&entity.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// List implements EntityRepository.
func (er *entityRepository) List(ctx context.Context) (*[]models.Entity, error) {
	var err error
	defer func() {
		logSQL(ctx, err)
	}()

	db := er.r.extractTx(ctx)

	// Execute the query with QueryContext
	rows, err := db.QueryContext(ctx, queryEntityList, models.StatusActive)
	if err != nil {
		err = databaseError(err)
		return nil, err
	}
	defer rows.Close()

	// Iterate over the result set and process the data
	var result []models.Entity
	for rows.Next() {
		var entity models.Entity
		if err := rows.Scan(
			&entity.ID,
			&entity.Code,
			&entity.Name,
			&entity.Description,
			&entity.Status,
			&entity.CreatedAt,
			&entity.UpdatedAt,
		); err != nil {
			err = databaseError(err)
			return nil, err
		}
		result = append(result, entity)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		err = databaseError(rows.Err())
		return nil, err
	}

	return &result, nil
}

func (er *entityRepository) Update(ctx context.Context, in models.UpdateEntity) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := er.r.extractTx(ctx)
	args, err := getFieldValues(in)
	if err != nil {
		return
	}
	_, err = db.ExecContext(ctx, queryEntityUpdate, args...)

	return
}
