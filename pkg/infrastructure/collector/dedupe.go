package collector

import (
	"fmt"

	"github.com/leorolland/sortir.in/pkg/application"
)

// dedupeEvents removes events sharing the same name and time span, a common
// artifact of sources expanding multi-occurrence events per date.
func dedupeEvents(events []application.Event) []application.Event {
	seen := make(map[string]bool)
	unique := make([]application.Event, 0, len(events))
	for _, event := range events {
		key := fmt.Sprintf("%s|%d|%d", event.Name, event.Begin.Unix(), event.End.Unix())
		if seen[key] {
			continue
		}
		seen[key] = true
		unique = append(unique, event)
	}
	return unique
}
