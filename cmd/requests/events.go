package requests

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/pocketbase/pocketbase/core"
)

func PutEvents(e *core.RequestEvent) error {
	var events []application.Event
	if err := json.NewDecoder(e.Request.Body).Decode(&events); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	eventsCollection, err := e.App.FindCollectionByNameOrId("events")
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find events collection: " + err.Error()})
	}

	for _, event := range events {
		eventRecord, err := e.App.FindFirstRecordByData(eventsCollection, "name", event.Name)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = createEvent(e, eventsCollection, event)
				if err != nil {
					return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create event: " + err.Error()})
				}

				continue
			} else {

				return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed check if event already exists: " + err.Error()})
			}
		}

		err = updateEvent(e, eventRecord, event)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update event: " + err.Error()})
		}
	}

	return e.JSON(http.StatusOK, map[string]string{"message": "Events batch updated"})
}

func createEvent(e *core.RequestEvent, eventsCollection *core.Collection, event application.Event) error {
	eventRecord := core.NewRecord(eventsCollection)

	eventRecord.Set("name", event.Name)
	eventRecord.Set("kind", event.Kind)
	eventRecord.Set("genres", event.Genres)
	eventRecord.Set("begin", event.Begin.Format(time.RFC3339))
	eventRecord.Set("end", event.End.Format(time.RFC3339))
	eventRecord.Set("place", event.Place)
	eventRecord.Set("address", event.Address)
	eventRecord.Set("price", event.Price)
	eventRecord.Set("price_currency", event.PriceCurrency)

	return e.App.Save(eventRecord)
}

func updateEvent(e *core.RequestEvent, eventRecord *core.Record, event application.Event) error {
	eventRecord.Set("kind", event.Kind)
	eventRecord.Set("genres", event.Genres)
	eventRecord.Set("begin", event.Begin.Format(time.RFC3339))
	eventRecord.Set("end", event.End.Format(time.RFC3339))
	eventRecord.Set("place", event.Place)
	eventRecord.Set("address", event.Address)
	eventRecord.Set("price", event.Price)
	eventRecord.Set("price_currency", event.PriceCurrency)

	return e.App.Save(eventRecord)
}
