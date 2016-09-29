package conf

import (
	"fmt"
	"strings"
	"time"
)

type DurationRange struct {
	Min, Max time.Duration
}

func (d *DurationRange) UnmarshalEnv(v string) error {
	values := strings.Split(v, "-")
	if len(values) != 2 {
		return fmt.Errorf("Expected DurationRange to be of format {min}-{max}")
	}
	var err error
	d.Min, err = time.ParseDuration(values[0])
	if err != nil {
		return fmt.Errorf("Error parsing DurationRange.Min: %s", err)
	}
	d.Max, err = time.ParseDuration(values[1])
	if err != nil {
		return fmt.Errorf("Error parsing DurationRange.Max: %s", err)
	}
	return nil
}
