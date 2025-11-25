package services

import (
	"context"
	"fmt"
	"strings"

	"bitbucket.org/Amartha/go-accounting/internal/repositories/cache"

	"bitbucket.org/Amartha/go-accounting/internal/models"
)

type EntityService interface {
	Create(ctx context.Context, req models.CreateEntityIn) (out *models.Entity, err error)
	GetAll(ctx context.Context) (out *[]models.Entity, err error)
	Update(ctx context.Context, req models.UpdateEntity) (out models.UpdateEntity, err error)
	GetByParam(ctx context.Context, in models.GetEntity) (out *models.Entity, err error)
}

type entity service

var _ EntityService = (*entity)(nil)

// Create implements EntityService.
func (es *entity) Create(ctx context.Context, req models.CreateEntityIn) (out *models.Entity, err error) {
	defer func() {
		logService(ctx, err)
	}()

	// Check exist
	exist, err := es.srv.mySqlRepo.GetEntityRepository().GetByCode(ctx, req.Code)
	if err != nil {
		return
	}
	if exist != nil {
		err = models.ErrDataExist
		return
	}

	// Insert
	if err = es.srv.mySqlRepo.GetEntityRepository().Create(ctx, &req); err != nil {
		return
	}

	out = &models.Entity{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
	}

	return
}

// GetAll implements EntityService.
func (es *entity) GetAll(ctx context.Context) (out *[]models.Entity, err error) {
	defer func() {
		logService(ctx, err)
	}()

	// Get data
	out, err = es.srv.mySqlRepo.GetEntityRepository().List(ctx)
	if err != nil {
		return
	}

	return
}

func (es *entity) Update(ctx context.Context, in models.UpdateEntity) (out models.UpdateEntity, err error) {
	defer func() {
		logService(ctx, err)
	}()

	// Check if not exist
	if err = es.srv.mySqlRepo.GetEntityRepository().CheckEntityByCode(ctx, in.Code, ""); err != nil {
		err = checkDatabaseError(err, models.ErrKeyEntityCodeNotFound)
		return
	}

	if in.Name == "" && in.Description == "" && in.Status == "" {
		return in, nil
	}

	if err = es.srv.mySqlRepo.GetEntityRepository().Update(ctx, in); err != nil {
		err = checkDatabaseError(err)
		return
	}

	return in, err
}

// GetEntityByParam implements EntityService. Where you can get entity details by param requested
func (es *entity) GetByParam(ctx context.Context, in models.GetEntity) (out *models.Entity, err error) {
	defer func() {
		logService(ctx, err)
	}()

	in.Status = models.StatusActive
	in.Name = strings.ToLower(in.Name)
	out, err = cache.GetOrSet(es.srv.cacheRepo, models.GetOrSetCacheOpts[*models.Entity]{
		Ctx: ctx,
		Key: pasEntityKey(fmt.Sprintf("%s%s", in.Code, in.Name)),
		TTL: es.srv.conf.CacheTTL.GetEntity,
		Callback: func() (*models.Entity, error) {
			out, err = es.srv.mySqlRepo.GetEntityRepository().GetByParams(ctx, in)
			if err != nil {
				return nil, checkDatabaseError(err)
			}
			return out, nil
		},
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}
