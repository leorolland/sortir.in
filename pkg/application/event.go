package application

import "time"

type EventLocation struct {
	Lat float64
	Lon float64
}

type Event struct {
	Name          string
	Kind          string
	Genres        []string
	Begin         time.Time
	End           time.Time
	Loc           EventLocation
	Place         string
	Address       string
	Price         *float64
	PriceCurrency *string
	Source        string
	Img           string
}
