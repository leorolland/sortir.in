package application

type CollectLocation struct {
	City   string
	Lat    float64
	Lon    float64
	Radius float64
}

type Collector interface {
	Collect(location CollectLocation) ([]Event, error)
}

// GlobalCollector is a collector for nationwide datasets that have no
// per-location query: it is collected once per run, not once per city.
type GlobalCollector interface {
	CollectAll() ([]Event, error)
}
