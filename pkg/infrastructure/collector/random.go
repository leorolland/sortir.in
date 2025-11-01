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
			Name:   "Random Event 1",
			Kind:   "Random Kind 1",
			Genres: []string{"Random Genre 1"},
			Begin:  time.Now().Add(time.Hour * 24),
			End:    time.Now().Add(time.Hour * 24 * 2),
			Loc: application.EventLocation{
				Lat: location.Lat + rand.Float64()*0.1,
				Lon: location.Lon + rand.Float64()*0.1,
			},
			Place:   "Random Place 1",
			Address: "Random Address 1",
			Price:   nil,
			Source:  "Random Source 1",
			Img:     "Random Img 1",
		},
		{
			Name:   "Random Event 2",
			Kind:   "Random Kind 2",
			Genres: []string{"Random Genre 2"},
			Begin:  time.Now().Add(time.Hour * 24 * 2),
			End:    time.Now().Add(time.Hour * 24 * 3),
			Loc: application.EventLocation{
				Lat: location.Lat + rand.Float64()*0.1,
				Lon: location.Lon + rand.Float64()*0.1,
			},
			Place:   "Random Place 2",
			Address: "Random Address 2",
			Price:   nil,
			Source:  "Random Source 2",
			Img:     "Random Img 2",
		},
		{
			Name:   "Random Event 3",
			Kind:   "Random Kind 3",
			Genres: []string{"Random Genre 3"},
			Begin:  time.Now().Add(time.Hour * 24 * 3),
			End:    time.Now().Add(time.Hour * 24 * 4),
			Loc: application.EventLocation{
				Lat: location.Lat + rand.Float64()*0.1,
				Lon: location.Lon + rand.Float64()*0.1,
			},
			Place:   "Random Place 3",
			Address: "Random Address 3",
			Price:   nil,
			Source:  "Random Source 3",
			Img:     "Random Img 3",
		},
	}, nil
}
