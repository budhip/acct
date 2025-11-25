package services

import (
	"errors"

	"bitbucket.org/Amartha/go-accounting/internal/models"
)

func checkDatabaseError(err error, code ...string) error {
	if errors.Is(err, models.ErrNoRows) {
		err = models.GetErrMap(models.ErrKeyDataNotFound)
		if len(code) > 0 {
			err = models.GetErrMap(code[0])
		}
	} else {
		err = models.GetErrMap(models.ErrKeyDatabaseError, err.Error())
	}

	return err
}
