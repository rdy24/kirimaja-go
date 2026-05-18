package opencage

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	neturl "net/url"
)

type Client struct{ apiKey string }

type Location struct {
	Lat float64
	Lng float64
}

type opencageResponse struct {
	Results []struct {
		Geometry struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"geometry"`
	} `json:"results"`
}

func New(apiKey string) *Client {
	return &Client{apiKey}
}

func (c *Client) Geocode(address string) (*Location, error) {
	url := fmt.Sprintf(
		"https://api.opencagedata.com/geocode/v1/json?q=%s&key=%s&limit=1",
		neturl.QueryEscape(address), c.apiKey,
	)
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("opencage request failed: %w", err)
	}
	defer resp.Body.Close()

	var result opencageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("opencage decode failed: %w", err)
	}
	if len(result.Results) == 0 {
		return nil, errors.New("alamat tidak ditemukan")
	}
	g := result.Results[0].Geometry
	return &Location{Lat: g.Lat, Lng: g.Lng}, nil
}

// HaversineKm returns straight-line distance in km — equivalent to geolib.getDistance(), no API call.
func HaversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return earthRadius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
