/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"bitbucket.org/Amartha/go-accounting/internal/contract"
	"bitbucket.org/Amartha/go-accounting/internal/deliveries/job"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "worker",
	Short: "Worker application to configuring and running a job",
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
	j *job.Job
)

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(runJobCmd)

	runJobCmd.Flags().StringP(runJobCmdName, "n", "", "job name")
	runJobCmd.MarkFlagRequired(runJobCmdName)
	runJobCmd.Flags().StringP(runJobCmdVersion, "v", "", "job version")
	runJobCmd.MarkFlagRequired(runJobCmdVersion)
	runJobCmd.Flags().StringP(runJobCmdDate, "d", "", "job running date")
}

var (
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List job name and version",
		Long:  ``,
		Run:   list,
	}
)

func list(ccmd *cobra.Command, args []string) {
	for version, l := range j.Routes {
		for name := range l {
			list := fmt.Sprintf("version=%s, name=%s", version, name)
			fmt.Println(list)
		}
	}
}

var (
	runJobCmd = &cobra.Command{
		Use:     "run",
		Short:   "Run execution job",
		Long:    ``,
		Example: "worker run -n={job-name} -v={job-version} -d={job-date}",
		Run:     runJob,
	}
	runJobCmdName    = "name"
	runJobCmdVersion = "version"
	runJobCmdDate    = "date"
)

func runJob(ccmd *cobra.Command, args []string) {
	ctx := context.Background()
	name, _ := ccmd.Flags().GetString(runJobCmdName)
	version, _ := ccmd.Flags().GetString(runJobCmdVersion)
	date, _ := ccmd.Flags().GetString(runJobCmdDate)

	c, _, err := contract.New(ctx)
	if err != nil {
		log.Fatalf("failed to setup app: %v", err)
	}

	defer func() {
		xlog.Sync()
		c.DB.Close()
		c.Cache.Close()
		c.Flagger.Close()
	}()

	j = job.New(c.Config, c.Service)
	j.Start(ctx, name, version, date)

	xlog.Info(ctx, "job server stopped!")
}
