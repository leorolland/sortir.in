package collector

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
	"golang.org/x/net/html"
)

// Infomaniak Events (https://infomaniak.events) exposes two public JSON
// endpoints used by its map view:
//   - /api/map/events?map_t_lat&map_t_lng&map_b_lat&map_b_lng&zoom
//     returns venue-level pins [{lat, lon, count, places_count}] for a bbox.
//     Coordinates are stable (they match the venue's coordinates exactly).
//   - /api/map/events-at-point?lat&lon
//     returns {"html": "..."} with one microdata card per event happening at
//     those exact coordinates (schema.org Event subtypes).
//
// The portal is Switzerland-centric (plus the French Geneva border area), so
// most French cities yield no or few events.
const (
	infomaniakBaseURL = "https://infomaniak.events"
	infomaniakZoom    = "16"
	// The API is rate-limited to 60 requests per minute.
	infomaniakRequestInterval = 1050 * time.Millisecond
	infomaniakMaxRetries      = 3
)

type infomaniakCollector struct {
	client      *http.Client
	lastRequest time.Time
}

func NewInfomaniakCollector() application.Collector {
	return &infomaniakCollector{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type infomaniakPin struct {
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	Count      int     `json:"count"`
	PlacesCount int    `json:"places_count"`
}

type infomaniakCard struct {
	Type        string
	Name        string
	URL         string
	Img         string
	Description string
	Place       string
	Begin       time.Time
	End         time.Time
}

// throttle keeps the request rate below the API's limit.
func (c *infomaniakCollector) throttle() {
	if elapsed := time.Since(c.lastRequest); elapsed < infomaniakRequestInterval {
		time.Sleep(infomaniakRequestInterval - elapsed)
	}
	c.lastRequest = time.Now()
}

// fetchJSON performs a rate-limited GET and decodes the JSON response,
// retrying with backoff on server errors.
func (c *infomaniakCollector) fetchJSON(requestURL string, out any) error {
	var lastErr error
	for attempt := 0; attempt <= infomaniakMaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 3 * time.Second)
		}
		c.throttle()

		resp, err := c.client.Get(requestURL)
		if err != nil {
			lastErr = fmt.Errorf("error executing request: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("error decoding response: %w", err)
		}
		return nil
	}
	return lastErr
}

func (c *infomaniakCollector) Collect(location application.CollectLocation) ([]application.Event, error) {
	// Radius is in kilometers; convert it to a lat/lon bounding box.
	dLat := location.Radius / 111.32
	dLon := location.Radius / (111.32 * math.Cos(location.Lat*math.Pi/180))

	params := url.Values{}
	params.Set("map_t_lat", strconv.FormatFloat(location.Lat+dLat, 'f', -1, 64))
	params.Set("map_t_lng", strconv.FormatFloat(location.Lon+dLon, 'f', -1, 64))
	params.Set("map_b_lat", strconv.FormatFloat(location.Lat-dLat, 'f', -1, 64))
	params.Set("map_b_lng", strconv.FormatFloat(location.Lon-dLon, 'f', -1, 64))
	params.Set("zoom", infomaniakZoom)

	var pins []infomaniakPin
	if err := c.fetchJSON(infomaniakBaseURL+"/api/map/events?"+params.Encode(), &pins); err != nil {
		return nil, fmt.Errorf("error fetching pins: %w", err)
	}
	slog.Info("Fetched Infomaniak pins", "city", location.City, "pins", len(pins))

	now := time.Now()
	events := []application.Event{}
	for _, pin := range pins {
		cardHTML, err := c.fetchCardsHTML(pin.Lat, pin.Lon)
		if err != nil {
			slog.Warn("error fetching Infomaniak events at point, skipping pin", "lat", pin.Lat, "lon", pin.Lon, "error", err)
			continue
		}
		events = append(events, infomaniakCardsToEvents(parseInfomaniakCards(cardHTML), application.EventLocation{Lat: pin.Lat, Lon: pin.Lon}, now)...)
	}

	return dedupeInfomaniakEvents(events), nil
}

func (c *infomaniakCollector) fetchCardsHTML(lat, lon float64) (string, error) {
	params := url.Values{}
	params.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	params.Set("lon", strconv.FormatFloat(lon, 'f', -1, 64))

	var response struct {
		HTML string `json:"html"`
	}
	if err := c.fetchJSON(infomaniakBaseURL+"/api/map/events-at-point?"+params.Encode(), &response); err != nil {
		return "", err
	}
	return response.HTML, nil
}

