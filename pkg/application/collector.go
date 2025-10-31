package application

type CollectLocation struct {
	City string
	Lat  float64
	Lon  float64
	Radius float64
}

type Collector interface {
	Collect(location CollectLocation) ([]Event, error)
}
