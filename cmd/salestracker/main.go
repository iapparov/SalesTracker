// @title           salesTracker API
// @version         1.0
// @description     API для управления продажами и транзакциями.
// @BasePath        /

package main

import (
	wbzlog "github.com/wb-go/wbf/zlog"
	"go.uber.org/fx"
	"salestracker/internal/app/analytics"
	"salestracker/internal/app/transactions"
	"salestracker/internal/config"
	"salestracker/internal/di"
	"salestracker/internal/storage/postgres"
	"salestracker/internal/web/handlers"
)

func main() {
	wbzlog.Init()
	app := fx.New(
		fx.Provide(
			config.NewAppConfig,
			postgres.NewPostgres,

			func(db *postgres.Postgres) analytics.AnalyticStorageProvider {
				return db
			},
			analytics.NewAnalyticService,

			func(db *postgres.Postgres) transactions.TransactionStorageProvider {
				return db
			},
			transactions.NewTransactionService,

			func(service *analytics.AnalyticService) handlers.AnalyticsIFace {
				return service
			},
			handlers.NewAnalyticHandler,

			func(service *transactions.TransactionService) handlers.TransactionIFace {
				return service
			},
			handlers.NewTransactionHandler,
		),
		fx.Invoke(
			di.StartHTTPServer,
			di.ClosePostgresOnStop,
		),
	)

	app.Run()
}
