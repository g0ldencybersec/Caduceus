package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/g0ldencybersec/Caduceus/pkg/scrape"
	"github.com/g0ldencybersec/Caduceus/pkg/types"
)

func main() {
	args := types.ScrapeArgs{}
	scrapeUsage := "-i <IPs/CIDRs or File> "

	flag.IntVar(&args.Concurrency, "c", 100, "How many goroutines running concurrently")
	flag.StringVar(&args.PortList, "p", "443", "TLS ports to check for certificates")
	flag.IntVar(&args.Timeout, "t", 4, "Timeout for TLS handshake")
	flag.StringVar(&args.Input, "i", "NONE", "Either IPs & CIDRs separated by commas, or a file with IPs/CIDRs on each line")
	flag.BoolVar(&args.Debug, "debug", false, "Add this flag if you want to see failures/timeouts")
	flag.BoolVar(&args.Help, "h", false, "Show the program usage message")
	flag.BoolVar(&args.JsonOutput, "j", false, "print cert data as jsonl")
	flag.BoolVar(&args.PrintWildcards, "wc", false, "print wildcards to stdout")
	//flag.BoolVar(&args.Help, "stats", false, "Print stats at the end")

	flag.Parse()

	//need at least 100
	if args.Concurrency < 100 {
		args.Concurrency = 100
	}

	if args.Help {
		fmt.Println(scrapeUsage)
		flag.PrintDefaults()
		return
	}

	if args.Input == "NONE" {
		fmt.Print("No input detected, please use the -i flag to add input!\n\n")
		fmt.Println(scrapeUsage)
		flag.PrintDefaults()
		os.Exit(1)
	}

	args.Ports = strings.Split(args.PortList, ",")

	scrape.RunScrape(args)
}
