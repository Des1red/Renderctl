package cmd

import (
	"os"
	"strings"
	"time"
	"tvctrl/internal"
	"tvctrl/logger"
)

func Run() {
	cfg := handleArguments()

	if internal.HandleCacheCommands(cfg) {
		os.Exit(0)
	}

	if cfg.ProbeOnly == false {
		if cfg.LFile == "" {
			logger.Fatal("Missing -Lf (media file)")
			os.Exit(1)
		}

		if err := internal.ValidateFile(cfg.LFile); err != nil {
			logger.Fatal("Invalid file: %v", err)
			os.Exit(1)
		}

		//1.
		cfg.LIP = internal.LocalIP(cfg.LIP)
		internal.ServeDirGo(cfg)
		// 2.
		time.Sleep(500 * time.Millisecond)
	}

	// 3. Run the AVTransport logic
	internal.RunScript(cfg)

	select {}
}

func handleArguments() internal.Config {
	cfg := internal.DefaultConfig
	args := os.Args

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h":
			showHelp()
			os.Exit(0)

		case "-Tip":
			if i+1 < len(args) {
				cfg.TIP = args[i+1]
				i++
			}

		case "-Tport":
			if i+1 < len(args) {
				cfg.TPort = args[i+1]
				i++
			}

		case "-Tpath":
			if i+1 < len(args) {
				cfg.TPath = args[i+1]
				i++
			}

		case "-type":
			if i+1 < len(args) {
				cfg.TVVendor = args[i+1]
				i++
			}

		case "-Lf":
			if i+1 < len(args) {
				cfg.LFile = args[i+1]
				i++
			}

		case "-Lip":
			if i+1 < len(args) {
				cfg.LIP = args[i+1]
				i++
			}

		case "-mode":
			if i+1 < len(args) {
				cfg.Mode = args[i+1] // auto | manual
				i++
			}
		case "-select":
			if i+1 < len(args) {
				cfg.CachedTv = args[i+1] // auto | manual
				i++
			}
		case "--no-cache:":
			cfg.UseCache = true

		case "--probe-only":
			cfg.ProbeOnly = true

		case "--forget-cache":
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				cfg.ForgetCache = args[i+1]
				i++
			} else {
				cfg.ForgetCache = "interactive"
			}
		case "--list-cache":
			cfg.ListCache = true

		}
	}

	return cfg
}

func showHelp() {
	help := `
	--probe-only : on auto mode after probe , do not trasmit message
	-mode		 : Execution mode (auto/manual)

	-Tip    	 : TV IP address
	-Tport  	 : TV SOAP port
	-Tpath  	 : TV SOAP control path
	-type		 : Tv vendor 
	
	-Lf     	 : local media file
	-Lip    	 : local IP for serving file

	--list-cache              List cached AVTransport devices
	--forget-cache           # interactive (list & choose)
	--forget-cache <IP>      # delete specific TV
	--forget-cache all       # wipe cache

	-h      	 : show help
	`

	logger.Notify(help)
}
