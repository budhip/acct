package accounting

import (
	"bitbucket.org/Amartha/go-accounting/internal/services"
	"github.com/labstack/echo/v4"
)

type accountingHandler struct {
	services.AccountingService
}

func New(app *echo.Group, accountingSvc services.AccountingService) {
	ah := accountingHandler{
		AccountingService: accountingSvc,
	}

	trialBalance := app.Group("/trial-balances")
	trialBalance.GET("/details", ah.getTrialBalanceDetails)
	trialBalance.GET("/details/download", ah.sendTrialBalanceDetailsToEmail)
	trialBalance.GET("/download", ah.sendTrialBalanceSummaryToEmail)

}
