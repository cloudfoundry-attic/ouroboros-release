package egress

import (
	loggregator "code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

//go:generate hel

type IngressServer interface {
	loggregator.IngressServer
}
