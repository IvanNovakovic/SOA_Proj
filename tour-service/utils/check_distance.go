package utils

import "math"

// calculate distance between two geographic coordinates using Haversine formula
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000.0

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// threshold in meters to consider a tourist "near" a key point
const KeyPointThreshold = 30.0

func IsNearby(lat1, lon1, lat2, lon2 float64) bool {
	distance := HaversineDistance(lat1, lon1, lat2, lon2)
	return distance <= KeyPointThreshold
}
