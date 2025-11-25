package main

import (
	"context"
	"log"
	"sync"

	"bitbucket.org/Amartha/go-accounting/internal/contract"
	"bitbucket.org/Amartha/go-accounting/internal/deliveries/http"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/graceful"
	xlog "bitbucket.org/Amartha/go-x/log"
)

func main() {
	var (
		ctx      = context.Background()
		starters []graceful.ProcessStarter
		stoppers []graceful.ProcessStopper
	)

	c, stopperContract, err := contract.New(ctx)
	if err != nil {
		log.Fatalf("failed init contract: %v", err)
	}
	stoppers = append(stoppers, stopperContract)

	httpServer := http.NewHTTPServer(ctx, c)
	starterApi, stopperApi := httpServer.Start(ctx)
	starters = append(starters, starterApi)
	stoppers = append(stoppers, stopperApi)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		graceful.StartProcessAtBackground(starters...)
		graceful.StopProcessAtBackground(c.Config.App.GracefulTimeout, stoppers...)
		wg.Done()
	}()

	wg.Wait()
	xlog.Info(ctx, "http server stopped!")
}
