package job

import (
	"context"
	"errors"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/config"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/services"
	"bitbucket.org/Amartha/go-x/log/ctxdata"

	v1accounting "bitbucket.org/Amartha/go-accounting/internal/deliveries/job/v1/accounting"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/google/uuid"
)

type JobRoutes map[string]map[string]func(ctx context.Context, date time.Time) error

type Job struct {
	Routes JobRoutes
}

func New(cfg *config.Configuration, srv *services.Services) *Job {
	v1group := "v1"

	jobRoutes := map[string]map[string]func(ctx context.Context, date time.Time) error{
		v1group: v1accounting.Routes(srv.Accounting),
		// add other version routes
	}

	return &Job{jobRoutes}
}

func (j *Job) Start(ctx context.Context, jobName, version, date string) {
	var err error
	runningDate, _ := atime.NowZeroTime()
	start := atime.Now()

	defer func() {
		logJob(ctx, jobName, version, start, runningDate, err)
	}()

	fn, ok := j.Routes[version][jobName]
	if !ok {
		err = errors.New("invalid version or job name")
		return
	}

	ctx = ctxdata.Sets(ctx, ctxdata.SetCorrelationId(uuid.New().String()))
	if date != "" {
		runningDate, err = atime.ParseStringToDatetime(atime.DateFormatYYYYMMDD, date)
		if err != nil {
			return
		}
	}
	if err = fn(ctx, runningDate); err != nil {
		return
	}
}

func logJob(ctx context.Context, jobName, version string, start, runningDate time.Time, err error) {
	field := []xlog.Field{
		xlog.String("job-name", jobName),
		xlog.String("version", version),
		xlog.Any("execution-date", runningDate),
		xlog.Duration("processing-time", time.Since(start)),
	}
	if err != nil {
		field = append(field, xlog.String("status", "fail"), xlog.Err(err))
		xlog.Warn(ctx, "[JOB]", field...)
		return
	}
	field = append(field, xlog.String("status", "success"))
	xlog.Info(ctx, "[JOB]", field...)
}
