package services

import (
	"context"

	"bitbucket.org/Amartha/go-accounting/internal/models"
)

func (as *account) GetAllCategoryCodeSeq(ctx context.Context) ([]models.DoGetAllCategoryCodeSeqResponse, error) {
	var err error
	defer func() {
		logService(ctx, err)
	}()

	key := "category_code_*"
	result, err := as.srv.cacheRepo.GetAll(ctx, key)
	if err != nil {
		return nil, models.GetErrMap(models.ErrKeyFailedGetFromCache, err.Error())
	}

	res := make([]models.DoGetAllCategoryCodeSeqResponse, 0, len(result))
	for k, v := range result {
		res = append(res, models.DoGetAllCategoryCodeSeqResponse{
			Kind:  models.KindAccount,
			Key:   k,
			Value: v,
		})
	}
	return res, nil
}

func (as *account) UpdateCategoryCodeSeq(ctx context.Context, in models.DoUpdateCategoryCodeSeqRequest) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	_, err = as.srv.cacheRepo.Get(ctx, in.Key)
	if err != nil {
		return err
	}

	if err = as.srv.cacheRepo.Set(ctx, in.Key, in.Value, 0); err != nil {
		err = models.GetErrMap(models.ErrKeyFailedSetToCache, err.Error())
		return
	}

	return
}

func (as *account) CreateCategoryCodeSeq(ctx context.Context, in models.DoCreateCategoryCodeSeqRequest) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	isExist, err := as.srv.cacheRepo.SetIfNotExists(ctx, in.Key, in.Value, 0)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyFailedSetToCache, err.Error())
		return
	}
	if !isExist {
		err = models.GetErrMap(models.ErrKeyDataIsExist)
		return
	}

	return
}
