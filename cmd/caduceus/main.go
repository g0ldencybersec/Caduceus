package main

import (
	"bufio"
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

	// If the input is '-', read from stdin
	if args.Input == "NONE" {
		var stdinIPs []string
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				stdinIPs = append(stdinIPs, line)
			}
		}

		// Handle any potential scanning errors
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}

		// Join all the IPs from stdin as a single comma-separated string
		args.Input = strings.Join(stdinIPs, ",")
	}

	args.Ports = strings.Split(args.PortList, ",")

	scrape.RunScrape(args)
}
