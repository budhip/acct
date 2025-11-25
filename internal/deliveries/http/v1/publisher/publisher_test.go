package account

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Handler_publish(t *testing.T) {
	testHelper := publisherTestHelper(t)

	type args struct {
		ctx         context.Context
		contentType string
		fileName    string
		topic       string
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
			name: "success case",
			expectation: expectation{
				wantRes:  `"processing"`,
				wantCode: 200,
			},
			args: args{
				ctx:      context.TODO(),
				topic:    "account",
				fileName: "../../../../../storages/test_publish.json",
			},
			doMock: func(args args) {
				testHelper.mockPublisherService.EXPECT().PublishMessage(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "error case - file not csv",
			args: args{
				ctx:      context.TODO(),
				topic:    "account",
				fileName: "publisher.go",
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"file not csv"}`,
				wantCode: 400,
			},
		},
		{
			name: "error case - contentType not allowed",
			args: args{
				ctx:         context.TODO(),
				contentType: echo.MIMEApplicationJSON,
				topic:       "account",
				fileName:    "../../../../../storages/test_publish.json",
			},
			expectation: expectation{
				wantRes:  `{"status":"error","code":400,"message":"request Content-Type isn't multipart/form-data"}`,
				wantCode: 400,
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

			f, err := os.Open(tt.args.fileName)
			require.NoError(t, err)
			_, err = io.Copy(dataPart, f)
			require.NoError(t, err)
			require.NoError(t, writer.Close())

			r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/topics/%s", tt.args.topic), body)
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
