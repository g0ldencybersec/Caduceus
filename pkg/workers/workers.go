package workers

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/g0ldencybersec/Caduceus/pkg/types"
	"github.com/g0ldencybersec/Caduceus/pkg/utils"
)

// Worker Types
type Worker struct {
	dialer  *net.Dialer
	input   <-chan string
	results chan<- types.Result
}

type WorkerPool struct {
	workers []*Worker
	input   chan string
	results chan types.Result
	dialer  *net.Dialer
	wg      sync.WaitGroup
}

func NewWorker(dialer *net.Dialer, input <-chan string, results chan<- types.Result) *Worker {
	return &Worker{
		dialer:  dialer,
		input:   input,
		results: results,
	}
}

func (w *Worker) run() {
	for ip := range w.input {
		cert, err := utils.GetSSLCert(ip, w.dialer)
		if err != nil {
			netErr, ok := err.(net.Error) // Type assertion to check if it is a net.Error
			if ok && netErr.Timeout() {
				// Specific handling for timeout
				w.results <- types.Result{IP: ip, Hit: false, Timeout: true}
			} else {
				// General error handling
				w.results <- types.Result{IP: ip, Error: err, Hit: false, Timeout: false}
			}
			continue
		}

		certInfo := types.CertificateInfo{
			OriginIP:         ip,
			Organization:     cert.Subject.Organization,
			OrganizationUnit: cert.Subject.OrganizationalUnit,
			CommonName:       cert.Subject.CommonName,
			SAN:              cert.DNSNames,
			Domains:          append([]string{cert.Subject.CommonName}, cert.DNSNames...),
			Emails:           cert.EmailAddresses,
			IPAddrs:          cert.IPAddresses,
		}

		w.results <- types.Result{IP: ip, Hit: true, Certificate: &certInfo, Timeout: false}
	}
}

func NewWorkerPool(size int, dialer *net.Dialer, input chan string, results chan types.Result) *WorkerPool {
	wp := &WorkerPool{
		workers: make([]*Worker, size),
		input:   input,
		results: results,
		dialer:  dialer,
		wg:      sync.WaitGroup{},
	}
	for i := range wp.workers {
		wp.workers[i] = NewWorker(wp.dialer, wp.input, wp.results)
	}
	return wp
}

func (wp *WorkerPool) Start() {
	for _, worker := range wp.workers {
		wp.wg.Add(1) // Properly adding to the waitgroup before the goroutine
		go func(w *Worker) {
			defer wp.wg.Done()
			w.run()
		}(worker) // Correctly passing the loop variable as a parameter
	}
}

func (wp *WorkerPool) Stop() {
	// Wait for all workers to finish their tasks.
	// This blocks until all workers have called wg.Done(),
	// signaling that they have completed.
	wp.wg.Wait()

	// Optionally, if the WorkerPool also manages the results channel, close it:
	// Make sure no more writes to the results channel are expected by this point.
	close(wp.results)
}

// Result Workers
type ResultsWorker struct {
	resultInput   <-chan types.Result
	outputChannel chan<- string
}

type ResultsWorkerPool struct {
	workers       []*ResultsWorker
	resultInput   chan types.Result // Note: This channel is closed by WorkerPool
	outputChannel chan string
	wg            sync.WaitGroup
}

func NewResultsWorker(resultInput <-chan types.Result, outputChannel chan<- string) *ResultsWorker {
	return &ResultsWorker{
		resultInput:   resultInput,
		outputChannel: outputChannel,
	}
}

func (rw *ResultsWorker) Run(args types.ScrapeArgs) {
	for result := range rw.resultInput {
		if result.Hit {
			if args.JsonOutput {
				outputJSON, _ := json.Marshal(result.Certificate)
				rw.outputChannel <- string(outputJSON)
			} else {
				for _, domain := range result.Certificate.Domains {
					if args.PrintWildcards {
						if utils.IsWilcard(domain) || utils.IsValidDomain(domain) {
							rw.outputChannel <- domain
							continue
						}
					}
					if utils.IsValidDomain(domain) && !utils.IsWilcard(domain) {
						rw.outputChannel <- domain
					}
				}
			}
		} else if args.Debug {
			if result.Timeout {
				rw.outputChannel <- fmt.Sprintf("Timed Out. No SSL certificate found for %s", result.IP)
			}
			if result.Error != nil {
				rw.outputChannel <- fmt.Sprintf("Failed to get SSL certificate from %s: %v", result.IP, result.Error)
			}
		}
	}
}

func NewResultWorkerPool(size int, resultInput chan types.Result, outputChannel chan string) *ResultsWorkerPool {
	rwp := &ResultsWorkerPool{
		workers:       make([]*ResultsWorker, size),
		resultInput:   resultInput,
		outputChannel: outputChannel,
		wg:            sync.WaitGroup{},
	}
	for i := range rwp.workers {
		rwp.workers[i] = NewResultsWorker(rwp.resultInput, rwp.outputChannel)
	}
	return rwp
}

func (rwp *ResultsWorkerPool) Start(args types.ScrapeArgs) {
	for _, worker := range rwp.workers {
		rwp.wg.Add(1)
		go func(rw *ResultsWorker) {
			defer rwp.wg.Done()
			rw.Run(args)
		}(worker)
	}
}

func (rwp *ResultsWorkerPool) Stop() {
	// Wait for all results workers to finish their tasks
	rwp.wg.Wait()

	// Since the output channel might be managed here, consider closing it if appropriate
	close(rwp.outputChannel)
}
