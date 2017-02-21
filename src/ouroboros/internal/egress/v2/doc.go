package egress

import (
	loggregator "ouroboros/internal/loggregator/v2"
)

//go:generate hel

type IngressServer interface {
	loggregator.IngressServer
}
