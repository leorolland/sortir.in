package application

type CollectLocation struct {
	City string
	Long float64
	Lat  float64
	Radius float64
}

type Collector interface {
	Collect(location CollectLocation) ([]Event, error)
}
