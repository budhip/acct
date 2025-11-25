package cache

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	mock2 "bitbucket.org/Amartha/go-accounting/internal/repositories/cache/mock"
)

func cacheTestHelper(t *testing.T) (redismock.ClientMock, CacheRepository) {
	t.Helper()
	t.Parallel()

	db, mock := redismock.NewClientMock()
	cacheRepo := NewCacheRepository(db)

	return mock, cacheRepo
}

func TestCacheRepository_Ping(t *testing.T) {
	mock, rc := cacheTestHelper(t)

	tests := []struct {
		name    string
		doMock  func()
		wantErr bool
	}{
		{
			name: "success case",
			doMock: func() {
				mock.ExpectPing().SetVal("success")
			},
			wantErr: false,
		},
		{
			name: "error case - ping error",
			doMock: func() {
				mock.ExpectPing().SetErr(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - redis nil",
			doMock: func() {
				mock.ExpectPing().RedisNil()
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock()
			}
			err := rc.Ping(context.TODO())
			assert.Equal(t, err != nil, tt.wantErr)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
			mock.ClearExpect()
		})
	}
}

func TestCacheRepository_SetIfNotExists(t *testing.T) {
	mock, rc := cacheTestHelper(t)

	type args struct {
		key  string
		data interface{}
		ttl  time.Duration
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test success",
			args: args{
				key:  "123456789",
				data: "Success",
				ttl:  30 * time.Second,
			},
			want:    true,
			wantErr: false,
			doMock: func(args args) {
				mock.ExpectSetNX(args.key, args.data, args.ttl).SetVal(true)
			},
		},
		{
			name: "test error",
			args: args{
				key:  "123456789",
				data: "Success",
				ttl:  30 * time.Second,
			},
			wantErr: true,
			doMock: func(args args) {
				mock.ExpectSetNX(args.key, args.data, args.ttl).SetErr(redis.ErrClosed)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}

			got, err := rc.SetIfNotExists(context.TODO(), tt.args.key, tt.args.data, tt.args.ttl)
			assert.Equal(t, got, tt.want)
			assert.Equal(t, tt.wantErr, err != nil)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
			mock.ClearExpect()
		})
	}
}

func TestCacheRepository_Set(t *testing.T) {
	mock, rc := cacheTestHelper(t)

	type args struct {
		key  string
		data interface{}
		ttl  time.Duration
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test success",
			args: args{
				key:  "123456789",
				data: "Success",
				ttl:  30 * time.Second,
			},
			want:    true,
			wantErr: false,
			doMock: func(args args) {
				mock.ExpectSet(args.key, args.data, args.ttl).SetVal(args.key)
			},
		},
		{
			name: "test error",
			args: args{
				key:  "123456789",
				data: "Success",
				ttl:  30 * time.Second,
			},
			wantErr: true,
			doMock: func(args args) {
				mock.ExpectSet(args.key, args.data, args.ttl).SetErr(redis.ErrClosed)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}

			err := rc.Set(context.TODO(), tt.args.key, tt.args.data, tt.args.ttl)
			assert.Equal(t, tt.wantErr, err != nil)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
			mock.ClearExpect()
		})
	}
}

func TestCacheRepository_Get(t *testing.T) {
	mock, rc := cacheTestHelper(t)

	type args struct {
		key  string
		data string
		ttl  time.Duration
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test success",
			args: args{
				key:  "123456789",
				data: "Success",
				ttl:  30 * time.Second,
			},
			want:    "Success",
			wantErr: false,
			doMock: func(args args) {
				mock.ExpectGet(args.key).SetVal(args.data)
			},
		},
		{
			name: "test error - RedisNil",
			args: args{
				key:  "123456789",
				data: "Success",
				ttl:  30 * time.Second,
			},
			want:    "",
			wantErr: true,
			doMock: func(args args) {
				mock.ExpectGet(args.key).RedisNil()
			},
		},
		{
			name: "test error",
			args: args{
				key:  "123456789",
				data: "",
				ttl:  30 * time.Second,
			},
			want:    "",
			wantErr: true,
			doMock: func(args args) {
				mock.ExpectGet(args.key).SetErr(redis.ErrClosed)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}

			got, err := rc.Get(context.TODO(), tt.args.key)
			assert.Equal(t, got, tt.want)
			assert.Equal(t, err != nil, tt.wantErr)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
			mock.ClearExpect()
		})
	}
}

