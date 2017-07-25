package converter

import (
	v2 "loggregator/v2"

	"github.com/cloudfoundry/sonde-go/events"
)

type Converter struct {
	usePreferredTags bool
}

func NewConverter(usePreferredTags bool) Converter {
	return Converter{
		usePreferredTags: usePreferredTags,
	}
}

func (c Converter) ToV1(e *v2.Envelope) *events.Envelope {
	return ToV1(e)[0]
}

func (c Converter) ToV2(e *events.Envelope) *v2.Envelope {
	return ToV2(e, c.usePreferredTags)
}
