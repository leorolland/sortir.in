package application

import (
	"strings"
)

type Kind string

const (
	KindUnknown        Kind = "unknown" // Default kind
	KindConcert        Kind = "concert"
	KindTheater        Kind = "theater"
	KindMovie          Kind = "movie"
	KindFestival       Kind = "festival"
	KindParty          Kind = "party"
	KindKaraoke        Kind = "karaoke"
	KindBusiness       Kind = "business"
	KindFoodDrinks     Kind = "food-drinks"
	KindSports         Kind = "sports"
	KindExhibitions    Kind = "exhibitions"
	KindHealthWellness Kind = "health-wellness"
)

func KindFromString(s string) Kind {
	s = strings.ToLower(s)

	switch s {
	case "concert", "concerts":
		return KindConcert
	case "theater", "theaters":
		return KindTheater
	case "movie", "movies":
		return KindMovie
	case "festival", "festivals":
		return KindFestival
	case "party", "dance", "live-music", "parties":
		return KindParty
	case "karaoke":
		return KindKaraoke
	case "business", "meetups", "workshops":
		return KindBusiness
	case "food-drinks":
		return KindFoodDrinks
	case "sports":
		return KindSports
	case "exhibitions":
		return KindExhibitions
	case "health-wellness":
		return KindHealthWellness
	}
	return KindUnknown
}

func FirstKindMatch(s []string) Kind {
	for _, kindString := range s {
		kind := KindFromString(kindString)
		if kind != KindUnknown {
			return kind
		}
	}
	return KindUnknown
}
