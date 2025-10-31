package application

import "time"

type Event struct {
	Name string
	Lat  float64
	Lon  float64
	Begin time.Time
	End   time.Time
}