func TestCacheRepository_Del(t *testing.T) {
	mock, rc := cacheTestHelper(t)

	type args struct {
		key  string
		data string
		ttl  time.Duration
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test success",
			args: args{
				key: "123456789",
			},
			wantErr: false,
			doMock: func(args args) {
				mock.ExpectDel(args.key).SetVal(1)
			},
		},
		{
			name: "test error",
			args: args{
				key:  "123456789",
				data: "",
				ttl:  30 * time.Second,
			},
			wantErr: true,
			doMock: func(args args) {
				mock.ExpectDel(args.key).SetErr(redis.ErrClosed)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}

			err := rc.Del(context.TODO(), tt.args.key)
			assert.Equal(t, err != nil, tt.wantErr)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
			mock.ClearExpect()
		})
	}
}

func TestCacheRepository_GetIncrement(t *testing.T) {
	mock, rc := cacheTestHelper(t)

	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		want    int64
		wantErr bool
	}{
		{
			name: "test success",
			args: args{
				key: "testing_success",
			},
			want:    1,
			wantErr: false,
			doMock: func(args args) {
				mock.ExpectIncr(args.key).SetVal(1)
			},
		},
		{
			name: "test error",
			args: args{
				key: "testing_error",
			},
			wantErr: true,
			doMock: func(args args) {
				mock.ExpectIncr(args.key).SetErr(redis.ErrClosed)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}

			got, err := rc.GetIncrement(context.TODO(), tt.args.key)
			assert.Equal(t, got, tt.want)
			assert.Equal(t, err != nil, tt.wantErr)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
			mock.ClearExpect()
		})
	}
}

