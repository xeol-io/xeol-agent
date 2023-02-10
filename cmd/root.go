package cmd

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/noqcks/xeol-agent/agent/mode"

	agent "github.com/noqcks/xeol-agent/agent"
	"github.com/noqcks/xeol-agent/agent/presenter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "xeol-agent",
	Short: "xeol-agent tells xeol.io which resources are in use in your Kubernetes Cluster",
	Long:  `xeol-agent can poll Kubernetes Cluster API(s) to tell xeol.io which resources are currently in-use`,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if appConfig.Dev.ProfileCPU {
			f, err := os.Create("cpu.profile")
			if err != nil {
				log.Errorf("unable to create CPU profile: %+v", err)
			} else {
				err := pprof.StartCPUProfile(f)
				if err != nil {
					log.Errorf("unable to start CPU profile: %+v", err)
				}
			}
		}

		if len(args) > 0 {
			err := cmd.Help()
			if err != nil {
				log.Errorf(err.Error())
				os.Exit(1)
			}
			os.Exit(1)
		}

		switch appConfig.RunMode {
		case mode.PeriodicPolling:
			agent.PeriodicallyGetInventoryReport(appConfig)
		default:
			report, err := agent.GetInventoryReport(appConfig)
			if appConfig.Dev.ProfileCPU {
				pprof.StopCPUProfile()
			}
			if err != nil {
				log.Errorf("Failed to get Image Results: %+v", err)
				os.Exit(1)
			} else {
				err := agent.HandleReport(report, appConfig)
				if err != nil {
					log.Errorf("Failed to handle Image Results: %+v", err)
					os.Exit(1)
				}
			}
		}
	},
}

func init() {
	// output & formatting options
	opt := "output"
	rootCmd.Flags().StringP(
		opt, "o", presenter.JSONPresenter.String(),
		fmt.Sprintf("report output formatter, options=%v", presenter.Options),
	)
	if err := viper.BindPFlag(opt, rootCmd.Flags().Lookup(opt)); err != nil {
		fmt.Printf("unable to bind flag '%s': %+v", opt, err)
		os.Exit(1)
	}

	opt = "kubeconfig"
	rootCmd.Flags().StringP(opt, "k", "", "(optional) absolute path to the kubeconfig file")
	if err := viper.BindPFlag(opt+".path", rootCmd.Flags().Lookup(opt)); err != nil {
		fmt.Printf("unable to bind flag '%s': %+v", opt, err)
		os.Exit(1)
	}

	opt = "mode"
	rootCmd.Flags().StringP(opt, "m", mode.AdHoc.String(), fmt.Sprintf("execution mode, options=%v", mode.Modes))
	if err := viper.BindPFlag(opt, rootCmd.Flags().Lookup(opt)); err != nil {
		fmt.Printf("unable to bind flag '%s': %+v", opt, err)
		os.Exit(1)
	}

	opt = "polling-interval-minutes"
	rootCmd.Flags().StringP(opt, "p", "300", "If mode is 'periodic', this specifies the interval")
	if err := viper.BindPFlag(opt, rootCmd.Flags().Lookup(opt)); err != nil {
		fmt.Printf("unable to bind flag '%s': %+v", opt, err)
		os.Exit(1)
	}
}
