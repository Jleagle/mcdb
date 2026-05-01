package scanner

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Jleagle/mcdb/storage"
)

var locationCache sync.Map

type ipAPIResponse struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	ISP         string  `json:"isp"`
}

func GetLocation(ip string) (*storage.Location, error) {
	// Remove port if present
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}

	// Check cache
	if val, ok := locationCache.Load(ip); ok {
		return val.(*storage.Location), nil
	}

	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ip-api returned status %d", resp.StatusCode)
	}

	var result ipAPIResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("ip-api failed: %s", result.Status)
	}

	loc := &storage.Location{
		Country:     result.Country,
		CountryCode: result.CountryCode,
		Region:      result.Region,
		RegionName:  result.RegionName,
		City:        result.City,
		Lat:         result.Lat,
		Lon:         result.Lon,
		ISP:         result.ISP,
		Geo: &storage.GeoJSONPoint{
			Type:        "Point",
			Coordinates: []float64{result.Lon, result.Lat},
		},
	}

	// Save to cache
	locationCache.Store(ip, loc)

	return loc, nil
}
