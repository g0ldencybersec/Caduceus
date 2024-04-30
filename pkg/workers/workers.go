package workers

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/g0ldencybersec/Caduceus/pkg/types"
	"github.com/g0ldencybersec/Caduceus/pkg/utils"
)

// Worker Types
type worker struct {
	dialer  *net.Dialer
	input   <-chan string
	results chan<- types.Result
}

type WorkerPool struct {
	workers []*worker
	input   chan string
	results chan types.Result
	dialer  *net.Dialer
}

func NewWorker(dialer *net.Dialer, input <-chan string, results chan<- types.Result) *worker {
	return &worker{
		dialer:  dialer,
		input:   input,
		results: results,
	}
}

func (w *worker) run() {
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

		names := utils.ExtractNames(cert)
		org := cert.Subject.Organization

		certInfo := types.CertificateInfo{
			IP:           ip,
			Organization: utils.GetOrganization(org),
			CommonName:   names[0],
			SAN:          utils.JoinNonEmpty(", ", names[1:]),
			Domains:      names,
		}

		w.results <- types.Result{IP: ip, Hit: true, Certificate: &certInfo, Timeout: false}
	}
}

func NewWorkerPool(size int, dialer *net.Dialer, input chan string, results chan types.Result) *WorkerPool {
	wp := &WorkerPool{
		workers: make([]*worker, size),
		input:   input,
		results: results,
		dialer:  dialer,
	}
	for i := range wp.workers {
		wp.workers[i] = NewWorker(wp.dialer, wp.input, wp.results)
	}
	return wp
}

func (wp *WorkerPool) Start() {
	for _, worker := range wp.workers {
		go worker.run()
	}
}

func (wp *WorkerPool) Stop() {
	close(wp.input)
}

// Result Workers
type ResultWorker struct {
	resultInput   <-chan types.Result
	outputChannel chan<- string
}

type ResultWorkerPool struct {
	workers       []*ResultWorker
	resultInput   chan types.Result
	outputChannel chan string
}

func NewResultWorker(resultInput <-chan types.Result, outputChannel chan<- string) *ResultWorker {
	return &ResultWorker{
		resultInput:   resultInput,
		outputChannel: outputChannel,
	}
}

func (rw *ResultWorker) Run(args types.ScrapeArgs) {
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

func NewResultWorkerPool(size int, resultInput chan types.Result, outputChannel chan string) *ResultWorkerPool {
	rwp := &ResultWorkerPool{
		workers:       make([]*ResultWorker, size),
		resultInput:   resultInput,
		outputChannel: outputChannel,
	}
	for i := range rwp.workers {
		rwp.workers[i] = NewResultWorker(rwp.resultInput, rwp.outputChannel)
	}
	return rwp
}

func (rwp *ResultWorkerPool) Start(args types.ScrapeArgs) {
	for _, worker := range rwp.workers {
		go worker.Run(args)
	}
}

func (rwp *ResultWorkerPool) Stop() {
	close(rwp.resultInput)
}
