package agonet

import "time"

// // Option is a function that will set up option.
// type Option func(opts *Options)

// func loadOptions(options ...Option) *Options {
// 	opts := new(Options)
// 	for _, option := range options {
// 		option(opts)
// 	}
// 	return opts
// }

// Options are configurations for the gnet application.
type Options struct {
	Multicore bool

	NumEventLoop int

	// LockOSThread bool

	Ticker bool

	KeepAlive struct {
		Enable   bool
		Idle     time.Duration
		Interval time.Duration
		Count    int
	}
}
