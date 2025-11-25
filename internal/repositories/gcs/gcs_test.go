package gcs

import (
	"context"
	"testing"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/config"
	"bitbucket.org/Amartha/go-accounting/internal/models"

	"cloud.google.com/go/storage"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/option"
)

type gcsHelper struct {
	server        *fakestorage.Server
	client        *storage.Client
	defaultConfig *config.CloudStorageConfig
}

func newGcsClientHelper(t *testing.T) *gcsHelper {
	t.Helper()
	t.Parallel()

	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{
		NoListener: true,
	})
	assert.NoError(t, err)

	client, err := storage.NewClient(
		context.Background(),
		option.WithoutAuthentication(),
		option.WithHTTPClient(server.HTTPClient()))
	assert.NoError(t, err)

	return &gcsHelper{
		server: server,
		client: client,
		defaultConfig: &config.CloudStorageConfig{
			BaseURL:    "http://test:1337",
			BucketName: "DUMMY_BUCKET",
		},
	}
}

func TestNewCloudStorageRepository(t *testing.T) {
	helper := newGcsClientHelper(t)

	type args struct {
		cfg  *config.Configuration
		opts []option.ClientOption
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success init cloud storage",
			args: args{
				cfg: &config.Configuration{
					App: config.App{
						Env:  "test",
						Name: "go-accounting[test]",
					},
					CloudStorageConfig: *helper.defaultConfig,
				},
				opts: []option.ClientOption{
					option.WithoutAuthentication(),
					option.WithHTTPClient(helper.server.HTTPClient()),
				},
			},
			wantErr: false,
		},
		{
			name: "failed init cloud storage (bucket name not set)",
			args: args{
				cfg: &config.Configuration{
					App: config.App{
						Env:  "test",
						Name: "go-accounting[test]",
					},
					CloudStorageConfig: config.CloudStorageConfig{
						BaseURL:    "",
						BucketName: "",
					},
				},
				opts: []option.ClientOption{
					option.WithoutAuthentication(),
					option.WithHTTPClient(helper.server.HTTPClient()),
				},
			},
			wantErr: true,
		},
		{
			name: "failed init cloud storage (no option provided for testing)",
			args: args{
				cfg: &config.Configuration{
					App: config.App{
						Env:  "test",
						Name: "go-accounting[test]",
					},
					CloudStorageConfig: *helper.defaultConfig,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCloudStorageRepository(tt.args.cfg, tt.args.opts...)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_cloudStorageClient_Close(t *testing.T) {
	helper := newGcsClientHelper(t)

	type fields struct {
		config *config.CloudStorageConfig
		client *storage.Client
	}
	tests := []struct {
		name    string
		fields  fields
		doMock  func(f fields)
		wantErr bool
	}{
		{
			name: "success close",
			fields: fields{
				config: helper.defaultConfig,
				client: helper.client,
			},
			doMock: func(f fields) {
				helper.server.CreateBucketWithOpts(fakestorage.CreateBucketOpts{
					Name: f.config.BucketName,
				})
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.fields)
			}

			cs := &cloudStorageClient{
				config: tt.fields.config,
				client: tt.fields.client,
			}

			err := cs.Close()
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_cloudStorageClient_NewWriter(t *testing.T) {
	helper := newGcsClientHelper(t)

	type fields struct {
		config *config.CloudStorageConfig
		client *storage.Client
	}
	type args struct {
		ctx     context.Context
		payload *models.CloudStoragePayload
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "success create new writer",
			fields: fields{
				config: helper.defaultConfig,
				client: helper.client,
			},
			args: args{
				ctx: context.TODO(),
				payload: &models.CloudStoragePayload{
					Filename: "my_writer.txt",
					Path:     "my_path_for_writer",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &cloudStorageClient{
				config: tt.fields.config,
				client: tt.fields.client,
			}
			assert.NotNilf(t, cs.NewWriter(tt.args.ctx, tt.args.payload), "NewWriter(%v, %v)", tt.args.ctx, tt.args.payload)
		})
	}
}

func Test_cloudStorageClient_GetSignedURL(t *testing.T) {
	helper := newGcsClientHelper(t)
	defaultDuration := 5 * time.Minute

	type fields struct {
		config *config.CloudStorageConfig
		client *storage.Client
	}
	type args struct {
		filePath models.CloudStoragePayload
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		doMock  func(a *args)
		wantErr bool
	}{
		{
			name: "failed get signed url",
			fields: fields{
				config: helper.defaultConfig,
				client: helper.client,
			},
			args: args{
				filePath: models.NewCloudStoragePayload("my_path_for_writer/my_writer.txt"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(&tt.args)
			}

			cs := &cloudStorageClient{
				config: tt.fields.config,
				client: tt.fields.client,
			}
			_, err := cs.GetSignedURL(tt.args.filePath, defaultDuration)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
