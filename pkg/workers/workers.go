package workers

import (
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
