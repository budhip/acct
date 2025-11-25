package services

import (
	"context"

	"bitbucket.org/Amartha/go-accounting/internal/models"
)

type COATypeService interface {
	Create(ctx context.Context, in models.CreateCOATypeIn) (out *models.COAType, err error)
	GetAll(ctx context.Context) ([]models.COATypeCategory, error)
	Update(ctx context.Context, in models.UpdateCOAType) (out models.UpdateCOAType, err error)
}

type coaTypes service

var _ COATypeService = (*coaTypes)(nil)

func (c coaTypes) Create(ctx context.Context, in models.CreateCOATypeIn) (out *models.COAType, err error) {
	defer func() {
		logService(ctx, err)
	}()

	// check coa type exists, regardless of status
	coaType, err := c.srv.mySqlRepo.GetCOATypeRepository().GetCOATypeByCode(ctx, in.Code)
	if err != nil {
		return nil, err
	}
	if coaType != nil {
		err = models.GetErrMap(models.ErrKeyDataIsExist)
		return
	}

	// insert coa type
	if err = c.srv.mySqlRepo.GetCOATypeRepository().Create(ctx, &in); err != nil {
		return
	}

	out = &models.COAType{
		Code:          in.Code,
		Name:          in.Name,
		NormalBalance: in.NormalBalance,
		Status:        in.Status,
	}

	return
}

func (c coaTypes) GetAll(ctx context.Context) (out []models.COATypeCategory, err error) {
	result, err := c.srv.mySqlRepo.GetCOATypeRepository().GetAll(ctx)
	if err != nil {
		return
	}

	for _, v := range result {
		coaType := models.COATypeCategory{
			Kind:          models.KindCOAType,
			Code:          v.Code,
			Name:          v.Name,
			NormalBalance: v.NormalBalance,
			Status:        v.Status,
			CreatedAt:     v.CreatedAt,
			UpdatedAt:     v.UpdatedAt,
		}
		cat, err := c.srv.mySqlRepo.GetCategoryRepository().GetByCoaCode(ctx, v.Code)
		if err != nil {
			return out, err
		}
		coaType.Categories = *cat
		out = append(out, coaType)
	}

	return
}

func (c coaTypes) Update(ctx context.Context, in models.UpdateCOAType) (out models.UpdateCOAType, err error) {
	defer func() {
		logService(ctx, err)
	}()

	// Check if not exist
	coaTypeData, err := c.srv.mySqlRepo.GetCOATypeRepository().GetCOATypeByCode(ctx, in.Code)
	if err != nil {
		return
	}
	if coaTypeData == nil {
		err = models.GetErrMap(models.ErrKeyCoaTypeNotFound)
		return
	}

	if in.Name == "" && in.Status == "" && in.NormalBalance == "" {
		return in, nil
	}

	if err = c.srv.mySqlRepo.GetCOATypeRepository().Update(ctx, in); err != nil {
		err = checkDatabaseError(err)
		return
	}

	return in, err
}
