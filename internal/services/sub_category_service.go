package services

import (
	"context"

	"bitbucket.org/Amartha/go-accounting/internal/pkg/gofptransaction"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/mysql"

	"bitbucket.org/Amartha/go-accounting/internal/models"
)

type SubCategoryService interface {
	Create(ctx context.Context, req models.CreateSubCategory) (out *models.SubCategory, err error)
	GetAll(ctx context.Context, param models.GetAllSubCategoryParam) (out *[]models.SubCategory, err error)
	GetByAccountType(ctx context.Context, accountType string) (out *models.SubCategory, err error)
	Update(ctx context.Context, in models.UpdateSubCategory) (out models.UpdateSubCategory, err error)
}

type subCategory service

var _ SubCategoryService = (*subCategory)(nil)

// Create implements SubCategoryService.
func (s *subCategory) Create(ctx context.Context, in models.CreateSubCategory) (out *models.SubCategory, err error) {
	defer func() {
		logService(ctx, err)
	}()

	// check category
	cat, err := s.srv.mySqlRepo.GetCategoryRepository().GetByCode(ctx, in.CategoryCode)
	if err != nil {
		return
	}
	if cat == nil {
		err = models.GetErrMap(models.ErrKeyCategoryCodeNotFound)
		return
	}

	// check subCategory
	subCat, err := s.srv.mySqlRepo.GetSubCategoryRepository().GetByCode(ctx, in.Code)
	if err != nil {
		return
	}
	if subCat != nil {
		err = models.GetErrMap(models.ErrKeySubCategoryCodeIsExist)
		return
	}

	// check productTypeCode
	if in.ProductTypeCode != "" {
		productType, err := s.srv.mySqlRepo.GetProductTypeRepository().GetByCode(ctx, in.ProductTypeCode)
		if err != nil {
			return nil, err
		}
		if productType == nil {
			err = models.GetErrMap(models.ErrKeyProductTypeNotFound)
			return nil, err
		}
	}

	// check currency
	if in.Currency != "" {
		_, err := s.srv.goDBLedger.GetCurrency(ctx, in.Currency)
		if err != nil {
			err = checkDatabaseError(err, models.ErrKeyCurrencyNotFound)
			return nil, err
		}
	}

	if err = s.srv.mySqlRepo.GetSubCategoryRepository().Create(ctx, &in); err != nil {
		return
	}

	out = &models.SubCategory{
		CategoryCode:    in.CategoryCode,
		Code:            in.Code,
		Name:            in.Name,
		Description:     in.Description,
		AccountType:     in.AccountType,
		ProductTypeCode: in.ProductTypeCode,
		Currency:        in.Currency,
		Status:          in.Status,
	}

	return
}

// GetAll implements SubCategoryService.
func (s *subCategory) GetAll(ctx context.Context, param models.GetAllSubCategoryParam) (out *[]models.SubCategory, err error) {
	defer func() {
		logService(ctx, err)
	}()

	out, err = s.srv.mySqlRepo.GetSubCategoryRepository().GetAll(ctx, param)
	if err != nil {
		return
	}

	return
}

func (s *subCategory) GetByAccountType(ctx context.Context, accountType string) (out *models.SubCategory, err error) {
	defer func() {
		logService(ctx, err)
	}()

	out, err = s.srv.mySqlRepo.GetSubCategoryRepository().GetByAccountType(ctx, accountType)
	if err != nil {
		err = checkDatabaseError(err, models.ErrKeyAccountTypeNotValid)
		return nil, err
	}

	return
}

func (s *subCategory) Update(ctx context.Context, in models.UpdateSubCategory) (out models.UpdateSubCategory, err error) {
	defer func() {
		logService(ctx, err)
	}()

	subCat, err := s.srv.mySqlRepo.GetSubCategoryRepository().GetByCode(ctx, in.Code)
	if err != nil {
		return
	}
	if subCat == nil {
		err = models.GetErrMap(models.ErrKeySubCategoryCodeNotFound)
		return
	}

	//skip and return success if all fields are empty
	if in.Name == "" && in.Description == "" && in.Status == "" && in.ProductTypeCode == nil && in.Currency == nil {
		return in, nil
	}

	// because we remove validation from json, we need to check productTypeCode and currency on DB
	var productTypeName *string
	if in.ProductTypeCode != nil {
		productTypeName = new(string)
		if *in.ProductTypeCode != "" {
			productTypeData, productTypeDataErr := s.srv.mySqlRepo.GetProductTypeRepository().GetByCode(ctx, *in.ProductTypeCode)
			if productTypeDataErr != nil {
				return in, productTypeDataErr
			}
			if productTypeData == nil {
				err = models.GetErrMap(models.ErrKeyProductTypeNotFound)
				return in, err
			}
			productTypeName = &productTypeData.Name
		}
	}

	if in.Currency != nil && *in.Currency != "" {
		_, err = s.srv.goDBLedger.GetCurrency(ctx, *in.Currency)
		if err != nil {
			err = checkDatabaseError(err, models.ErrKeyCurrencyNotFound)
			return in, err
		}
	}

	if err = s.srv.mySqlRepo.Atomic(ctx, func(actx context.Context, r mysql.SQLRepository) (err error) {
		if err = r.GetSubCategoryRepository().Update(actx, in); err != nil {
			err = checkDatabaseError(err)
			return
		}

		// assume that one of these 2 variables are not nil
		if in.ProductTypeCode != nil || in.Currency != nil {
			if err = r.GetAccountRepository().UpdateBySubCategory(actx, models.UpdateBySubCategory{
				Code:            in.Code,
				ProductTypeCode: in.ProductTypeCode,
				Currency:        in.Currency,
			}); err != nil {
				err = checkDatabaseError(err)
				return
			}

			err = s.srv.fpTransactionClient.UpdateAccountBySubCategory(actx, gofptransaction.UpdateBySubCategory{
				Code:            in.Code,
				ProductTypeName: productTypeName,
				Currency:        in.Currency,
			})
		}

		return
	}); err != nil {
		return
	}

	return in, nil
}
