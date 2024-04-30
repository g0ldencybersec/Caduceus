package scrape

import (
	"fmt"
	"net"
	"sync"

	"time"

	"github.com/g0ldencybersec/Caduceus/pkg/types"
	"github.com/g0ldencybersec/Caduceus/pkg/utils"
	"github.com/g0ldencybersec/Caduceus/pkg/workers"
)

func RunScrape(args types.ScrapeArgs) {
	var wg sync.WaitGroup

	dialer := &net.Dialer{
		Timeout: time.Duration(args.Timeout) * time.Second,
	}

	inputChannel := make(chan string)
	resultChannel := make(chan types.Result)
	outputChannel := make(chan string, 1000)

	workerPool := workers.NewWorkerPool(args.Concurrency, dialer, inputChannel, resultChannel)
	resultWorkerPool := workers.NewResultWorkerPool(10, resultChannel, outputChannel)

	// Start worker pools
	workerPool.Start()
	resultWorkerPool.Start(args)

	// Handle input feeding
	wg.Add(1)
	go func() {
		defer wg.Done()
		utils.IntakeFunction(inputChannel, args.Ports, args.Input)
	}()

	// Handle outputs
	go func() {
		for output := range outputChannel {
			fmt.Println(output)
		}
	}()

	// Wait for all inputs to be processed before closing inputChannel
	wg.Wait()
	close(inputChannel)
	workerPool.Stop() // Wait internally for all workers to finish before closing resultChannel
	close(resultChannel)
	resultWorkerPool.Stop() // Wait internally for all result workers to finish before closing outputChannel
	close(outputChannel)

	// if args.PrintStats {
	// 	stats.Display() // Display updated stats
	// }

}
