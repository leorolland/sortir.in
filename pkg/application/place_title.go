package application

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// ShortPlaceTitle reduces a venue name to its two or three most identifying
// words, used as the label displayed on the pin ("Cinéma Saint-André des
// Prés" → "Saint André", "Librairie Le Delta" → "Le Delta"). Empty input
// yields an empty label.
func ShortPlaceTitle(place string) string {
	words := strings.Fields(place)
	if len(words) == 0 {
		return ""
	}

	// Street addresses ("13 Rue Henri", "Avenue Jean Jaurès") locate the
	// pin but evoke no venue: no label at all.
	if isAddressLikePlace(place) {
		return ""
	}

	// Strip the leading words naming the venue type ("Cinéma", "Bar", ...)
	// rather than the place itself. A name made only of such words keeps them.
	i := 0
	for i < len(words) && isVenueTypeWord(words[i]) {
		i++
	}
	core := words[i:]
	if len(core) == 0 {
		core = words
	}
	full := core

	// Drop the trailing prepositional phrase ("... des Prés"): it locates
	// the place more than it identifies it ("Cinéma Saint-André des Prés"
	// → "Saint-André").
	for j := len(core) - 1; j >= 0; j-- {
		if isTrailingStopWord(core[j]) {
			core = core[:j]
			break
		}
	}

	// A drop that leaves a generic word ("Place du Marché" → "Place") loses
	// the name: keep the full phrase instead.
	if len(core) == 1 && (isVenueTypeWord(core[0]) || isGenericPlaceWord(core[0])) {
		core = full
	}

	// Keep at most three words, then drop any remaining trailing stop words.
	if len(core) > 3 {
		core = core[:3]
	}
	for len(core) > 0 && isTrailingStopWord(core[len(core)-1]) {
		core = core[:len(core)-1]
	}
	if len(core) == 0 {
		core = full
		if len(core) > 3 {
			core = core[:3]
		}
	}

	// A single generic word ("Place") evokes nothing: prefer no label.
	if len(core) == 1 && isGenericPlaceWord(core[0]) {
		return ""
	}

	// Bound the label: drop words while too long, ellipsize a long single word.
	for len(core) > 1 && len([]rune(strings.Join(core, " "))) > 24 {
		core = core[:len(core)-1]
	}
	label := titleCaseWords(core)
	if runes := []rune(label); len(runes) > 24 {
		label = strings.TrimRight(string(runes[:21]), "- ") + "…"
	}

	return label
}

// venueTypeWords are the leading words that name what a place is, not which
// one it is. Matched accent- and case-insensitively.
var venueTypeWords = map[string]bool{
	"arena": true, "arenes": true, "bar": true, "basilique": true, "bibliotheque": true,
	"boulangerie": true, "brasserie": true, "cafe": true, "cathedrale": true, "chateau": true,
	"cinema": true, "cirque": true, "eglise": true, "espace": true, "galerie": true,
	"gare": true, "halle": true, "halles": true, "hotel": true, "hyper": true,
	"hypermarket": true, "librairie": true, "maison": true, "mairie": true, "marche": true,
	"mediatheque": true, "museum": true, "musee": true, "parc": true, "pharmacie": true,
	"patisserie": true, "piscine": true, "restaurant": true, "salle": true, "stade": true,
	"supermarche": true, "supermarket": true, "synagogue": true, "temple": true, "theatre": true,
	"zoo": true,
}

// trailingStopWords are the small words that can end a venue name without
// identifying it. They are dropped from the end of the label and kept
// lowercase in the middle of it ("Château le Bois").
var trailingStopWords = map[string]bool{
	"a": true, "au": true, "aux": true, "d": true, "de": true, "des": true,
	"du": true, "en": true, "et": true, "la": true, "le": true, "les": true,
	"l": true, "of": true, "ou": true, "sur": true, "sous": true, "the": true,
}

// streetWords begin French street addresses ("13 Rue Henri"): such places
// locate the pin but make poor labels, so they get none.
var streetWords = map[string]bool{
	"allee": true, "avenue": true, "bd": true, "boulevard": true, "chemin": true,
	"impasse": true, "route": true, "rue": true, "voie": true, "zone": true,
}

// genericWords are place words too generic to stand alone as a label.
var genericWords = map[string]bool{
	"cour": true, "cours": true, "esplanade": true, "parvis": true,
	"place": true, "square": true,
}

func isAddressLikePlace(place string) bool {
	words := strings.Fields(place)
	if len(words) == 0 {
		return false
	}
	if unicode.IsDigit([]rune(words[0])[0]) {
		return true
	}
	return streetWords[foldWord(words[0])]
}

func isGenericPlaceWord(word string) bool {
	return genericWords[foldWord(word)]
}

func isVenueTypeWord(word string) bool {
	return venueTypeWords[foldWord(word)]
}

func isTrailingStopWord(word string) bool {
	return trailingStopWords[foldWord(word)]
}

// foldWord lowercases and strips accents, so "Cinéma" and "cinema" match.
func foldWord(word string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(strings.ToLower(word)) {
		if unicode.IsMark(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// titleCaseWords capitalizes each word (and letters after "-" and "'"),
// keeping stop words lowercase unless they start the label.
func titleCaseWords(words []string) string {
	out := make([]string, 0, len(words))
	for i, word := range words {
		if i > 0 && trailingStopWords[foldWord(word)] {
			out = append(out, strings.ToLower(word))
			continue
		}
		out = append(out, titleCaseWord(word))
	}
	return strings.Join(out, " ")
}

func titleCaseWord(word string) string {
	runes := []rune(word)
	capitalize := true
	for i, r := range runes {
		if capitalize && unicode.IsLetter(r) {
			runes[i] = unicode.ToUpper(r)
			capitalize = false
		} else if r == '-' || r == '\'' {
			capitalize = true
		}
	}
	return string(runes)
}
