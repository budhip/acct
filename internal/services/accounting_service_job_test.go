package services_test

import (
	"bitbucket.org/Amartha/go-accounting/internal/models"
)

func dummyResponseGetAllCategorySubCategoryCOAType() (map[string][]models.CategorySubCategoryCOAType, map[string][]models.CategorySubCategoryCOAType, map[string]models.CategorySubCategoryCOAType) {
	result := []models.CategorySubCategoryCOAType{
		{
			CoaTypeCode:     "AST",
			CoaTypeName:     "Asset",
			CategoryCode:    "114",
			CategoryName:    "Bank HO",
			SubCategoryCode: "11403",
			SubCategoryName: "Bank BRI",
		},
		{
			CoaTypeCode:     "LIA",
			CoaTypeName:     "Liability",
			CategoryCode:    "211",
			CategoryName:    "Marketplace Payable (Lender's Cash)",
			SubCategoryCode: "21108",
			SubCategoryName: "Lender's Cash - Earn (Fixed Income)",
		},
	}

	mapCoaType := make(map[string][]models.CategorySubCategoryCOAType)
	mapCateg := make(map[string][]models.CategorySubCategoryCOAType)
	mapSubCateg := make(map[string]models.CategorySubCategoryCOAType)

	for _, v := range result {
		mapCoaType[v.CoaTypeCode] = append(mapCoaType[v.CoaTypeCode], v)
		mapCateg[v.CategoryCode] = append(mapCoaType[v.CategoryCode], v)
		mapSubCateg[v.SubCategoryCode] = v
	}
	return mapCoaType, mapCateg, mapSubCateg
}

// func Test_accounting_GenerateAccountTrialBalanceDaily(t *testing.T) {
// 	testHelper := serviceTestHelper(t)

// 	date := atime.Now()
// 	entities := []models.Entity{
// 		{
// 			Code:        "001",
// 			Name:        "AMF",
// 			Description: "PT. Amartha Mikro Fintek",
// 			Status:      "active",
// 			CreatedAt:   &date,
// 			UpdatedAt:   &date,
// 		},
// 	}
// 	subCategories := []models.SubCategory{
// 		{

// 			CategoryCode:    "111",
// 			Code:            "11403",
// 			Name:            "Bank BRI",
// 			Description:     "Bank BRI",
// 			AccountType:     "BANK_HO_BRI",
// 			ProductTypeCode: "",
// 			ProductTypeName: "",
// 			Currency:        "",
// 			Status:          "active",
// 			CreatedAt:       &date,
// 			UpdatedAt:       &date,
// 		},
// 	}
// 	mapCoaType, mapCateg, mapSubCateg := dummyResponseGetAllCategorySubCategoryCOAType()
// 	databaseError := models.GetErrMap(models.ErrKeyDatabaseError)

// 	type args struct {
// 		req time.Time
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		doMock  func(args args)
// 		wantErr bool
// 	}{
// 		{
// 			name: "success case - generate account trial balance daily",
// 			args: args{
// 				req: date,
// 			},
// 			doMock: func(args args) {
// 				testHelper.mockEntityRepository.EXPECT().List(gomock.Any()).Return(&entities, nil)
// 				testHelper.mockSubCategoryRepository.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return(&subCategories, nil)
// 				testHelper.mockMySQLRepository.EXPECT().GetAllCategorySubCategoryCOAType(gomock.Any()).Return(mapCoaType, mapCateg, mapSubCateg, nil)

// 				testHelper.mockAcctRepository.EXPECT().CalculateFromAccountBalanceDaily(gomock.Any(), gomock.Any()).Return(
// 					models.AccountTrialBalance{}, models.ErrNoRows).MaxTimes(1)
// 				testHelper.mockAcctRepository.EXPECT().CalculateFromTransactions(gomock.Any(), gomock.Any()).Return(
// 					models.AccountTrialBalance{}, nil).MaxTimes(1)
// 				testHelper.mockAcctRepository.EXPECT().GetOpeningBalanceFromAccountTrialBalance(gomock.Any(), gomock.Any()).Return(
// 					decimal.Decimal{}, models.ErrNoRows).MaxTimes(1)
// 				testHelper.mockAcctRepository.EXPECT().CalculateOpeningClosingBalanceFromAccountBalance(gomock.Any(), gomock.Any()).Return(
// 					decimal.Decimal{}, nil).MaxTimes(1)
// 				testHelper.mockAcctRepository.EXPECT().InsertAccountTrialBalance(gomock.Any(), gomock.Any()).Return(nil)
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "success case - generate account trial balance daily - no data",
// 			args: args{
// 				req: date,
// 			},
// 			doMock: func(args args) {
// 				testHelper.mockEntityRepository.EXPECT().List(gomock.Any()).Return(&entities, nil)
// 				testHelper.mockSubCategoryRepository.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return(&subCategories, nil)
// 				testHelper.mockMySQLRepository.EXPECT().GetAllCategorySubCategoryCOAType(gomock.Any()).Return(mapCoaType, mapCateg, mapSubCateg, nil)

