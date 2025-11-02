package application

type LocationsIterator interface {
	Next() *CollectLocation
}

type FrenchCitiesIterator struct {
	cities []CollectLocation
	index  int
}

func NewFrenchCitiesIterator() LocationsIterator {
	cities := []CollectLocation{
		{City: "Paris", Lat: 48.8566, Lon: 2.3522, Radius: 10.0},
		{City: "Marseille", Lat: 43.2965, Lon: 5.3698, Radius: 8.0},
		{City: "Lyon", Lat: 45.7640, Lon: 4.8357, Radius: 8.0},
		{City: "Toulouse", Lat: 43.6047, Lon: 1.4442, Radius: 7.0},
		{City: "Nice", Lat: 43.7102, Lon: 7.2620, Radius: 6.0},
		{City: "Nantes", Lat: 47.2184, Lon: -1.5536, Radius: 6.0},
		{City: "Montpellier", Lat: 43.6108, Lon: 3.8767, Radius: 6.0},
		{City: "Strasbourg", Lat: 48.5734, Lon: 7.7521, Radius: 6.0},
		{City: "Bordeaux", Lat: 44.8378, Lon: -0.5792, Radius: 6.0},
		{City: "Lille", Lat: 50.6292, Lon: 3.0573, Radius: 6.0},
		{City: "Rennes", Lat: 48.1173, Lon: -1.6778, Radius: 5.0},
		{City: "Reims", Lat: 49.2583, Lon: 4.0317, Radius: 5.0},
		{City: "Le Havre", Lat: 49.4944, Lon: 0.1079, Radius: 5.0},
		{City: "Saint-Étienne", Lat: 45.4397, Lon: 4.3872, Radius: 5.0},
		{City: "Toulon", Lat: 43.1242, Lon: 5.9280, Radius: 5.0},
		{City: "Angers", Lat: 47.4784, Lon: -0.5632, Radius: 5.0},
		{City: "Grenoble", Lat: 45.1885, Lon: 5.7245, Radius: 5.0},
		{City: "Dijon", Lat: 47.3220, Lon: 5.0415, Radius: 5.0},
		{City: "Nîmes", Lat: 43.8367, Lon: 4.3601, Radius: 5.0},
		{City: "Aix-en-Provence", Lat: 43.5297, Lon: 5.4474, Radius: 5.0},
		{City: "Saint-Denis", Lat: 48.9358, Lon: 2.3596, Radius: 5.0},
		{City: "Le Mans", Lat: 48.0061, Lon: 0.1996, Radius: 5.0},
		{City: "Clermont-Ferrand", Lat: 45.7772, Lon: 3.0870, Radius: 5.0},
		{City: "Tours", Lat: 47.3941, Lon: 0.6848, Radius: 5.0},
		{City: "Limoges", Lat: 45.8336, Lon: 1.2611, Radius: 5.0},
		{City: "Villeurbanne", Lat: 45.7712, Lon: 4.8800, Radius: 4.0},
		{City: "Amiens", Lat: 49.8942, Lon: 2.2957, Radius: 4.0},
		{City: "Metz", Lat: 49.1193, Lon: 6.1757, Radius: 4.0},
		{City: "Besançon", Lat: 47.2380, Lon: 6.0243, Radius: 4.0},
		{City: "Perpignan", Lat: 42.6986, Lon: 2.8956, Radius: 4.0},
	}

	return &FrenchCitiesIterator{
		cities: cities,
		index:  0,
	}
}

func (f *FrenchCitiesIterator) Next() *CollectLocation {
	if f.index >= len(f.cities) {
		return nil
	}
	location := f.cities[f.index]
	f.index++
	return &location
}
