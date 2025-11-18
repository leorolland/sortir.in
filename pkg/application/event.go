package application

import "time"

type EventLocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
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

func (e Event) IsValid() bool {
	// Avoid events that are terminated
	if e.End.Before(time.Now()) {
		return false
	}

	// Avoid events that are too long to be saved
	if e.End.Sub(e.Begin) > time.Hour*24*15 { // 15 days
		return false
	}

	return true
}
