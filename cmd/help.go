package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func printHelp() {
	w := tabwriter.NewWriter(
		os.Stdout,
		0,   // min width
		0,   // tab width
		2,   // padding
		' ', // pad char
		0,   // flags
	)

	fmt.Fprintln(w, "renderctl - Simple TV controller using AVTransport")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "\trenderctl [flags]")
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Execution:")
	printFlag(w, "--tui", "", "					    Start program as TUI")
	printFlag(w, "--probe-only", "", " Probe AVTransport only")
	printFlag(w, "--mode", "string", "Execution mode (auto/manual/scan/stream)")

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Cache:")
	printFlag(w, "--auto-cache", "", "			      Skip cache save confirmation")
	printFlag(w, "--no-cache", "", "			        Disable cache usage")
	printFlag(w, "--list-cache", "", "			          List cached devices")
	printFlag(w, "--forget-cache", "string", "	 Forget cache (interactive | IP | all)")
	printFlag(w, "--select-cache", "int", "		    Select cached device by index")

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Scan:")
	printFlag(w, "--subnet", "  string", "   Subnet to scan (e.g. 192.168.1.0/24)")
	printFlag(w, "--deep-search", "", "	Extended endpoint probing")
	printFlag(w, "--ssdp", "", "		         Enable SSDP discovery")

	fmt.Fprintln(w)
	fmt.Fprintln(w, "TV:")
	printFlag(w, "--Tip         ", "string", "	   TV IP")
	printFlag(w, "--Tport    ", "string", "  SOAP port")
	printFlag(w, "--Tpath   ", "string", " SOAP path")
	printFlag(w, "--type      ", "string", "	 Vendor")

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Media:")
	printFlag(w, "--Lf		     ", "string", "	   Local media file or url (url is stream explicit)")
	printFlag(w, "--Lip		   ", "string", "	 Local IP")
	printFlag(w, "--Ldir		  ", "string", "  Local directory")
	printFlag(w, "--LPort		", "string", "Local port")

	w.Flush()
}

func printFlag(w *tabwriter.Writer, name, arg, desc string) {
	left := name
	if arg != "" {
		left = name + " " + arg
	}

	// EXACTLY two columns, separated by ONE tab
	fmt.Fprintf(w, "\t%s\t%s\n", left, desc)
}
