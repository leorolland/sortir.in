package application

type LocationsIterator interface {
	Next() *CollectLocation
}

// EuropeCitiesIterator is a temporary hard-coded catalog of main European
// cities (docs/adr/0004: the France -> Europe -> world ramp). The long-term
// plan replaces it with the GeoNames-backed locations table.
type EuropeCitiesIterator struct {
	cities []CollectLocation
	index  int
}

func NewEuropeCitiesIterator() LocationsIterator {
	cities := []CollectLocation{
		// France
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
		// United Kingdom & Ireland
		{City: "London", Lat: 51.5074, Lon: -0.1278, Radius: 10.0},
		{City: "Manchester", Lat: 53.4808, Lon: -2.2426, Radius: 7.0},
		{City: "Birmingham", Lat: 52.4862, Lon: -1.8904, Radius: 6.0},
		{City: "Glasgow", Lat: 55.8642, Lon: -4.2518, Radius: 6.0},
		{City: "Edinburgh", Lat: 55.9533, Lon: -3.1883, Radius: 5.0},
		{City: "Dublin", Lat: 53.3498, Lon: -6.2603, Radius: 7.0},
		// Germany
		{City: "Berlin", Lat: 52.5200, Lon: 13.4050, Radius: 10.0},
		{City: "Hamburg", Lat: 53.5511, Lon: 9.9937, Radius: 8.0},
		{City: "Munich", Lat: 48.1351, Lon: 11.5820, Radius: 8.0},
		{City: "Cologne", Lat: 50.9375, Lon: 6.9603, Radius: 6.0},
		{City: "Frankfurt", Lat: 50.1109, Lon: 8.6821, Radius: 6.0},
		{City: "Stuttgart", Lat: 48.7758, Lon: 9.1829, Radius: 6.0},
		{City: "Düsseldorf", Lat: 51.2277, Lon: 6.7735, Radius: 5.0},
		{City: "Leipzig", Lat: 51.3397, Lon: 12.3731, Radius: 5.0},
		// Spain
		{City: "Madrid", Lat: 40.4168, Lon: -3.7038, Radius: 10.0},
		{City: "Barcelona", Lat: 41.3874, Lon: 2.1686, Radius: 9.0},
		{City: "Valencia", Lat: 39.4699, Lon: -0.3763, Radius: 6.0},
		{City: "Seville", Lat: 37.3891, Lon: -5.9845, Radius: 6.0},
		{City: "Zaragoza", Lat: 41.6488, Lon: -0.8891, Radius: 5.0},
		{City: "Málaga", Lat: 36.7213, Lon: -4.4213, Radius: 5.0},
		{City: "Bilbao", Lat: 43.2630, Lon: -2.9350, Radius: 4.0},
		// Italy
		{City: "Rome", Lat: 41.9028, Lon: 12.4964, Radius: 10.0},
		{City: "Milan", Lat: 45.4642, Lon: 9.1900, Radius: 9.0},
		{City: "Naples", Lat: 40.8518, Lon: 14.2681, Radius: 7.0},
		{City: "Turin", Lat: 45.0703, Lon: 7.6869, Radius: 7.0},
		{City: "Bologna", Lat: 44.4949, Lon: 11.3426, Radius: 5.0},
		{City: "Florence", Lat: 43.7696, Lon: 11.2558, Radius: 4.0},
		{City: "Venice", Lat: 45.4408, Lon: 12.3155, Radius: 4.0},
		{City: "Palermo", Lat: 38.1157, Lon: 13.3615, Radius: 6.0},
		// Portugal
		{City: "Lisbon", Lat: 38.7223, Lon: -9.1393, Radius: 8.0},
		{City: "Porto", Lat: 41.1579, Lon: -8.6291, Radius: 6.0},
		// Benelux
		{City: "Amsterdam", Lat: 52.3676, Lon: 4.9041, Radius: 7.0},
		{City: "Rotterdam", Lat: 51.9244, Lon: 4.4777, Radius: 5.0},
		{City: "The Hague", Lat: 52.0705, Lon: 4.3007, Radius: 4.0},
		{City: "Brussels", Lat: 50.8503, Lon: 4.3517, Radius: 8.0},
		{City: "Antwerp", Lat: 51.2194, Lon: 4.4025, Radius: 5.0},
		{City: "Luxembourg", Lat: 49.6116, Lon: 6.1319, Radius: 4.0},
		// Switzerland
		{City: "Zurich", Lat: 47.3769, Lon: 8.5417, Radius: 7.0},
		{City: "Geneva", Lat: 46.2044, Lon: 6.1432, Radius: 7.0},
		{City: "Basel", Lat: 47.5596, Lon: 7.5886, Radius: 4.0},
		{City: "Lausanne", Lat: 46.5197, Lon: 6.6323, Radius: 4.0},
		{City: "Bern", Lat: 46.9480, Lon: 7.4474, Radius: 4.0},
		// Austria
		{City: "Vienna", Lat: 48.2082, Lon: 16.3738, Radius: 9.0},
		{City: "Graz", Lat: 47.0707, Lon: 15.4395, Radius: 4.0},
		{City: "Salzburg", Lat: 47.8095, Lon: 13.0550, Radius: 4.0},
		// Nordics
		{City: "Stockholm", Lat: 59.3293, Lon: 18.0686, Radius: 8.0},
		{City: "Gothenburg", Lat: 57.7089, Lon: 11.9746, Radius: 5.0},
		{City: "Copenhagen", Lat: 55.6761, Lon: 12.5683, Radius: 7.0},
		{City: "Oslo", Lat: 59.9139, Lon: 10.7522, Radius: 7.0},
		{City: "Helsinki", Lat: 60.1699, Lon: 24.9384, Radius: 6.0},
		{City: "Reykjavik", Lat: 64.1466, Lon: -21.9426, Radius: 4.0},
		// Poland
		{City: "Warsaw", Lat: 52.2297, Lon: 21.0122, Radius: 8.0},
		{City: "Kraków", Lat: 50.0647, Lon: 19.9450, Radius: 6.0},
		{City: "Wrocław", Lat: 51.1079, Lon: 17.0385, Radius: 5.0},
		{City: "Gdańsk", Lat: 54.3520, Lon: 18.6466, Radius: 5.0},
		{City: "Poznań", Lat: 52.4064, Lon: 16.9252, Radius: 4.0},
		// Central Europe
		{City: "Prague", Lat: 50.0755, Lon: 14.4378, Radius: 7.0},
		{City: "Brno", Lat: 49.1951, Lon: 16.6068, Radius: 4.0},
		{City: "Budapest", Lat: 47.4979, Lon: 19.0402, Radius: 8.0},
		{City: "Bratislava", Lat: 48.1486, Lon: 17.1077, Radius: 4.0},
		{City: "Ljubljana", Lat: 46.0569, Lon: 14.5058, Radius: 4.0},
		// Baltics
		{City: "Tallinn", Lat: 59.4370, Lon: 24.7536, Radius: 4.0},
		{City: "Riga", Lat: 56.9496, Lon: 24.1052, Radius: 5.0},
		{City: "Vilnius", Lat: 54.6872, Lon: 25.2797, Radius: 4.0},
		// Balkans & South-East Europe
		{City: "Zagreb", Lat: 45.8150, Lon: 15.9819, Radius: 5.0},
		{City: "Belgrade", Lat: 44.7866, Lon: 20.4489, Radius: 6.0},
		{City: "Sarajevo", Lat: 43.8563, Lon: 18.4131, Radius: 4.0},
		{City: "Skopje", Lat: 41.9981, Lon: 21.4254, Radius: 4.0},
		{City: "Tirana", Lat: 41.3275, Lon: 19.8187, Radius: 4.0},
		{City: "Sofia", Lat: 42.6977, Lon: 23.3219, Radius: 5.0},
		{City: "Bucharest", Lat: 44.4268, Lon: 26.1025, Radius: 7.0},
		// Greece, Cyprus, Malta
		{City: "Athens", Lat: 37.9838, Lon: 23.7275, Radius: 8.0},
		{City: "Thessaloniki", Lat: 40.6401, Lon: 22.9444, Radius: 5.0},
		{City: "Nicosia", Lat: 35.1856, Lon: 33.3823, Radius: 4.0},
		{City: "Valletta", Lat: 35.8989, Lon: 14.5146, Radius: 3.0},
		// Eastern Europe & Turkey
		{City: "Kyiv", Lat: 50.4501, Lon: 30.5234, Radius: 8.0},
		{City: "Chișinău", Lat: 47.0105, Lon: 28.8638, Radius: 4.0},
		{City: "Istanbul", Lat: 41.0082, Lon: 28.9784, Radius: 10.0},
	}

	return &EuropeCitiesIterator{
		cities: cities,
		index:  0,
	}
}

func (e *EuropeCitiesIterator) Next() *CollectLocation {
	if e.index >= len(e.cities) {
		return nil
	}
	location := e.cities[e.index]
	e.index++
	return &location
}
