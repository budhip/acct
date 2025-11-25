package mysql

import (
	"context"
	"errors"
	"reflect"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/godbledger"
	xlog "bitbucket.org/Amartha/go-x/log"
)

var CurrencyIDR = godbledger.CurrencyIDR

const logMessageDatabase = "[DATABASE]"

func logSQL(ctx context.Context, err error, times ...time.Time) {
	var process time.Duration
	if len(times) != 0 {
		process = time.Since(times[0])
	}
	if err != nil {
		logFields := []xlog.Field{
			xlog.String("status", "error"),
			xlog.Duration("query-execution-time", process),
			xlog.Err(err),
		}
		if errors.Is(err, models.ErrNoRows) {
			xlog.Debug(ctx, logMessageDatabase, logFields...)
		} else {
			xlog.Debug(ctx, logMessageDatabase, logFields...)
		}
	} else {
		xlog.Debug(ctx, logMessageDatabase, xlog.String("status", "success"), xlog.Duration("query-execution-time", process))
	}
}

// get value each fields from entities
func getFieldValues(i interface{}) ([]interface{}, error) {
	entities := reflect.ValueOf(i)
	if entities.Kind() != reflect.Struct {
		return nil, errors.New("invalid entity for get field values")
	}

	values := make([]interface{}, entities.NumField())
	for i := 0; i < entities.NumField(); i++ {
		v := entities.Field(i).Interface()
		values[i] = v
	}
	return values, nil
}

func databaseError(err error) error {
	return models.GetErrMap(models.ErrKeyDatabaseError, err.Error())
}

func toInterface(strings []string) []interface{} {
	values := make([]interface{}, len(strings))
	for i := 0; i < len(strings); i++ {
		v := strings[i]
		values[i] = v
	}
	return values
}

func toPlaceholders(strings []string) []string {
	placeholders := make([]string, len(strings))
	for i := range strings {
		placeholders[i] = "?"
	}
	return placeholders
}

func buildInClause(values []string) (placeholders string, args []interface{}) {
	for i, v := range values {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, v)
	}
	return
}
