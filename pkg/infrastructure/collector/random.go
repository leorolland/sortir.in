package collector

import (
	"math/rand/v2"
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
)

type randomCollector struct {
}

func NewRandomCollector() application.Collector {
	return &randomCollector{}
}

func (c *randomCollector) Collect(location application.CollectLocation) ([]application.Event, error) {
	return []application.Event{
		{
			Name: "Random Event 1",
			Lat:  location.Lat + rand.Float64() * 0.1,
			Lon:  location.Lon + rand.Float64() * 0.1,
			Begin: time.Now().Add(time.Hour * 24),
			End:   time.Now().Add(time.Hour * 24 * 2),
		},
		{
			Name: "Random Event 2",
			Lat:  location.Lat + rand.Float64() * 0.1,
			Lon:  location.Lon + rand.Float64() * 0.1,
			Begin: time.Now().Add(time.Hour * 24 * 2),
			End:   time.Now().Add(time.Hour * 24 * 3),
		},
		{
			Name: "Random Event 3",
			Lat:  location.Lat + rand.Float64() * 0.1,
			Lon:  location.Lon + rand.Float64() * 0.1,
			Begin: time.Now().Add(time.Hour * 24 * 3),
			End:   time.Now().Add(time.Hour * 24 * 4),
		},
	}, nil
}