func TestCacheRepository_GetAll(t *testing.T) {
	mock, rc := cacheTestHelper(t)

	key := "category_code_212_seq"
	type args struct {
		key  string
		data string
	}
	tests := []struct {
		name     string
		args     args
		doMock   func(args args)
		wantResp map[string]string
		wantErr  bool
	}{
		{
			name: "success case",
			args: args{
				key:  key,
				data: "100",
			},
			doMock: func(args args) {
				mock.ExpectScan(0, key, 0).SetVal([]string{key}, 0)
				mock.ExpectGet(args.key).SetVal(args.data)
			},
			wantResp: map[string]string{"category_code_212_seq": "100"},
			wantErr:  false,
		},
		{
			name: "error case - redis get key",
			args: args{
				key:  key,
				data: "100",
			},
			doMock: func(args args) {
				mock.ExpectScan(0, key, 0).SetVal([]string{key}, 0)
				mock.ExpectGet(args.key).SetErr(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - redis ScanIterator",
			args: args{
				key:  key,
				data: "100",
			},
			doMock: func(args args) {
				mock.ExpectScan(0, key, 0).SetErr(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}

			got, err := rc.GetAll(context.TODO(), tt.args.key)
			assert.Equal(t, got, tt.wantResp)
			assert.Equal(t, err != nil, tt.wantErr)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
			mock.ClearExpect()
		})
	}
}

func TestGetOrSet(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockRepo := mock2.NewMockCacheRepository(mockCtrl)

	type genericOutput map[string]string

	type args[T any] struct {
		cacheRepo CacheRepository
		opts      models.GetOrSetCacheOpts[genericOutput]
	}
	type testCase[T any] struct {
		name      string
		args      args[T]
		doMock    func(args args[T])
		wantValue T
		wantErr   bool
	}
	tests := []testCase[genericOutput]{
		{
			name: "success case - get from cache",
			args: args[genericOutput]{
				cacheRepo: mockRepo,
				opts: models.GetOrSetCacheOpts[genericOutput]{
					Key: "random_key_here",
					Callback: func() (genericOutput, error) {
						return genericOutput{
							"key1": "val1",
							"key2": "val2",
						}, nil
					},
				},
			},
			doMock: func(args args[genericOutput]) {
				mockRepo.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(`{"key1": "val1", "key2": "val2"}`, nil)
			},
			wantValue: genericOutput{
				"key1": "val1",
				"key2": "val2",
			},
			wantErr: false,
		},
		{
			name: "success case - get from callback and save to cache",
			args: args[genericOutput]{
				cacheRepo: mockRepo,
				opts: models.GetOrSetCacheOpts[genericOutput]{
					Key: "random_key_here",
					Callback: func() (genericOutput, error) {
						return genericOutput{
							"key1": "val1",
							"key2": "val2",
						}, nil
					},
				},
			},
			doMock: func(args args[genericOutput]) {
				mockRepo.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return("", models.GetErrMap(models.ErrKeyDataNotFound))
				mockRepo.EXPECT().
					Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantValue: genericOutput{
				"key1": "val1",
				"key2": "val2",
			},
			wantErr: false,
		},
		{
			name: "failed case - error set value callback to cache",
			args: args[genericOutput]{
				cacheRepo: mockRepo,
				opts: models.GetOrSetCacheOpts[genericOutput]{
					Key: "random_key_here",
					Callback: func() (genericOutput, error) {
						return genericOutput{
							"key1": "val1",
							"key2": "val2",
						}, nil
					},
				},
			},
			doMock: func(args args[genericOutput]) {
				mockRepo.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return("", models.GetErrMap(models.ErrKeyDataNotFound))
				mockRepo.EXPECT().
					Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(assert.AnError)
			},
			wantValue: genericOutput{
				"key1": "val1",
				"key2": "val2",
			},
			wantErr: false,
		},
		{
			name: "failed case - error from callback",
			args: args[genericOutput]{
				cacheRepo: mockRepo,
				opts: models.GetOrSetCacheOpts[genericOutput]{
					Key: "random_key_here",
					Callback: func() (genericOutput, error) {
						return nil, assert.AnError
					},
				},
			},
			doMock: func(args args[genericOutput]) {
				mockRepo.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return("", models.GetErrMap(models.ErrKeyDataNotFound))
			},
			wantValue: nil,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}

			gotValue, err := GetOrSet(tt.args.cacheRepo, tt.args.opts)
			assert.Equalf(t, err != nil, tt.wantErr, "GetOrSet(%v, %v)", tt.args.cacheRepo, tt.args.opts)
			assert.Equalf(t, tt.wantValue, gotValue, "GetOrSet(%v, %v)", tt.args.cacheRepo, tt.args.opts)

		})
	}
}

func TestCacheRepository_DeleteKeysWithPrefix(t *testing.T) {
	mock, rc := cacheTestHelper(t)
	key := "pas_loan_partner_key_211001000000022_*"
	type args struct {
		key  string
		data string
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "success case",
			args: args{
				key:  key,
				data: "100",
			},
			doMock: func(args args) {
				mock.ExpectScan(0, args.key, 100).SetVal([]string{args.key}, 0)
				mock.ExpectDel(args.key).SetVal(1)
			},
			wantErr: false,
		},
		{
			name: "error case - redis del key",
			args: args{
				key:  key,
				data: "100",
			},
			doMock: func(args args) {
				mock.ExpectScan(0, args.key, 100).SetVal([]string{args.key}, 0)
				mock.ExpectDel(args.key).SetErr(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - redis ScanIterator",
			args: args{
				key:  key,
				data: "100",
			},
			doMock: func(args args) {
				mock.ExpectScan(0, args.key, 100).SetErr(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}

			err := rc.DeleteKeysWithPrefix(context.TODO(), tt.args.key)
			assert.Equal(t, err != nil, tt.wantErr)
			t.Log("asdasd", err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
			mock.ClearExpect()
		})
	}
}
