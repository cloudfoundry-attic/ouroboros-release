package ranger

import (
	"errors"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Ranger struct {
	min, max  time.Duration
	randRange int
}

func New(min, max time.Duration) (*Ranger, error) {
	if min >= max-1 {
		return nil, errors.New("ranger: min must be at least two less than max")
	}
	return &Ranger{
		min:       min,
		max:       max,
		randRange: int(max - min),
	}, nil
}

func (r *Ranger) DelayRange() (min, max time.Duration) {
	rMin := r.min + time.Duration(rand.Intn(r.randRange))
	rMax := r.min + time.Duration(rand.Intn(r.randRange))
	switch {
	case rMin > rMax:
		rMin, rMax = rMax, rMin
	case rMin == rMax:
		rMin, rMax = r.separate(rMin)
	}
	return rMin, rMax
}

func (r *Ranger) separate(delay time.Duration) (min, max time.Duration) {
	if delay > r.min {
		return delay - 1, delay
	}
	return delay, delay + 1
}
