package application_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/leorolland/sortir.in/pkg/application"
	applicationmocks "github.com/leorolland/sortir.in/pkg/application/mocks"
)

func TestGetPinsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventRepo := applicationmocks.NewMockEventRepository(ctrl)

	mockEventRepo.EXPECT().
		ByBoundsAndMaxDate(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("error"))

	pinsService := application.NewPins(mockEventRepo)

	_, err := pinsService.GetPins(application.Bounds{
		North: 48.9,
		South: 48.8,
		East:  2.4,
		West:  2.3,
	}, time.Date(2025, 11, 24, 0, 0, 0, 0, time.UTC))

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGetPinsSuccess(t *testing.T) {
	testCases := map[string]struct {
		bounds         application.Bounds
		maxDate        time.Time
		eventsReturned []application.Event
		expected       []application.Pin
	}{
		"with 1 event": {
			bounds: application.Bounds{
				North: 48.9,
				South: 48.8,
				East:  2.4,
				West:  2.3,
			},
			maxDate: time.Date(2025, 11, 24, 0, 0, 0, 0, time.UTC),
			eventsReturned: []application.Event{
				{
					Name: "Event 1",
					Loc:  application.EventLocation{Lat: 48.8, Lon: 2.3},
					Kind: application.KindConcert,
				},
			},
			expected: []application.Pin{
				{
					Loc:    application.EventLocation{Lat: 48.8, Lon: 2.3},
					Kind:   application.KindConcert,
					Amount: 1,
				},
			},
		},
		"with 2 events, not in same place": {
			bounds: application.Bounds{
				North: 48.9,
				South: 48.8,
				East:  2.4,
				West:  2.3,
			},
			maxDate: time.Date(2025, 11, 24, 0, 0, 0, 0, time.UTC),
			eventsReturned: []application.Event{
				{
					Name: "Event 1",
					Loc:  application.EventLocation{Lat: 48.8, Lon: 2.3},
					Kind: application.KindConcert,
				},
				{
					Name: "Event 2",
					Loc:  application.EventLocation{Lat: 48.9, Lon: 2.4},
					Kind: application.KindConcert,
				},
			},
			expected: []application.Pin{
				{
					Loc:    application.EventLocation{Lat: 48.8, Lon: 2.3},
					Kind:   application.KindConcert,
					Amount: 1,
				},
				{
					Loc:    application.EventLocation{Lat: 48.9, Lon: 2.4},
					Kind:   application.KindConcert,
					Amount: 1,
				},
			},
		},
		"with 3 events, 2 in same place": {
			bounds: application.Bounds{
				North: 48.9,
				South: 48.8,
				East:  2.4,
				West:  2.3,
			},
			maxDate: time.Date(2025, 11, 24, 0, 0, 0, 0, time.UTC),
			eventsReturned: []application.Event{
				{
					Name: "Event 1",
					Loc:  application.EventLocation{Lat: 48.8, Lon: 2.3},
					Kind: application.KindConcert,
				},
				{
					Name: "Event 2",
					Loc:  application.EventLocation{Lat: 48.8, Lon: 2.3},
					Kind: application.KindConcert,
				},
				{
					Name: "Event 3",
					Loc:  application.EventLocation{Lat: 48.9, Lon: 2.4},
					Kind: application.KindConcert,
				},
			},
			expected: []application.Pin{
				{
					Loc:    application.EventLocation{Lat: 48.8, Lon: 2.3},
					Kind:   application.KindConcert,
					Amount: 2,
				},
				{
					Loc:    application.EventLocation{Lat: 48.9, Lon: 2.4},
					Kind:   application.KindConcert,
					Amount: 1,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockEventRepo := applicationmocks.NewMockEventRepository(ctrl)

			mockEventRepo.EXPECT().
				ByBoundsAndMaxDate(tc.bounds, tc.maxDate).
				Return(tc.eventsReturned, nil)

			pinsService := application.NewPins(mockEventRepo)

			pins, err := pinsService.GetPins(tc.bounds, tc.maxDate)

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if !reflect.DeepEqual(pins, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, pins)
			}
		})
	}
}
