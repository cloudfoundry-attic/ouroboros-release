package app

import (
	"conf"
	"math/rand"
	"time"
)

type Killer struct {
	killDelay conf.DurationRange
	kill      func()
}

// NewKiller calls a function after a random delay
func NewKiller(killDelay conf.DurationRange, kill func()) *Killer {
	return &Killer{
		killDelay: killDelay,
		kill:      kill,
	}
}

func (k *Killer) Start() {
	k.killAfterRandomDelay()
}

func (v *Killer) killAfterRandomDelay() {
	delta := int(v.killDelay.Max - v.killDelay.Min)
	killDelay := v.killDelay.Min + time.Duration(rand.Intn(delta))
	time.AfterFunc(killDelay, v.kill)
}
