package services

import (
	"context"

	"bitbucket.org/Amartha/go-accounting/internal/models"
)

type CategoryService interface {
	Create(ctx context.Context, req models.CreateCategoryIn) (output *models.Category, err error)
	GetAll(ctx context.Context) (output *[]models.Category, err error)
	Update(ctx context.Context, req models.UpdateCategoryIn) (output models.UpdateCategoryIn, err error)
}

type category service

var _ CategoryService = (*category)(nil)

/*
	create new category

1. validate category code is exist & status = active
2. validate coa type code is exist & status = active
3. insert into acct_category
*/
func (s *category) Create(ctx context.Context, req models.CreateCategoryIn) (output *models.Category, err error) {
	defer func() {
		logService(ctx, err)
	}()

	exist, err := s.srv.mySqlRepo.GetCategoryRepository().GetByCode(ctx, req.Code)
	if err != nil {
		err = checkDatabaseError(err)
		return
	}
	if exist != nil {
		err = models.GetErrMap(models.ErrKeyDataIsExist)
		return
	}

	if err := s.srv.mySqlRepo.GetCOATypeRepository().CheckCOATypeByCode(ctx, req.CoaTypeCode); err != nil {
		err = checkDatabaseError(err, models.ErrKeyCoaTypeNotFound)
		return nil, err
	}

	if err = s.srv.mySqlRepo.GetCategoryRepository().Create(ctx, &req); err != nil {
		return
	}

	output = &models.Category{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		CoaTypeCode: req.CoaTypeCode,
		Status:      req.Status,
	}

	return
}

/*
	get all category

1. get all data from acct_category with status = active
*/
func (s *category) GetAll(ctx context.Context) (output *[]models.Category, err error) {
	defer func() {
		logService(ctx, err)
	}()

	categories, err := s.srv.mySqlRepo.GetCategoryRepository().List(ctx)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyUnableToGetData)
		return
	}
	output = &categories

	return
}

func (s *category) Update(ctx context.Context, in models.UpdateCategoryIn) (out models.UpdateCategoryIn, err error) {
	defer func() {
		logService(ctx, err)
	}()

	exist, err := s.srv.mySqlRepo.GetCategoryRepository().GetByCode(ctx, in.Code)
	if err != nil {
		return
	}
	if exist == nil {
		err = models.GetErrMap(models.ErrKeyCategoryCodeNotFound)
		return
	}

	if in.Name == "" && in.Description == "" && in.CoaTypeCode == "" {
		return in, nil
	}

	if in.CoaTypeCode != "" {
		if err = s.srv.mySqlRepo.GetCOATypeRepository().CheckCOATypeByCode(ctx, in.CoaTypeCode); err != nil {
			err = checkDatabaseError(err, models.ErrKeyCoaTypeNotFound)
			return
		}
	}

	if err = s.srv.mySqlRepo.GetCategoryRepository().Update(ctx, in); err != nil {
		err = checkDatabaseError(err)
		return
	}

	return in, nil
}
