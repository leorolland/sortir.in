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
	KindCircus         Kind = "circus"
	KindWorkshop       Kind = "workshop"
	KindFleaMarket     Kind = "flea-market"
	KindSolidarity     Kind = "solidarity"
)

func KindFromString(s string) Kind {
	s = strings.ToLower(s)

	switch s {
	case "movie", "movies", "ecrans":
		return KindMovie
	case "concert", "concerts", "spectacle musical":
		return KindConcert
	case "festival", "festivals":
		return KindFestival
	case "theater", "theaters", "théâtre", "humour":
		return KindTheater
	case "solidarité":
		return KindSolidarity
	case "party", "dance", "live-music", "parties", "danse":
		return KindParty
	case "karaoke":
		return KindKaraoke
	case "business", "meetups", "workshops":
		return KindBusiness
	case "food-drinks", "gourmand":
		return KindFoodDrinks
	case "sports", "sport":
		return KindSports
	case "exhibitions", "expo", "conférence", "salon", "art contemporain":
		return KindExhibitions
	case "health-wellness":
		return KindHealthWellness
	case "cirque":
		return KindCircus
	case "workshop", "atelier", "littérature", "enfants", "loisirs", "nature":
		return KindWorkshop
	case "marché", "brocante":
		return KindFleaMarket
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
