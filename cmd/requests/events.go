package requests

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func PutEvents(e *core.RequestEvent) error {
	var events []application.Event
	if err := json.NewDecoder(e.Request.Body).Decode(&events); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	for _, event := range events {
		genresJSON, err := json.Marshal(event.Genres)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal genres: " + err.Error()})
		}

		locJSON, err := json.Marshal(event.Loc)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal loc: " + err.Error()})
		}

		if !event.IsValid() {
			continue
		}

		priceFloat := 0.0
		if event.Price != nil {
			priceFloat = *event.Price
		}

		currencyString := ""
		if event.PriceCurrency != nil {
			currencyString = *event.PriceCurrency
		}

		_, err = e.App.DB().NewQuery(`
			INSERT INTO events (name, kind, genres, begin, end, loc, place, address, price, price_currency, source, img)
			VALUES ({:name}, {:kind}, {:genres}, {:begin}, {:end}, {:loc}, {:place}, {:address}, {:price}, {:price_currency}, {:source}, {:img})
			ON CONFLICT (name, begin, end) DO UPDATE SET
				kind = {:kind},
				genres = {:genres},
				loc = {:loc},
				place = {:place},
				address = {:address},
				price = {:price},
				price_currency = {:price_currency},
				source = {:source},
				img = {:img}
		`).Bind(dbx.Params{
			"name":           event.Name,
			"kind":           event.Kind,
			"genres":         genresJSON,
			"begin":          event.Begin.Format(time.RFC3339),
			"end":            event.End.Format(time.RFC3339),
			"loc":            locJSON,
			"place":          event.Place,
			"address":        event.Address,
			"price":          priceFloat,
			"price_currency": currencyString,
			"source":         event.Source,
			"img":            event.Img,
		}).Execute()
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update event: " + err.Error()})
		}
	}

	return e.JSON(http.StatusOK, map[string]string{"message": "Events batch updated"})
}