// 				testHelper.mockAcctRepository.EXPECT().CalculateFromAccountBalanceDaily(gomock.Any(), gomock.Any()).Return(
// 					models.AccountTrialBalance{}, models.ErrNoRows).MaxTimes(1)
// 				testHelper.mockAcctRepository.EXPECT().CalculateFromTransactions(gomock.Any(), gomock.Any()).Return(
// 					models.AccountTrialBalance{}, models.ErrNoRows).MaxTimes(1)
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "error case - database error get all entity",
// 			args: args{
// 				req: date,
// 			},
// 			doMock: func(args args) {
// 				testHelper.mockEntityRepository.EXPECT().List(gomock.Any()).Return(&entities, databaseError)
// 				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(gomock.Any(), gomock.Any())
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "error case - database error get all sub category",
// 			args: args{
// 				req: date,
// 			},
// 			doMock: func(args args) {
// 				testHelper.mockEntityRepository.EXPECT().List(gomock.Any()).Return(&entities, nil)
// 				testHelper.mockSubCategoryRepository.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return(&subCategories, databaseError)
// 				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(gomock.Any(), gomock.Any())
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "error case - database error GetAllCategorySubCategoryCOAType",
// 			args: args{
// 				req: date,
// 			},
// 			doMock: func(args args) {
// 				testHelper.mockEntityRepository.EXPECT().List(gomock.Any()).Return(&entities, nil)
// 				testHelper.mockSubCategoryRepository.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return(&subCategories, nil)
// 				testHelper.mockMySQLRepository.EXPECT().GetAllCategorySubCategoryCOAType(gomock.Any()).Return(mapCoaType, mapCateg, mapSubCateg, databaseError)
// 				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(gomock.Any(), gomock.Any())
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "error case - database error CalculateFromAccountBalanceDaily",
// 			args: args{
// 				req: date,
// 			},
// 			doMock: func(args args) {
// 				testHelper.mockEntityRepository.EXPECT().List(gomock.Any()).Return(&entities, nil)
// 				testHelper.mockSubCategoryRepository.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return(&subCategories, nil)
// 				testHelper.mockMySQLRepository.EXPECT().GetAllCategorySubCategoryCOAType(gomock.Any()).Return(mapCoaType, mapCateg, mapSubCateg, nil)

// 				testHelper.mockAcctRepository.EXPECT().CalculateFromAccountBalanceDaily(gomock.Any(), gomock.Any()).Return(
// 					models.AccountTrialBalance{}, databaseError).MaxTimes(1)

// 				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(gomock.Any(), gomock.Any())
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "error case - database error CalculateFromTransactions",
// 			args: args{
// 				req: date,
// 			},
// 			doMock: func(args args) {
// 				testHelper.mockEntityRepository.EXPECT().List(gomock.Any()).Return(&entities, nil)
// 				testHelper.mockSubCategoryRepository.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return(&subCategories, nil)
// 				testHelper.mockMySQLRepository.EXPECT().GetAllCategorySubCategoryCOAType(gomock.Any()).Return(mapCoaType, mapCateg, mapSubCateg, nil)

// 				testHelper.mockAcctRepository.EXPECT().CalculateFromAccountBalanceDaily(gomock.Any(), gomock.Any()).Return(
// 					models.AccountTrialBalance{}, models.ErrNoRows).MaxTimes(1)
// 				testHelper.mockAcctRepository.EXPECT().CalculateFromTransactions(gomock.Any(), gomock.Any()).Return(
// 					models.AccountTrialBalance{}, databaseError)

// 				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(gomock.Any(), gomock.Any())
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "error case - database error GetOpeningBalanceFromAccountTrialBalance",
// 			args: args{
// 				req: date,
// 			},
// 			doMock: func(args args) {
// 				testHelper.mockEntityRepository.EXPECT().List(gomock.Any()).Return(&entities, nil).MaxTimes(1)
// 				testHelper.mockSubCategoryRepository.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return(&subCategories, nil).MaxTimes(1)
// 				testHelper.mockMySQLRepository.EXPECT().GetAllCategorySubCategoryCOAType(gomock.Any()).Return(mapCoaType, mapCateg, mapSubCateg, nil).MaxTimes(1)

