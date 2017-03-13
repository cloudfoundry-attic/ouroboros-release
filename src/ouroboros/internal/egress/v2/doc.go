package egress

import (
	loggregator "loggregator/v2"
)

//go:generate hel

type IngressServer interface {
	loggregator.IngressServer
}
