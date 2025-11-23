package applicationtest

import (
	"testing"

	"github.com/leorolland/sortir.in/pkg/application"
	"github.com/stretchr/testify/require"
)

func MustValidateEvent(t *testing.T, event application.Event) {
	t.Helper()

	require.True(t, event.IsValid(), "Event should be valid")
}

func MustValidateEvents(t *testing.T, events []application.Event) []application.Event {
	t.Helper()

	for _, event := range events {
		MustValidateEvent(t, event)
	}

	return events
}
