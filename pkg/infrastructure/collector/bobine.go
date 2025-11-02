package collector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
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

type bobineResponse struct {
	Movie    bobineMovie     `json:"movie"`
	Theaters []bobineTheater `json:"theaters"`
}

func (c *bobineCollector) Collect(location application.CollectLocation) ([]application.Event, error) {
	page := 1
	pageSize := 20
	allEvents := []application.Event{}

	for {
		// Format dates in the required format (YYYY-MM-DDThh:mm:ssZ)
		// Start date is 2 days ago
		startDate := time.Now().AddDate(0, 0, -2).UTC().Format("2006-01-02T15:04:05Z")
		// End date must be within 7 days of start date per API requirement
		endDate := time.Now().AddDate(0, 0, 5).UTC().Format("2006-01-02T15:04:05Z")

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

		// Extract genres as a slice
		var genres []string
		if movie.Genres != nil {
			// Split by comma if genres is not nil
			genres = []string{*movie.Genres}
		} else {
			genres = []string{"movie"}
		}

		// Process each theater
		for _, theater := range movieData.Theaters {
			// Process each showtime
			for _, showtime := range theater.Showtimes {
				// Calculate end time based on movie duration
				endTime := showtime.Showtime.Add(time.Duration(movie.Duration) * time.Minute)

				// Determine movie title (prefer VF if available, otherwise use VO)
				title := movie.TitleVF
				if title == "" {
					title = movie.TitleVO
				}

				// Create price pointer
				price := theater.FullPrice

				// Create event
				event := application.Event{
					Name:   title,
					Kind:   "movie",
					Genres: genres,
					Begin:  showtime.Showtime,
					End:    endTime,
					Loc: application.EventLocation{
						Lat: theater.Latitude,
						Lon: theater.Longitude,
					},
					Place:   theater.Name,
					Address: theater.Address,
					Price:   &price,
					Source:  fmt.Sprintf("https://bobine.art/movie/%d", movie.ID),
					Img:     movie.PosterPath,
				}

				// Add currency if price is set
				if price > 0 {
					currency := "EUR"
					event.PriceCurrency = &currency
				}

				events = append(events, event)
			}
		}
	}

	return events, nil
}
