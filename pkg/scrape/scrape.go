package scrape

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"time"

	"github.com/g0ldencybersec/Caduceus/pkg/stats"
	"github.com/g0ldencybersec/Caduceus/pkg/types"
	"github.com/g0ldencybersec/Caduceus/pkg/utils"
	"github.com/g0ldencybersec/Caduceus/pkg/workers"
)

func RunScrape(args types.ScrapeArgs) {
	dialer := &net.Dialer{
		Timeout: time.Duration(args.Timeout) * time.Second,
	}

	inputChannel := make(chan string)
	resultChannel := make(chan types.Result)

	stats := &stats.Stats{}

	workerPool := workers.NewWorkerPool(args.Concurrency, dialer, inputChannel, resultChannel)
	workerPool.Start()

	go utils.IntakeFunction(inputChannel, args.Ports, args.Input)

	defer func() {
		workerPool.Stop()
		close(resultChannel)
	}()

	file, err := os.Create(args.OutputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output file: %v\n", err)
		return
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for result := range resultChannel {

		stats.Update(result)
		stats.Display() // Display updated stats

		if result.Hit {
			outputJSON, _ := json.Marshal(result.Certificate)
			writer.Write(outputJSON)
			writer.WriteString("\n")
		} else if args.Debug {
			if result.Timeout {
				fmt.Printf("Timed Out. No SSL certificate found for %s\n", result.IP)
			}
			if result.Error != nil {
				fmt.Printf("Failed to get SSL certificate from %s: %v\n", result.IP, result.Error)
			}
		}
	}

}
