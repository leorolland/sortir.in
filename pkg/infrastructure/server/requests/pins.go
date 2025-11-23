package requests

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/pocketbase/pocketbase/core"
)

func GetPins(e *core.RequestEvent) error {
	bounds, err := getBoundsFromQueryParams(e.Request.URL.Query())
	if err != nil {
		return e.Error(http.StatusBadRequest, fmt.Sprintf("invalid bounds: %v", err), nil)
	}

	maxTime, err := getMaxTimeFromQueryParams(e.Request.URL.Query())
	if err != nil {
		return e.Error(http.StatusBadRequest, fmt.Sprintf("invalid max time: %v", err), nil)
	}

	pinsService, ok := e.App.Store().Get("pinsService").(application.PinsService)
	if !ok {
		return e.Error(http.StatusInternalServerError, "pins service not found", nil)
	}

	pins, err := pinsService.GetPins(bounds, maxTime)
	if err != nil {
		return e.Error(http.StatusInternalServerError, fmt.Sprintf("failed to get pins: %v", err), nil)
	}

	return e.JSON(http.StatusOK, pins)
}

func getBoundsFromQueryParams(queryParams url.Values) (application.Bounds, error) {
	north, err := strconv.ParseFloat(queryParams.Get("north"), 64)
	if err != nil {
		return application.Bounds{}, fmt.Errorf("invalid north: %w", err)
	}

	south, err := strconv.ParseFloat(queryParams.Get("south"), 64)
	if err != nil {
		return application.Bounds{}, fmt.Errorf("invalid south: %w", err)
	}

	east, err := strconv.ParseFloat(queryParams.Get("east"), 64)
	if err != nil {
		return application.Bounds{}, fmt.Errorf("invalid east: %w", err)
	}

	west, err := strconv.ParseFloat(queryParams.Get("west"), 64)
	if err != nil {
		return application.Bounds{}, fmt.Errorf("invalid west: %w", err)
	}

	return application.Bounds{
		North: north,
		South: south,
		East:  east,
		West:  west,
	}, nil
}

func getMaxTimeFromQueryParams(queryParams url.Values) (time.Time, error) {
	maxTimeStr := queryParams.Get("max_time")
	maxTime, err := time.Parse(time.RFC3339, maxTimeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid max time: %w", err)
	}

	return maxTime, nil
}
