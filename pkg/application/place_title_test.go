package application_test

import (
	"testing"

	"github.com/leorolland/sortir.in/pkg/application"
)

func TestShortPlaceTitle(t *testing.T) {
	testCases := map[string]struct {
		place    string
		expected string
	}{
		"empty": {
			place:    "",
			expected: "",
		},
		"whitespace only": {
			place:    "   ",
			expected: "",
		},
		"cinéma example": {
			place:    "Cinéma saint andré des prés",
			expected: "Saint André",
		},
		"hyphenated name drops the trailing phrase": {
			place:    "Cinéma Saint-André des Prés",
			expected: "Saint-André",
		},
		"librairie example keeps the article": {
			place:    "Librairie le delta",
			expected: "Le Delta",
		},
		"name without venue type": {
			place:    "Tour Eiffel",
			expected: "Tour Eiffel",
		},
		"brand name keeps its article": {
			place:    "Le Bon Marché",
			expected: "Le Bon Marché",
		},
		"accented venue type": {
			place:    "Théâtre du Rond-Point",
			expected: "Du Rond-Point",
		},
		"venue type only keeps the words": {
			place:    "Cinéma",
			expected: "Cinéma",
		},
		"venue type followed by stop words": {
			place:    "Bar de la gare",
			expected: "De la Gare",
		},
		"trailing stop words are dropped": {
			place:    "Cinéma Saint-André des",
			expected: "Saint-André",
		},
		"hyphenated words stay whole": {
			place:    "Marché Saint-André",
			expected: "Saint-André",
		},
		"apostrophes are capitalized": {
			place:    "l'Atelier des Lumières",
			expected: "L'Atelier",
		},
		"more than three words keeps the first three": {
			place:    "Café de la Gare du Nord",
			expected: "De la Gare",
		},
		"long single word is ellipsized": {
			place:    "Saint-André-des-Prés-sur-Loire-en-Berry",
			expected: "Saint-André-Des-Prés…",
		},
		"street address yields no label": {
			place:    "13 Rue Henri Barbusse",
			expected: "",
		},
		"street address without number yields no label": {
			place:    "Avenue Jean Jaurès",
			expected: "",
		},
		"generic word alone yields no label": {
			place:    "Place",
			expected: "",
		},
		"generic word with a name keeps the phrase": {
			place:    "Place du Marché",
			expected: "Place du Marché",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := application.ShortPlaceTitle(tc.place)
			if got != tc.expected {
				t.Fatalf("ShortPlaceTitle(%q) = %q, want %q", tc.place, got, tc.expected)
			}
		})
	}
}
