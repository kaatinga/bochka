package bochka

import "time"

type options struct {
	timeout time.Duration
	port    string
}

type option func(*options)

func getOptions(opts []option) (opt options) {
	opt.timeout = 30 * time.Second
	for _, o := range opts {
		o(&opt)
	}
	return
}

func WithTimeout(timeout time.Duration) option {
	return func(opt *options) {
		opt.timeout = timeout
	}
}

func WithPort(port string) option {
	return func(opt *options) {
		opt.port = port
	}
}
