package cmd

import (
	"os"
	"os/signal"
	"time"

	"tvctrl/internal"
	"tvctrl/logger"

	"github.com/spf13/cobra"
)

var cfg = internal.DefaultConfig
var noCache bool
var rootCmd = &cobra.Command{
	Use:   "tvctrl",
	Short: "Simple TV controller using AVTransport",
	Run: func(cmd *cobra.Command, args []string) {
		stop := make(chan struct{})
		//FLAG INVERSION HERE
		cfg.UseCache = !noCache
		// Cache commands exit early
		if internal.HandleCacheCommands(cfg) {
			os.Exit(0)
		}
		if cfg.SelectCache >= 0 {
			internal.LoadCachedTV(&cfg)
		}

		serverRunning := false

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
			internal.ServeDirGo(cfg, stop)
			time.Sleep(500 * time.Millisecond)
			serverRunning = true
		}
		// ---- END ----

		internal.RunScript(cfg)
		if !serverRunning {
			return
		}

		logger.Info("tvctrl running — press Ctrl+C to exit")
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)

		<-sig       // wait for Ctrl+C
		close(stop) // trigger shutdown

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

	// ---- VERSION ----
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("tvctrl {{.Version}}\n")

	rootCmd.CompletionOptions.DisableDefaultCmd = false

	// ---- execution flags ----
	rootCmd.Flags().BoolVar(&cfg.ProbeOnly, "probe-only", cfg.ProbeOnly, "Probe AVTransport only when using mode: auto")
	rootCmd.Flags().StringVar(&cfg.Mode, "mode", cfg.Mode, "Execution mode (auto/manual/scan)")

	// ---- cache flags ----
	rootCmd.Flags().BoolVar(&cfg.AutoCache, "auto-cache", cfg.AutoCache, "Skip cache save confirmation")
	rootCmd.Flags().BoolVar(&noCache, "no-cache", false, "Disable cache usage")
	rootCmd.Flags().BoolVar(&cfg.ListCache, "list-cache", cfg.ListCache, "List cached AVTransport devices")
	rootCmd.Flags().StringVar(&cfg.ForgetCache, "forget-cache", cfg.ForgetCache, "Forget cache (interactive | IP | all)")
	rootCmd.Flags().IntVar(&cfg.SelectCache, "select-cache", -1, "Select cached device by index")

	// ---- scan flags ----
	rootCmd.Flags().StringVar(&cfg.Subnet, "subnet", cfg.Subnet, "Subnet to scan (e.g. 192.168.1.0/24)")
	rootCmd.Flags().BoolVar(&cfg.DeepSearch, "deep-search", cfg.DeepSearch, "Use a bigger list when probing for device enpoints (Method:slower and more noisy)")
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
	rootCmd.Flags().StringVar(&cfg.ServePort, "LPort", cfg.ServePort, "Local port to serve")
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
  --select-cache	  Select cached device by index

Scan:
  --deep-search		  Use a bigger list when probing for device enpoints (Method:slower and more noisy) 
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


Autocomplete mechanism (Optional UI helper)-> 
		tvctrl install-completion
		exec $SHELL
`)
}
