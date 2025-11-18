package collector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/leorolland/sortir.in/pkg/application"
	"golang.org/x/text/unicode/norm"
)

type bobineCollector struct {
	client *http.Client
}

func NewBobineCollector() application.Collector {
	return &bobineCollector{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func removeAccents(s string) string {
	t := norm.NFD.String(s)
	result := make([]rune, 0, len(t))
	for _, r := range t {
		if unicode.IsMark(r) {
			continue
		}
		result = append(result, r)
	}
	return string(result)
}

type bobineMovie struct {
	ID          int     `json:"id"`
	TitleVO     string  `json:"title_vo"`
	TitleVF     string  `json:"title_vf"`
	Duration    int     `json:"duration"`
	Synopsis    string  `json:"synopsis"`
	PosterPath  string  `json:"poster_path"`
	StillPath   string  `json:"still_path"`
	Director    string  `json:"director"`
	Casting     string  `json:"casting"`
	Genres      *string `json:"genres"`
	MainLang    string  `json:"main_lang"`
	ForChildren bool    `json:"for_children"`
}

func (m *bobineMovie) GetTitle() string {
	if m.TitleVF != "" {
		return m.TitleVF
	}
	return m.TitleVO
}

func (m *bobineMovie) GetURL() string {
	titleVF := strings.ToLower(strings.ReplaceAll(m.TitleVF, " ", "-"))
	titleVF = removeAccents(titleVF)
	return fmt.Sprintf("https://bobine.art/film/%s-%d", m.TitleVO, m.ID)
}

func (m *bobineMovie) GetGenres() []string {
	if m.Genres == nil {
		return []string{"movie"}
	}
	return strings.Split(*m.Genres, ", ")
}

type bobineShowtime struct {
	ID        int       `json:"id"`
	Showtime  time.Time `json:"showtime"`
	AudioLang string    `json:"audio_lang"`
	ExtraInfo *string   `json:"extra_info"`
	EventInfo *string   `json:"event_info"`
}

type bobineTheater struct {
	ID        int              `json:"id"`
	Name      string           `json:"name"`
	Address   string           `json:"address"`
	Distance  int              `json:"distance"`
	Latitude  float64          `json:"latitude"`
	Longitude float64          `json:"longitude"`
	FullPrice float64          `json:"full_price"`
	Showtimes []bobineShowtime `json:"showtimes"`
}

func (t *bobineTheater) GetPriceCurrency() *string {
	if t.FullPrice > 0 {
		currency := "EUR"
		return &currency
	}
	return nil
}

type bobineResponse struct {
	Movie    bobineMovie     `json:"movie"`
	Theaters []bobineTheater `json:"theaters"`
}

func (c *bobineCollector) Collect(location application.CollectLocation) ([]application.Event, error) {
	page := 1
	pageSize := 20
	allEvents := []application.Event{}

	for {
		// Search for showtimes in the next 2 days
		startDate := time.Now().UTC().Format("2006-01-02T15:04:05Z")
		endDate := time.Now().AddDate(0, 0, 2).UTC().Format("2006-01-02T15:04:05Z")

		// Build the URL with query parameters
		baseURL := "https://bobine.art/api/showtimes/search"
		params := url.Values{}
		params.Add("range", "10")
		params.Add("order_by", "next_showtime")
		params.Add("order_dir", "asc")
		params.Add("latitude", fmt.Sprintf("%.7f", location.Lat))
		params.Add("longitude", fmt.Sprintf("%.7f", location.Lon))
		params.Add("start", startDate)
		params.Add("end", endDate)
		params.Add("page", fmt.Sprintf("%d", page))
		params.Add("page_size", fmt.Sprintf("%d", pageSize))

		requestURL := baseURL + "?" + params.Encode()

		// Create and execute the request
		req, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %w", err)
		}

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error executing request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		var bobineResp []bobineResponse
		if err := json.NewDecoder(resp.Body).Decode(&bobineResp); err != nil {
			return nil, fmt.Errorf("error decoding response: %w", err)
		}

		// If the response is empty, we've reached the end of the pages
		if len(bobineResp) == 0 {
			break
		}

		// Convert the response to events
		events, err := toBobineEvents(bobineResp)
		if err != nil {
			return nil, err
		}

		allEvents = append(allEvents, events...)
		page++
	}

	return allEvents, nil
}

func toBobineEvents(bobineResp []bobineResponse) ([]application.Event, error) {
	events := []application.Event{}

	for _, movieData := range bobineResp {
		movie := movieData.Movie

		for _, theater := range movieData.Theaters {
			for _, showtime := range theater.Showtimes {
				endTime := showtime.Showtime.Add(time.Duration(movie.Duration) * time.Minute)
				price := theater.FullPrice

				event := application.Event{
					Name:   movie.GetTitle(),
					Kind:   "movie",
					Genres: movie.GetGenres(),
					Begin:  showtime.Showtime,
					End:    endTime,
					Loc: application.EventLocation{
						Lat: theater.Latitude,
						Lon: theater.Longitude,
					},
					Place:         theater.Name,
					Address:       theater.Address,
					Price:         &price,
					PriceCurrency: theater.GetPriceCurrency(),
					Source:        movie.GetURL(),
					Img:           movie.PosterPath,
				}

				events = append(events, event)
			}
		}
	}

	return events, nil
}
