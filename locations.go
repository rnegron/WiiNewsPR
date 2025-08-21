package main

import (
	"WiiNewsPR/news"
	"unicode/utf16"
)

type Location struct {
	TextOffset   uint32
	Latitude     int16
	Longitude    int16
	CountryCode  uint8
	RegionCode   uint8
	LocationCode uint16
	Zoom         uint8
	_            [3]byte
}

// FORK UPDATE: Removed floatCompare function - no longer needed with hardcoded location

func CoordinateEncode(value float64) int16 {
	value /= 0.0054931640625
	return int16(value)
}

// FORK UPDATE: Use San Juan constants from news package

func (n *News) MakeLocationTable() {
	n.Header.LocationTableOffset = n.GetCurrentSize()

	// FORK UPDATE: Use San Juan constants from news package
	n.Locations = append(n.Locations, Location{
		TextOffset:   0,
		Latitude:     CoordinateEncode(news.SanJuanLatitude),
		Longitude:    CoordinateEncode(news.SanJuanLongitude),
		CountryCode:  0,
		RegionCode:   0,
		LocationCode: 0,
		Zoom:         6,
	})

	// Set text offset and add location name
	n.Locations[0].TextOffset = n.GetCurrentSize()
	encoded := utf16.Encode([]rune(news.SanJuanName))
	n.LocationText = append(n.LocationText, encoded...)
	n.LocationText = append(n.LocationText, 0)
	for n.GetCurrentSize()%4 != 0 {
		n.LocationText = append(n.LocationText, 0)
	}

	n.Header.NumberOfLocations = 1
}