// 				testHelper.mockAcctRepository.EXPECT().CalculateFromAccountBalanceDaily(gomock.Any(), gomock.Any()).Return(
// 					models.AccountTrialBalance{}, models.ErrNoRows).MaxTimes(1)
// 				testHelper.mockAcctRepository.EXPECT().CalculateFromTransactions(gomock.Any(), gomock.Any()).Return(
// 					models.AccountTrialBalance{}, nil).MaxTimes(1)
// 				testHelper.mockAcctRepository.EXPECT().GetOpeningBalanceFromAccountTrialBalance(gomock.Any(), gomock.Any()).Return(
// 					decimal.Decimal{}, databaseError)

// 				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(gomock.Any(), gomock.Any())
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "error case - database error CalculateOpeningClosingBalanceFromAccountBalance",
// 			args: args{
// 				req: date,
// 			},
// 			doMock: func(args args) {
// 				testHelper.mockEntityRepository.EXPECT().List(gomock.Any()).Return(&entities, nil)
// 				testHelper.mockSubCategoryRepository.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return(&subCategories, nil)
// 				testHelper.mockMySQLRepository.EXPECT().GetAllCategorySubCategoryCOAType(gomock.Any()).Return(mapCoaType, mapCateg, mapSubCateg, nil)

// 				testHelper.mockAcctRepository.EXPECT().CalculateFromAccountBalanceDaily(gomock.Any(), gomock.Any()).Return(
// 					models.AccountTrialBalance{}, models.ErrNoRows).MaxTimes(1)
// 				testHelper.mockAcctRepository.EXPECT().CalculateFromTransactions(gomock.Any(), gomock.Any()).Return(
// 					models.AccountTrialBalance{}, nil).MaxTimes(1)
// 				testHelper.mockAcctRepository.EXPECT().GetOpeningBalanceFromAccountTrialBalance(gomock.Any(), gomock.Any()).Return(
// 					decimal.Decimal{}, models.ErrNoRows).MaxTimes(1)
// 				testHelper.mockAcctRepository.EXPECT().CalculateOpeningClosingBalanceFromAccountBalance(gomock.Any(), gomock.Any()).Return(
// 					decimal.Decimal{}, databaseError).MaxTimes(1)

// 				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(gomock.Any(), gomock.Any()).Return(assert.AnError)
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "error case - database error InsertAccountTrialBalance",
// 			args: args{
// 				req: date,
// 			},
// 			doMock: func(args args) {
// 				testHelper.mockEntityRepository.EXPECT().List(gomock.Any()).Return(&entities, nil)
// 				testHelper.mockSubCategoryRepository.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return(&subCategories, nil)
// 				testHelper.mockMySQLRepository.EXPECT().GetAllCategorySubCategoryCOAType(gomock.Any()).Return(mapCoaType, mapCateg, mapSubCateg, nil)

// 				testHelper.mockAcctRepository.EXPECT().CalculateFromAccountBalanceDaily(gomock.Any(), gomock.Any()).Return(
// 					models.AccountTrialBalance{}, models.ErrNoRows).MaxTimes(1)
// 				testHelper.mockAcctRepository.EXPECT().CalculateFromTransactions(gomock.Any(), gomock.Any()).Return(
// 					models.AccountTrialBalance{}, nil).MaxTimes(1)
// 				testHelper.mockAcctRepository.EXPECT().GetOpeningBalanceFromAccountTrialBalance(gomock.Any(), gomock.Any()).Return(
// 					decimal.Decimal{}, models.ErrNoRows).MaxTimes(1)
// 				testHelper.mockAcctRepository.EXPECT().CalculateOpeningClosingBalanceFromAccountBalance(gomock.Any(), gomock.Any()).Return(
// 					decimal.Decimal{}, nil).MaxTimes(1)
// 				testHelper.mockAcctRepository.EXPECT().InsertAccountTrialBalance(gomock.Any(), gomock.Any()).Return(databaseError)

// 				testHelper.mockDDDNotification.EXPECT().SendMessageToSlack(gomock.Any(), gomock.Any()).Return(assert.AnError)
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.doMock != nil {
// 				tt.doMock(tt.args)
// 			}
// 			err := testHelper.accountingService.GenerateAccountTrialBalanceDaily(context.Background(), tt.args.req)
// 			assert.Equal(t, tt.wantErr, err != nil)
// 		})
// 	}
// }
