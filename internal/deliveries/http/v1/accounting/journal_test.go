package accounting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Handler_createJournal(t *testing.T) {
	testHelper := accountingTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         *models.JournalRequest
	}
	trx := make([]models.Transaction, 0)
	trx = append(trx, models.Transaction{
		TransactionType:     "DSBAB",
		TransactionTypeName: "DSBAB",
		Account:             "11100100000001",
		Narrative:           "Debit",
		Amount: decimal.NewFromInt(
			int64(5000001),
		),
		IsDebit: true,
	},
		models.Transaction{
			TransactionType:     "DSBAB",
			TransactionTypeName: "DSBAB",
			Account:             "11200100000006",
			Narrative:           "Credit",
			Amount: decimal.NewFromInt(
				int64(5000001),
			),
			IsDebit: false,
		},
	)
	mockCreateJournal := models.JournalRequest{
		ReferenceNumber: "123456",
		TransactionId:   "6de11650-dbee-4f67-9ade-ececc7a02571900",
		OrderType:       "DSB",
		TransactionDate: "2023-12-29 00:00:00",
		ProcessingDate:  "2023-12-29 00:00:00",
		Currency:        "IDR",
		Transactions:    trx,
		Metadata: &models.Metadata{
			"description": "BROILERX_LOAN",
		},
	}

	mockCreateJournalNotComplete := models.JournalRequest{
		ReferenceNumber: "",
		TransactionId:   "6de11650-dbee-4f67-9ade-ececc7a02571900",
		OrderType:       "DSB",
		TransactionDate: "2023-12-29 00:00:00",
		ProcessingDate:  "2023-12-29 00:00:00",
		Currency:        "IDR",
		Transactions:    trx,
		Metadata: &models.Metadata{
			"description": "BROILERX_LOAN",
		},
	}
	type expectation struct {
		wantRes  string
		wantCode int
	}
	tests := []struct {
		name        string
		args        args
		expectation expectation
		doMock      func(args args, expectation expectation)
	}{
		{
			name: "success",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         &mockCreateJournal,
			},
			expectation: expectation{
				wantRes:  `{"kind":"journal","referenceNumber":"123456","transactionId":"6de11650-dbee-4f67-9ade-ececc7a02571900","orderType":"DSB","transactionDate":"2023-12-29 00:00:00","processingDate":"2023-12-29 00:00:00","currency":"IDR","transactions":[{"transactionType":"DSBAB","transactionTypeName":"DSBAB","account":"11100100000001","narrative":"Debit","amount":"5000001","isDebit":true},{"transactionType":"DSBAB","transactionTypeName":"DSBAB","account":"11200100000006","narrative":"Credit","amount":"5000001","isDebit":false}],"metadata":{"description":"BROILERX_LOAN"}}`,
				wantCode: 201,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockJournalService.EXPECT().InsertJournalTransaction(args.ctx, gomock.Any()).Return(
					[]models.JournalEntryCreatedRequest{}, nil,
				)
			},
		},
		{
			name: "error contentType",
			args: args{
				ctx:         context.Background(),
				contentType: "",
				req:         &mockCreateJournal,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"code=415, message=Unsupported Media Type"}`,
				wantCode: 400,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name: "error case - data is exist",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         &mockCreateJournal,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_IS_EXIST","message":"data is exist"}`,
				wantCode: 409,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockJournalService.EXPECT().InsertJournalTransaction(args.ctx, gomock.Any()).Return([]models.JournalEntryCreatedRequest{}, models.GetErrMap(models.ErrKeyDataIsExist))
			},
		},
		{
			name: "error case - data not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         &mockCreateJournal,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"account number not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockJournalService.EXPECT().InsertJournalTransaction(args.ctx, gomock.Any()).Return([]models.JournalEntryCreatedRequest{}, models.GetErrMap(models.ErrKeyAccountNumberNotFound))
			},
		},
		{
			name: "error - Internal server error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         &mockCreateJournal,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATABASE_ERROR","message":"database error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockJournalService.EXPECT().InsertJournalTransaction(args.ctx, gomock.Any()).Return([]models.JournalEntryCreatedRequest{}, models.GetErrMap(models.ErrKeyDatabaseError))
			},
		},
		{
			name: "error - validation",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         &mockCreateJournalNotComplete,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"referenceNumber","message":"field is missing"}]}`,
				wantCode: 422,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockJournalService.EXPECT().InsertJournalTransaction(args.ctx, gomock.Any()).Return(
					[]models.JournalEntryCreatedRequest{}, nil,
				)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.expectation)
			}
			var b bytes.Buffer
			err := json.NewEncoder(&b).Encode(tt.args.req)
			require.NoError(t, err)
			r := httptest.NewRequest(http.MethodPost, "/api/v1/journals", &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()
			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)
			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_publishJournal(t *testing.T) {
	testHelper := accountingTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		req         *models.JournalRequest
	}
	trx := make([]models.Transaction, 0)
	trx = append(trx, models.Transaction{
		TransactionType:     "DSBAB",
		TransactionTypeName: "DSBAB",
		Account:             "11100100000001",
		Narrative:           "Debit",
		Amount: decimal.NewFromInt(
			int64(5000001),
		),
		IsDebit: true,
	},
		models.Transaction{
			TransactionType:     "DSBAB",
			TransactionTypeName: "DSBAB",
			Account:             "11200100000006",
			Narrative:           "Credit",
			Amount: decimal.NewFromInt(
				int64(5000001),
			),
			IsDebit: false,
		},
	)
	mockCreateJournal := models.JournalRequest{
		ReferenceNumber: "123456",
		TransactionId:   "6de11650-dbee-4f67-9ade-ececc7a02571900",
		OrderType:       "DSB",
		TransactionDate: "2023-12-29 00:00:00",
		ProcessingDate:  "2023-12-29 00:00:00",
		Currency:        "IDR",
		Transactions:    trx,
		Metadata: &models.Metadata{
			"description": "BROILERX_LOAN",
		},
	}

	mockCreateJournalNotComplete := models.JournalRequest{
		ReferenceNumber: "",
		TransactionId:   "6de11650-dbee-4f67-9ade-ececc7a02571900",
		OrderType:       "DSB",
		TransactionDate: "2023-12-29 00:00:00",
		ProcessingDate:  "2023-12-29 00:00:00",
		Currency:        "IDR",
		Transactions:    trx,
		Metadata: &models.Metadata{
			"description": "BROILERX_LOAN",
		},
	}
	type expectation struct {
		wantRes  string
		wantCode int
	}
	tests := []struct {
		name        string
		args        args
		expectation expectation
		doMock      func(args args, expectation expectation)
	}{
		{
			name: "success",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         &mockCreateJournal,
			},
			expectation: expectation{
				wantRes:  `{"kind":"journal","referenceNumber":"123456","transactionId":"6de11650-dbee-4f67-9ade-ececc7a02571900","orderType":"DSB","transactionDate":"2023-12-29 00:00:00","processingDate":"2023-12-29 00:00:00","currency":"IDR","transactions":[{"transactionType":"DSBAB","transactionTypeName":"DSBAB","account":"11100100000001","narrative":"Debit","amount":"5000001","isDebit":true},{"transactionType":"DSBAB","transactionTypeName":"DSBAB","account":"11200100000006","narrative":"Credit","amount":"5000001","isDebit":false}],"metadata":{"description":"BROILERX_LOAN"}}`,
				wantCode: 201,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockJournalService.EXPECT().PublishJournalTransaction(args.ctx, gomock.Any()).Return(
					nil,
				)
			},
		},
		{
			name: "error contentType",
			args: args{
				ctx:         context.Background(),
				contentType: "",
				req:         &mockCreateJournal,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"code=415, message=Unsupported Media Type"}`,
				wantCode: 400,
			},
			doMock: func(args args, expectation expectation) {},
		},
		{
			name: "error case - data is exist",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         &mockCreateJournal,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_IS_EXIST","message":"data is exist"}`,
				wantCode: 409,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockJournalService.EXPECT().PublishJournalTransaction(args.ctx, gomock.Any()).Return(models.GetErrMap(models.ErrKeyDataIsExist))
			},
		},
		{
			name: "error case - data not found",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         &mockCreateJournal,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"account number not found"}`,
				wantCode: 404,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockJournalService.EXPECT().PublishJournalTransaction(args.ctx, gomock.Any()).Return(models.GetErrMap(models.ErrKeyAccountNumberNotFound))
			},
		},
		{
			name: "error - Internal server error",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         &mockCreateJournal,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATABASE_ERROR","message":"database error"}`,
				wantCode: 500,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockJournalService.EXPECT().PublishJournalTransaction(args.ctx, gomock.Any()).Return(models.GetErrMap(models.ErrKeyDatabaseError))
			},
		},
		{
			name: "error - validation",
			args: args{
				ctx:         context.Background(),
				contentType: echo.MIMEApplicationJSON,
				req:         &mockCreateJournalNotComplete,
			},
			expectation: expectation{
				wantRes:  `{"status":"error","message":"validation failed","errors":[{"code":"MISSING_FIELD","field":"referenceNumber","message":"field is missing"}]}`,
				wantCode: 422,
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockJournalService.EXPECT().PublishJournalTransaction(args.ctx, gomock.Any()).Return(
					nil,
				)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.expectation)
			}
			var b bytes.Buffer
			err := json.NewEncoder(&b).Encode(tt.args.req)
			require.NoError(t, err)
			r := httptest.NewRequest(http.MethodPost, "/api/v1/journals/publish", &b)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()
			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)
			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_uploadJournal(t *testing.T) {
	testHelper := accountingTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		fileName    string
	}
	type expectation struct {
		wantRes  string
		wantCode int
	}
	tests := []struct {
		name        string
		args        args
		expectation expectation
		doMock      func(args args)
	}{
		{
			name: "happy path",
			expectation: expectation{
				wantRes:  `"processing"`,
				wantCode: 200,
			},
			args: args{
				ctx:      context.TODO(),
				fileName: "../../../../../storages/upload_journals.csv",
			},
			doMock: func(args args) {
				testHelper.mockJournalService.EXPECT().ProcessUploadJournal(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "error contentType not allowed",
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"request Content-Type isn't multipart/form-data"}`,
				wantCode: 400,
			},
			args: args{
				ctx:         context.TODO(),
				contentType: echo.MIMEApplicationJSON,
				fileName:    "../../../../../storages/upload_journals.csv",
			},
			doMock: func(args args) {
				testHelper.mockJournalService.EXPECT().ProcessUploadJournal(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "error file not csv",
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"file not csv"}`,
				wantCode: 400,
			},
			args: args{
				ctx:      context.TODO(),
				fileName: "journal.go",
			},
			doMock: func(args args) {
				testHelper.mockJournalService.EXPECT().ProcessUploadJournal(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}

			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)
			dataPart, err := writer.CreateFormFile("file", tt.args.fileName)
			require.NoError(t, err)

			// copy file content into multipart section dataPart
			f, err := os.Open(tt.args.fileName)
			require.NoError(t, err)
			_, err = io.Copy(dataPart, f)
			require.NoError(t, err)
			require.NoError(t, writer.Close())

			r := httptest.NewRequest(http.MethodPost, "/api/v1/journals/upload", body)
			if tt.args.contentType == "" {
				r.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
			}
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}

func Test_Handler_getByTransactionId(t *testing.T) {
	testHelper := accountingTestHelper(t)
	result := []models.GetJournalDetailOut{
		{
			TransactionId:   "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			JournalId:       "2024060601093678",
			AccountNumber:   "214003000000194",
			AccountName:     "Mails Morales",
			AltId:           "Mails002",
			EntityCode:      "003",
			EntityName:      "AFA",
			SubCategoryCode: "21401",
			SubCategoryName: "eWallet User",
			TransactionType: "PAYGL",
			Amount:          decimal.NewFromFloat(10000),
			Narrative:       "Repayment Group Loan via Poket jindankarasuno",
			// TransactionDate: now,
			IsDebit: true,
		},
		{

			TransactionId:   "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			JournalId:       "2024060601093679",
			AccountNumber:   "261003000000005",
			AccountName:     "IA Ewallet Repayment Group Loan",
			AltId:           "",
			EntityCode:      "003",
			EntityName:      "AFA",
			SubCategoryCode: "26101",
			SubCategoryName: "Related Party Payable",
			Amount:          decimal.NewFromFloat(10000),
			TransactionType: "PAYGL",
			// TransactionDate: now,
			Narrative: "Repayment Group Loan via Poket jindankarasuno",
			IsDebit:   false,
		},
		{
			TransactionId:   "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			JournalId:       "2024060601093680",
			AccountNumber:   "143001000000168",
			AccountName:     "IA Ewallet Repayment Group Loan",
			AltId:           "",
			EntityCode:      "001",
			EntityName:      "AMF",
			SubCategoryCode: "14301",
			SubCategoryName: "Related Parties Receivable",
			TransactionType: "PAYGL",
			Amount:          decimal.NewFromFloat(10000),
			// TransactionDate: now,
			Narrative: "Repayment Group Loan via Poket jindankarasuno",
			IsDebit:   true,
		},
		{
			TransactionId:   "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			JournalId:       "2024060601093681",
			AccountNumber:   "121001000000008",
			AccountName:     "Cash in Transit - Repayment Group Loan",
			AltId:           "",
			EntityCode:      "001",
			EntityName:      "AMF",
			SubCategoryCode: "12101",
			SubCategoryName: "Cash in Transit",
			TransactionType: "PAYGL",
			Amount:          decimal.NewFromFloat(10000),
			// TransactionDate: now,
			Narrative: "Repayment Group Loan via Poket jindankarasuno",
			IsDebit:   false,
		},
	}
	type args struct {
		contentType string
		req         string
	}
	type expectation struct {
		wantRes  string
		wantCode int
	}
	tests := []struct {
		name        string
		args        args
		expectation expectation
		doMock      func(args args, expectation expectation)
	}{
		{
			name: "success case",
			args: args{
				contentType: echo.MIMEApplicationJSON,
				req:         "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockJournalService.EXPECT().GetJournalByTransactionId(gomock.AssignableToTypeOf(context.Background()), args.req).Return(result, nil)
			},
			expectation: expectation{
				wantRes:  `{"kind":"collection","contents":[{"kind":"journal","transactionId":"12c5692e-cfbd-4cee-a4ce-86eac1447d48","journalId":"2024060601093678","accountNumber":"214003000000194","accountName":"Mails Morales","altId":"Mails002","entityCode":"003","entityName":"AFA","subCategoryCode":"21401","subCategoryName":"eWallet User","transactionType":"PAYGL","amount":"10.000,00","transactionDate":"0001-01-01 07:07:12","narrative":"Repayment Group Loan via Poket jindankarasuno","isDebit":true},{"kind":"journal","transactionId":"12c5692e-cfbd-4cee-a4ce-86eac1447d48","journalId":"2024060601093679","accountNumber":"261003000000005","accountName":"IA Ewallet Repayment Group Loan","altId":"","entityCode":"003","entityName":"AFA","subCategoryCode":"26101","subCategoryName":"Related Party Payable","transactionType":"PAYGL","amount":"10.000,00","transactionDate":"0001-01-01 07:07:12","narrative":"Repayment Group Loan via Poket jindankarasuno","isDebit":false},{"kind":"journal","transactionId":"12c5692e-cfbd-4cee-a4ce-86eac1447d48","journalId":"2024060601093680","accountNumber":"143001000000168","accountName":"IA Ewallet Repayment Group Loan","altId":"","entityCode":"001","entityName":"AMF","subCategoryCode":"14301","subCategoryName":"Related Parties Receivable","transactionType":"PAYGL","amount":"10.000,00","transactionDate":"0001-01-01 07:07:12","narrative":"Repayment Group Loan via Poket jindankarasuno","isDebit":true},{"kind":"journal","transactionId":"12c5692e-cfbd-4cee-a4ce-86eac1447d48","journalId":"2024060601093681","accountNumber":"121001000000008","accountName":"Cash in Transit - Repayment Group Loan","altId":"","entityCode":"001","entityName":"AMF","subCategoryCode":"12101","subCategoryName":"Cash in Transit","transactionType":"PAYGL","amount":"10.000,00","transactionDate":"0001-01-01 07:07:12","narrative":"Repayment Group Loan via Poket jindankarasuno","isDebit":false}],"total_rows":4}`,
				wantCode: 200,
			},
		},
		{
			name: "error case - transaction id not found",
			args: args{
				contentType: echo.MIMEApplicationJSON,
				req:         "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockJournalService.EXPECT().GetJournalByTransactionId(gomock.AssignableToTypeOf(context.Background()), args.req).Return(result, models.GetErrMap(models.ErrKeyTransactionIdNotFound))
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATA_NOT_FOUND","message":"transaction id not found"}`,
				wantCode: 404,
			},
		},
		{
			name: "error case - database error",
			args: args{
				contentType: echo.MIMEApplicationJSON,
				req:         "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			},
			doMock: func(args args, expectation expectation) {
				testHelper.mockJournalService.EXPECT().GetJournalByTransactionId(gomock.AssignableToTypeOf(context.Background()), args.req).Return(result, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":"DATABASE_ERROR","message":"database error"}`,
				wantCode: 500,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.expectation)
			}

			var b bytes.Buffer
			err := json.NewEncoder(&b).Encode(tt.args.req)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/journals/%s", tt.args.req), nil)
			r.Header.Set(echo.HeaderContentType, tt.args.contentType)
			w := httptest.NewRecorder()

			testHelper.router.NewContext(r, w)
			testHelper.router.ServeHTTP(w, r)

			require.Equal(t, tt.expectation.wantCode, w.Code)
			require.Equal(t, tt.expectation.wantRes, strings.Trim(w.Body.String(), "\n"))
		})
	}
}
