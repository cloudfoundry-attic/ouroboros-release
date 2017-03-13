package converter

import (
	v2 "loggregator/v2"

	"github.com/cloudfoundry/sonde-go/events"
)

type Converter struct{}

func NewConverter() Converter {
	return Converter{}
}

func (c Converter) ToV1(e *v2.Envelope) *events.Envelope {
	return ToV1(e)
}

func (c Converter) ToV2(e *events.Envelope) *v2.Envelope {
	return ToV2(e)
}
