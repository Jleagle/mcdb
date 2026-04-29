package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/mcstatus-io/mcutil/v2"
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

func GetLocation(ip string) (*Location, error) {
	// Check cache
	if val, ok := locationCache.Load(ip); ok {
		return val.(*Location), nil
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

	loc := &Location{
		Country:     result.Country,
		CountryCode: result.CountryCode,
		Region:      result.Region,
		RegionName:  result.RegionName,
		City:        result.City,
		Lat:         result.Lat,
		Lon:         result.Lon,
		ISP:         result.ISP,
		Geo: &GeoJSONPoint{
			Type:        "Point",
			Coordinates: []float64{result.Lon, result.Lat},
		},
	}

	// Save to cache
	locationCache.Store(ip, loc)

	return loc, nil
}

func Probe(ctx context.Context, host string, loc *Location) (*Server, error) {
	s := &Server{
		IP:        host,
		UpdatedAt: time.Now(),
		IsOnline:  false,
	}

	// Try to get location if not provided
	if loc != nil {
		s.Location = loc
	} else {
		newLoc, err := GetLocation(host)
		if err == nil {
			s.Location = newLoc
		} else {
			log.Printf("Failed to get location for %s: %v", host, err)
		}
	}

	found := false

	// 1. Try Java Status
	javaStatus, err := mcutil.Status(ctx, host, 25565)
	if err == nil {
		s.IsJava = true
		s.Java = javaStatus
		found = true

		// Map to legacy fields for web UI compatibility
		s.Version.Name = javaStatus.Version.NameClean
		s.Version.Protocol = int(javaStatus.Version.Protocol)
		if javaStatus.Players.Online != nil {
			s.Players.Online = int(*javaStatus.Players.Online)
		}
		if javaStatus.Players.Max != nil {
			s.Players.Max = int(*javaStatus.Players.Max)
		}
		if javaStatus.Favicon != nil {
			s.Favicon = Icon(*javaStatus.Favicon)
		}
		s.Delay = javaStatus.Latency
	}

	// 2. Try Bedrock Status
	bedrockStatus, err := mcutil.StatusBedrock(ctx, host, 19132)
	if err == nil {
		s.IsBedrock = true
		s.Bedrock = bedrockStatus
		found = true
	}

	// 3. Try Query (GS4)
	queryStatus, err := mcutil.FullQuery(ctx, host, 25565)
	if err == nil {
		s.HasQuery = true
		s.Query = queryStatus
		found = true
	} else {
		// Try Bedrock port for query
		queryStatus, err = mcutil.FullQuery(ctx, host, 19132)
		if err == nil {
			s.HasQuery = true
			s.Query = queryStatus
			found = true
		}
	}

	s.IsOnline = found
	s.Tags = s.GetTags()
	return s, nil
}
