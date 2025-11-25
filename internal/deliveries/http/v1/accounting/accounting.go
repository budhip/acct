package accounting

import (
	"bitbucket.org/Amartha/go-accounting/internal/services"

	"github.com/labstack/echo/v4"
)

type accountingHandler struct {
	services.AccountingService
	services.JournalService
	services.TrialBalanceService
}

func New(app *echo.Group, accountingSrv services.AccountingService, journalSrv services.JournalService, trialBalanceSrv services.TrialBalanceService) {
	ah := accountingHandler{
		accountingSrv,
		journalSrv,
		trialBalanceSrv,
	}
	trialBalance := app.Group("/trial-balances")
	trialBalance.GET("", ah.getTrialBalance)
	trialBalance.GET("/download", ah.downloadCSVgetTrialBalance)
	trialBalance.POST("/:period/close", ah.closeTrialBalance)
	trialBalance.POST("/adjustment", ah.adjustmentTrialBalance)

	trialBalance.GET("/details", ah.getTrialBalanceDetails)
	trialBalance.GET("/sub-categories/:subCategoryCode", ah.getTrialBalanceBySubCategoryCode)
	trialBalance.GET("/details/download", ah.sendToEmailCSVgetTrialBalanceDetails)

	generalLedger := app.Group("/general-ledgers")
	generalLedger.GET("", ah.generalLedger)
	generalLedger.GET("/download", ah.download)

	subLedger := app.Group("/sub-ledgers")
	subLedger.GET("", ah.getSubLedger)
	subLedger.GET("/count", ah.getSubLedgerCount)
	subLedger.GET("/accounts", ah.getSubLedgerAccounts)
	subLedger.GET("/send-email", ah.sendSubLedgerCSVToEmail)
	subLedger.GET("/download", ah.downloadSubLedgerCSV)

	journal := app.Group("/journals")
	journal.POST("", ah.create)
	journal.POST("/publish", ah.publish)
	journal.POST("/upload", ah.uploadJournal)
	journal.GET("/:transactionId", ah.getByTransactionId)

	balanceSheet := app.Group("/balance-sheets")
	balanceSheet.GET("", ah.getBalanceSheet)
	balanceSheet.GET("/download", ah.downloadCSVBalanceSheet)

	app.POST("/jobs/trial-balance", ah.runJobTrialBalance)
}
