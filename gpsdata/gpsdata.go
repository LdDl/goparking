package gpsdata

import (
	"time"
)

// GPSData - Struct for storing gps data from bu-353 USB
type GPSData struct {
	GPSTime      time.Time
	GPSLongitude float64
	GPSLatitude  float64
}
