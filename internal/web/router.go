package web

import (
	httpSwagger "github.com/swaggo/http-swagger"
	wbgin "github.com/wb-go/wbf/ginext"
	_ "salestracker/docs"
	"salestracker/internal/web/handlers"
)

func RegisterRoutes(engine *wbgin.Engine, transactionHandler *handlers.TransactionHandler, analyticsHandler *handlers.AnalyticsHandler) {
	api := engine.Group("/api")
	api.GET("/swagger/*any", func(c *wbgin.Context) {
		httpSwagger.WrapHandler(c.Writer, c.Request)
	})

	api.POST("/items", transactionHandler.CreateTransaction)
	api.GET("/items", transactionHandler.GetAllTransactions)
	api.GET("/items/:id", transactionHandler.GetTransaction)
	api.PUT("/items/:id", transactionHandler.PutTransaction)
	api.DELETE("/items/:id", transactionHandler.DeleteTransaction)
	api.GET("/items/export", transactionHandler.GetCSV)

	api.GET("/analytics", analyticsHandler.GetAnalys)
	api.GET("/analytics/export", analyticsHandler.GetCSV)

}
