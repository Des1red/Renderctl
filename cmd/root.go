package cmd

import (
	"os"
	"time"

	"tvctrl/internal"
	"tvctrl/logger"

	"github.com/spf13/cobra"
)

var cfg = internal.DefaultConfig

var rootCmd = &cobra.Command{
	Use:   "tvctrl",
	Short: "Simple TV controller using AVTransport",
	Run: func(cmd *cobra.Command, args []string) {

		// Cache commands exit early
		if internal.HandleCacheCommands(cfg) {
			os.Exit(0)
		}

		// ---- PRE-RUN LOGIC ----
		if cfg.Mode != "scan" && !cfg.ProbeOnly {
			if cfg.LFile == "" {
				logger.Fatal("Missing -Lf (media file)")
				os.Exit(1)
			}

			if err := internal.ValidateFile(cfg.LFile); err != nil {
				logger.Fatal("Invalid file: %v", err)
				os.Exit(1)
			}

			cfg.LIP = internal.LocalIP(cfg.LIP)
			internal.ServeDirGo(cfg)
			time.Sleep(500 * time.Millisecond)
		}
		// ---- END ----

		internal.RunScript(cfg)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal("%v", err)
		os.Exit(1)
	}
}

func init() {
	initHelpTemplate()

	// ---- execution flags ----
	rootCmd.Flags().BoolVar(&cfg.ProbeOnly, "probe-only", cfg.ProbeOnly, "Probe AVTransport only")
	rootCmd.Flags().StringVar(&cfg.Mode, "mode", cfg.Mode, "Execution mode (auto/manual/scan)")

	// ---- cache flags ----
	rootCmd.Flags().BoolVar(&cfg.AutoCache, "auto-cache", cfg.AutoCache, "Skip cache save confirmation")
	rootCmd.Flags().BoolVar(&cfg.UseCache, "no-cache", cfg.UseCache, "Disable cache usage")
	rootCmd.Flags().BoolVar(&cfg.ListCache, "list-cache", cfg.ListCache, "List cached AVTransport devices")
	rootCmd.Flags().StringVar(&cfg.ForgetCache, "forget-cache", cfg.ForgetCache, "Forget cache (interactive | IP | all)")

	// ---- scan flags ----
	rootCmd.Flags().StringVar(&cfg.Subnet, "subnet", cfg.Subnet, "Subnet to scan (e.g. 192.168.1.0/24)")
	rootCmd.Flags().BoolVar(&cfg.Discover, "ssdp", cfg.Discover, "Enable SSDP discovery")

	// ---- TV flags ----
	rootCmd.Flags().StringVar(&cfg.TIP, "Tip", cfg.TIP, "TV IP address")
	rootCmd.Flags().StringVar(&cfg.TPort, "Tport", cfg.TPort, "TV SOAP port")
	rootCmd.Flags().StringVar(&cfg.TPath, "Tpath", cfg.TPath, "TV SOAP control path")
	rootCmd.Flags().StringVar(&cfg.TVVendor, "type", cfg.TVVendor, "TV vendor")

	// ---- media flags ----
	rootCmd.Flags().StringVar(&cfg.LFile, "Lf", cfg.LFile, "Local media file")
	rootCmd.Flags().StringVar(&cfg.LIP, "Lip", cfg.LIP, "Local IP for serving media")
	rootCmd.Flags().StringVar(&cfg.LDir, "Ldir", cfg.LDir, "Local directory to serve")
	rootCmd.Flags().StringVar(&cfg.ServePort, "LPort", cfg.LDir, "Local port to serve")
}

func initHelpTemplate() {
	rootCmd.SetHelpTemplate(`{{with (or .Long .Short)}}{{.}}

{{end}}Usage:
  {{.UseLine}}

Execution:
  --probe-only        Probe AVTransport only
  --mode string       Execution mode (auto/manual/scan)

Cache:
  --auto-cache        Skip cache save confirmation
  --no-cache          Disable cache usage
  --list-cache        List cached AVTransport devices
  --forget-cache      Forget cache (interactive | IP | all)

Scan:
  --subnet string     Subnet to scan (e.g. 192.168.1.0/24)
  --ssdp              Enable SSDP discovery

TV:
  --Tip string        TV IP address
  --Tport string      TV SOAP port
  --Tpath string      TV SOAP control path
  --type string       TV vendor

Media:
  --Lf string         Local media file
  --Lip string        Local IP for serving media
  --Ldir string       Local directory to serve
  --LPort string      Local port to serve
`)
}
