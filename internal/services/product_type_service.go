package services

import (
	"context"

	"bitbucket.org/Amartha/go-accounting/internal/models"
)

type ProductTypeService interface {
	GetAll(ctx context.Context) (out []models.ProductType, err error)
	Create(ctx context.Context, req models.CreateProductTypeRequest) (output *models.ProductType, err error)
	Update(ctx context.Context, in models.UpdateProductType) (out models.UpdateProductType, err error)
}

type productType service

var _ ProductTypeService = (*productType)(nil)

func (p *productType) GetAll(ctx context.Context) (out []models.ProductType, err error) {
	defer func() {
		logService(ctx, err)
	}()

	out, err = p.srv.mySqlRepo.GetProductTypeRepository().List(ctx)
	if err != nil {
		return
	}

	return
}

func (p *productType) Create(ctx context.Context, req models.CreateProductTypeRequest) (output *models.ProductType, err error) {
	defer func() {
		logService(ctx, err)
	}()

	// check productTypeCode
	productType, err := p.srv.mySqlRepo.GetProductTypeRepository().GetByCode(ctx, req.Code)
	if err != nil {
		return
	}
	if productType != nil {
		err = models.GetErrMap(models.ErrKeyProductTypeCodeIsExist)
		return
	}

	if req.EntityCode != "" {
		// check entityCode
		entity, errGet := p.srv.mySqlRepo.GetEntityRepository().GetByCode(ctx, req.EntityCode)
		if errGet != nil {
			err = errGet
			return
		}
		if entity == nil {
			err = models.GetErrMap(models.ErrKeyEntityCodeNotFound)
			return
		}
	}

	if err = p.srv.mySqlRepo.GetProductTypeRepository().Create(ctx, &req); err != nil {
		return
	}

	output = &models.ProductType{
		Code:       req.Code,
		Name:       req.Name,
		EntityCode: req.EntityCode,
		Status:     req.Status,
	}

	return
}

func (p *productType) Update(ctx context.Context, in models.UpdateProductType) (out models.UpdateProductType, err error) {
	defer func() {
		logService(ctx, err)
	}()

	if err = p.srv.mySqlRepo.GetProductTypeRepository().CheckProductTypeIsExist(ctx, in.Code); err != nil {
		err = checkDatabaseError(err, models.ErrKeyProductTypeNotFound)
		return
	}

	if in.EntityCode != "" {
		if err = p.srv.mySqlRepo.GetEntityRepository().CheckEntityByCode(ctx, in.EntityCode, models.StatusActive); err != nil {
			err = checkDatabaseError(err, models.ErrKeyEntityCodeNotFound)
			return
		}
	}

	if err = p.srv.mySqlRepo.GetProductTypeRepository().Update(ctx, in); err != nil {
		err = checkDatabaseError(err)
		return
	}

	return in, err
}
