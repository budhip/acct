package godbledger

import (
	"context"
	"errors"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/darcys22/godbledger/godbledger/core"
	"github.com/darcys22/godbledger/godbledger/db/mysqldb"
	"github.com/darcys22/godbledger/godbledger/ledger"
)

type GoDBLedger interface {
	InsertAccount(ctx context.Context, accountID string) (err error)
	GetAccount(ctx context.Context, code string) (resp *core.Account, err error)
	InsertTransaction(ctx context.Context, txn Transaction) (resp string, err error)
	FindTransaction(ctx context.Context, txnID string) (resp *core.Transaction, err error)
	GetCurrency(ctx context.Context, currency string) (resp Currency, err error)
}

type godbledger struct {
	ledger ledger.Ledger
}

func New(db *mysqldb.Database) GoDBLedger {
	l := ledger.Ledger{
		LedgerDb: db,
	}
	return &godbledger{ledger: l}
}

func (g *godbledger) InsertAccount(ctx context.Context, accountID string) (err error) {
	defer func() {
		logDB(ctx, err)
	}()
	err = g.ledger.InsertAccount(accountID)
	return
}

func (g *godbledger) GetAccount(ctx context.Context, code string) (resp *core.Account, err error) {
	defer func() {
		logDB(ctx, err)
	}()
	resp, err = g.ledger.LedgerDb.FindAccount(code)
	return
}

func (g *godbledger) InsertTransaction(ctx context.Context, txn Transaction) (resp string, err error) {
	defer func() {
		logDB(ctx, err)
	}()
	resp, err = g.ledger.LedgerDb.AddTransaction(txn)
	return
}

func (g *godbledger) FindTransaction(ctx context.Context, txnID string) (resp *core.Transaction, err error) {
	defer func() {
		logDB(ctx, err)
	}()
	resp, err = g.ledger.LedgerDb.FindTransaction(txnID)
	return
}

func (g *godbledger) GetCurrency(ctx context.Context, currency string) (resp Currency, err error) {
	defer func() {
		logDB(ctx, err)
	}()
	resp, err = g.ledger.GetCurrency(currency)
	return
}

func logDB(ctx context.Context, err error) {
	if err != nil {
		if errors.Is(err, models.ErrNoRows) {
			xlog.Debug(ctx, "[GO-DB-LEDGER]", xlog.String("status", "error"), xlog.Err(err))
		} else {
			xlog.Debug(ctx, "[GO-DB-LEDGER]", xlog.String("status", "error"), xlog.Err(err))
		}
	} else {
		xlog.Debug(ctx, "[GO-DB-LEDGER]", xlog.String("status", "success"))
	}
}
