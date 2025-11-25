package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/contract"
	"bitbucket.org/Amartha/go-accounting/internal/deliveries/consumer"
	"bitbucket.org/Amartha/go-accounting/internal/deliveries/http"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/graceful"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "consumer",
	Short: "Consumer is a consumer application for handling message from kafka",
	Long:  ``,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	runConsumerCmd = &cobra.Command{
		Use:     "run",
		Short:   "Run consumer",
		Long:    fmt.Sprintf("Run consumer for handling message from kafka, available consumer type: %s", strings.Join(consumer.ListConsumerName, ",")),
		Example: "consumer run -n={consumer-type-name}",
		Run:     runConsumer,
	}
	runConsumerCmdName = "name"
)

func init() {
	rootCmd.AddCommand(runConsumerCmd)
	runConsumerCmd.Flags().StringP(runConsumerCmdName, "n", "", "consumer name")
	runConsumerCmd.MarkFlagRequired(runConsumerCmdName)
}

func runConsumer(ccmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(context.Background())
	name, _ := ccmd.Flags().GetString(runConsumerCmdName)
	var (
		starters []graceful.ProcessStarter
		stoppers []graceful.ProcessStopper
	)

	c, stopperContract, err := contract.New(ctx)
	if err != nil {
		timeout := 30 * time.Second
		if c != nil && c.Config.App.GracefulTimeout != 0 {
			timeout = c.Config.App.GracefulTimeout
		}
		graceful.StopProcess(timeout, stoppers...)
		log.Fatalf("failed to setup app: %v", err)
	}
	stoppers = append(stoppers, stopperContract)

	consumerServer, stopperConsumer, err := consumer.NewKafkaConsumer(ctx, name, c)
	if err != nil {
		xlog.Fatal(ctx, "error initializing kafka consumer", xlog.Err(err))
	}

	starterConsumer := consumerServer.Start()
	starters = append(starters, starterConsumer)
	stoppers = append(stoppers, stopperConsumer)

	httpServer := http.NewHealthCheck(c)
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
	cancel()
	xlog.Info(ctx, "consumer server stopped!")
}
