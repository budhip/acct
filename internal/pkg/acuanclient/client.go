package acuanclient

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-accounting/internal/config"

	goAcuanLib "bitbucket.org/Amartha/go-acuan-lib"
	goAcuanLibModel "bitbucket.org/Amartha/go-acuan-lib/model"
	xlog "bitbucket.org/Amartha/go-x/log"
)

type AcuanClient interface {
	PublishAccount(ctx context.Context, data PublishAccountData)
}

type client struct {
	acuanClient *goAcuanLib.AcuanLib
}

func New(cfg *config.Configuration) (AcuanClient, error) {
	acuanConfig := &goAcuanLib.Config{
		Kafka: &goAcuanLib.KafkaConfig{
			BrokerList:        cfg.AcuanLibConfig.Kafka.BrokerList,
			PartitionStrategy: cfg.AcuanLibConfig.Kafka.PartitionStrategy,
		},
		SourceSystem:          cfg.AcuanLibConfig.SourceSystem,
		Topic:                 cfg.AcuanLibConfig.Topic,
		TopicAccounting:       cfg.AcuanLibConfig.TopicAccounting,
		TopUpKey:              cfg.AcuanLibConfig.TopUpKey,
		InvestmentKey:         cfg.AcuanLibConfig.InvestmentKey,
		CashoutKey:            cfg.AcuanLibConfig.CashoutKey,
		DisbursementKey:       cfg.AcuanLibConfig.DisbursementKey,
		DisbursementFailedKey: cfg.AcuanLibConfig.DisbursementFailedKey,
		RepaymentKey:          cfg.AcuanLibConfig.RepaymentKey,
		RefundKey:             cfg.AcuanLibConfig.RefundKey,
	}

	if acuanConfig.SourceSystem == "" {
		acuanConfig.SourceSystem = cfg.App.Name
	}

	acuanClient, err := goAcuanLib.NewClient(acuanConfig)
	if err != nil {
		return nil, fmt.Errorf("failed connect to acuan client: %v", err)
	}

	return &client{acuanClient}, nil
}

func (c *client) PublishAccount(ctx context.Context, data PublishAccountData) {
	var legacyId *goAcuanLibModel.AccountLegacyId

	if data.LegacyId != nil {
		acuanLegacy := goAcuanLibModel.AccountLegacyId(*data.LegacyId)
		legacyId = &acuanLegacy
	}

	err := c.acuanClient.Accounting.Publish(goAcuanLibModel.Account{
		Type:            data.Type,
		AccountNumber:   data.AccountNumber,
		Name:            data.Name,
		ProductTypeName: data.ProductTypeName,
		OwnerId:         data.OwnerId,
		CategoryCode:    data.CategoryCode,
		SubCategoryCode: data.SubCategoryCode,
		EntityCode:      data.EntityCode,
		Currency:        data.Currency,
		AltId:           data.AltId,
		Status:          data.Status,
		LegacyId:        legacyId,
		Metadata:        data.Metadata,
	})
	if err != nil {
		xlog.Error(ctx, "[PUBLISH-ACCOUNT]", xlog.String("status", "fail"), xlog.Any("message", data), xlog.Err(err))
		return
	}
	xlog.Info(ctx, "[PUBLISH-ACCOUNT]", xlog.String("status", "success"), xlog.Any("message", data))
}
