package main

import (
	"errors"
)

var (
	ErrOperation = errors.New("request operation not valid")
	// ring
	ErrRingEmpty = errors.New("ring buffer empty")
	ErrRingFull  = errors.New("ring buffer full")
)
