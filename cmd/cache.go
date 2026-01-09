package cmd

import (
	"tvctrl/internal"

	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage cached AVTransport devices",
}

var cacheListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cached devices",
	Run: func(cmd *cobra.Command, args []string) {
		cfg.ListCache = true
		internal.HandleCacheCommands(cfg)
	},
}

var cacheForgetCmd = &cobra.Command{
	Use:   "forget [IP|all]",
	Short: "Forget cached device(s)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cfg.ForgetCache = "interactive"
		} else {
			cfg.ForgetCache = args[0]
		}
		internal.HandleCacheCommands(cfg)
	},
}

func init() {
	cacheCmd.AddCommand(cacheListCmd)
	cacheCmd.AddCommand(cacheForgetCmd)
}
