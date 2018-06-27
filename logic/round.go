package main

import (
	"im/libs/cityhash"
	itime "im/libs/time"
)

type RoundOptions struct {
	Timer     int
	TimerSize int
}

type Round struct {
	timers  []itime.Timer
	options RoundOptions
}

func NewRound(options RoundOptions) (r *Round) {
	var i int
	r = new(Round)
	r.options = options

	r.timers = make([]itime.Timer, options.Timer)
	for i = 0; i < options.Timer; i++ {
		r.timers[i].Init(options.TimerSize)
	}
	return
}

func (r *Round) Timer(rn int) *itime.Timer {
	return &(r.timers[rn%r.options.Timer])
}

func (r *Round) TimerWithKey(subKey string) *itime.Timer {
	value := uint(cityhash.CityHash64([]byte(subKey), uint32(len(subKey))))
	return &(r.timers[value%uint(r.options.Timer)])
}