// parseInfomaniakCards extracts the microdata event cards from the carousel
// HTML returned by /api/map/events-at-point.
func parseInfomaniakCards(cardHTML string) []infomaniakCard {
	if strings.TrimSpace(cardHTML) == "" {
		return nil
	}

	doc, err := html.Parse(strings.NewReader(cardHTML))
	if err != nil {
		slog.Warn("error parsing Infomaniak cards HTML", "error", err)
		return nil
	}

	cards := []infomaniakCard{}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if card, ok := infomaniakCardFromNode(n); ok {
				cards = append(cards, card)
				return
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return cards
}

func infomaniakCardFromNode(n *html.Node) (infomaniakCard, bool) {
	itemType := ""
	for _, attr := range n.Attr {
		if attr.Key == "itemtype" {
			itemType = attr.Val
		}
	}
	if !strings.Contains(itemType, "schema.org/") || !strings.HasSuffix(itemType, "Event") {
		return infomaniakCard{}, false
	}

	card := infomaniakCard{Type: itemType[strings.LastIndex(itemType, "/")+1:]}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			var itemprop string
			for _, attr := range n.Attr {
				if attr.Key == "itemprop" {
					itemprop = attr.Val
				}
			}
			switch itemprop {
			case "name":
				card.Name = strings.TrimSpace(nodeText(n))
			case "url":
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						card.URL = attr.Val
					}
				}
			case "image":
				for _, attr := range n.Attr {
					if attr.Key == "data-src" || (attr.Key == "src" && card.Img == "") {
						card.Img = attr.Val
					}
				}
			case "startDate":
				card.Begin = parseInfomaniakDate(contentOrText(n))
			case "endDate":
				card.End = parseInfomaniakDate(contentOrText(n))
			case "description":
				card.Description = strings.TrimSpace(contentOrText(n))
			case "address":
				card.Place = strings.TrimSpace(nodeText(n))
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(n)

	if card.Name == "" || card.URL == "" || card.Begin.IsZero() {
		return infomaniakCard{}, false
	}
	return card, true
}

func nodeText(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(n)
	return sb.String()
}

func contentOrText(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "content" {
			return attr.Val
		}
	}
	return nodeText(n)
}

func parseInfomaniakDate(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		slog.Warn("error parsing Infomaniak date", "date", s, "error", err)
		return time.Time{}
	}
	return t
}

// infomaniakSchemaKinds maps schema.org event subtypes to application kinds.
var infomaniakSchemaKinds = map[string]application.Kind{
	"MusicEvent":      application.KindConcert,
	"TheaterEvent":    application.KindTheater,
	"ComedyEvent":     application.KindTheater,
	"ScreeningEvent":  application.KindMovie,
	"CinemaEvent":     application.KindMovie,
	"VisualArtsEvent": application.KindExhibitions,
	"ExhibitionEvent": application.KindExhibitions,
	"ConferenceEvent": application.KindExhibitions,
	"BusinessEvent":   application.KindBusiness,
	"FoodEvent":       application.KindFoodDrinks,
	"DrinkEvent":      application.KindFoodDrinks,
	"DanceEvent":      application.KindParty,
	"SportsEvent":     application.KindSports,
	"EducationEvent":  application.KindWorkshop,
}

// infomaniakSlugKinds maps the event URL category slug to application kinds,
// used when the schema.org subtype is too generic (plain "Event").
var infomaniakSlugKinds = map[string]application.Kind{
	"concerts":           application.KindConcert,
	"musique":            application.KindConcert,
	"musique-classique":  application.KindConcert,
	"theatre":            application.KindTheater,
	"humour":             application.KindTheater,
	"humour-et-comedie":  application.KindTheater,
	"cinema":             application.KindMovie,
	"festivals":          application.KindFestival,
	"festival":           application.KindFestival,
	"sport":              application.KindSports,
	"conferences":        application.KindExhibitions,
	"famille":            application.KindWorkshop,
	"ateliers-et-stages": application.KindWorkshop,
	"balades-et-visites": application.KindWorkshop,
	"danse":              application.KindParty,
}

// infomaniakEventSlug extracts the category slug from an event URL like
// https://infomaniak.events/fr-ch/concerts/suzane/<uuid>/events/390180.
func infomaniakEventSlug(eventURL string) string {
	parts := strings.Split(strings.TrimPrefix(eventURL, "https://infomaniak.events/"), "/")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func infomaniakCardToEvent(card infomaniakCard, loc application.EventLocation, now time.Time) (application.Event, bool) {
	end := card.End
	if end.IsZero() {
		end = card.Begin
	}

	// Skip events that already ended (the server would reject them anyway).
	if end.Before(now) {
		return application.Event{}, false
	}

	// Skip recurring-event series spanning weeks (the server rejects
	// events longer than 15 days).
	if end.Sub(card.Begin) > time.Hour*24*15 {
		return application.Event{}, false
	}

	slug := infomaniakEventSlug(card.URL)
	genres := []string{}
	if slug != "" {
		genres = append(genres, slug)
	}

	kind := infomaniakSchemaKinds[card.Type]
	if kind == "" {
		kind = infomaniakSlugKinds[slug]
	}
	if kind == "" {
		kind = application.FirstKindMatch(genres)
	}

	return application.Event{
		Name:   card.Name,
		Kind:   kind,
		Genres: genres,
		Begin:  card.Begin,
		End:    end,
		Loc:    loc,
		Place:  card.Place,
		Source: card.URL,
		Img:    card.Img,
	}, true
}

func infomaniakCardsToEvents(cards []infomaniakCard, loc application.EventLocation, now time.Time) []application.Event {
	events := []application.Event{}
	for _, card := range cards {
		event, ok := infomaniakCardToEvent(card, loc, now)
		if ok {
			events = append(events, event)
		}
	}
	return events
}

func dedupeInfomaniakEvents(events []application.Event) []application.Event {
	seen := make(map[string]bool)
	unique := make([]application.Event, 0, len(events))
	for _, event := range events {
		if seen[event.Source] {
			continue
		}
		seen[event.Source] = true
		unique = append(unique, event)
	}
	return unique
}
