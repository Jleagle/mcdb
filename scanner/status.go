package scanner

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"sort"
	"strings"
	"time"

	"github.com/mcstatus-io/mcutil/v2/response"
)

type Server struct {
	IP        string                  `json:"ip" bson:"ip"`
	IsOnline  bool                    `json:"is_online" bson:"is_online"`
	Location  *Location               `json:"location,omitempty" bson:"location,omitempty"`
	Players   Players                 `json:"players" bson:"players"`
	Version   Version                 `json:"version" bson:"version"`
	Favicon   Icon                    `json:"favicon" bson:"favicon"`
	Delay     time.Duration           `json:"delay" bson:"delay"`
	Tags      []string                `json:"tags" bson:"tags"`
	IsJava    bool                    `json:"is_java" bson:"is_java"`
	IsBedrock bool                    `json:"is_bedrock" bson:"is_bedrock"`
	HasQuery  bool                    `json:"has_query" bson:"has_query"`
	Java      *response.JavaStatus    `json:"java,omitempty" bson:"java,omitempty"`
	Bedrock   *response.BedrockStatus `json:"bedrock,omitempty" bson:"bedrock,omitempty"`
	Query     *response.FullQuery     `json:"query,omitempty" bson:"query,omitempty"`
	UpdatedAt time.Time               `json:"updated_at" bson:"updated_at"`
}

const pngPrefix = "data:image/png;base64,"

type Icon string

func (i Icon) ToImage() (icon image.Image, err error) {

	if !strings.HasPrefix(string(i), pngPrefix) {
		return nil, fmt.Errorf("server icon should prepended with %q", pngPrefix)
	}

	base64png := strings.TrimPrefix(string(i), pngPrefix)
	r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64png))

	return png.Decode(r)
}

type Location struct {
	Country     string        `json:"country" bson:"country"`
	CountryCode string        `json:"country_code" bson:"country_code"`
	Region      string        `json:"region" bson:"region"`
	RegionName  string        `json:"region_name" bson:"region_name"`
	City        string        `json:"city" bson:"city"`
	Lat         float64       `json:"lat" bson:"lat"`
	Lon         float64       `json:"lon" bson:"lon"`
	ISP         string        `json:"isp" bson:"isp"`
	Geo         *GeoJSONPoint `json:"geo,omitempty" bson:"geo,omitempty"`
}

type GeoJSONPoint struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"` // [longitude, latitude]
}

type Players struct {
	Max    int `json:"max"`
	Online int `json:"online"`
}

type Version struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

func (s Server) MOTD() string {
	if s.Java != nil {
		return s.Java.MOTD.Clean
	}
	if s.Bedrock != nil && s.Bedrock.MOTD != nil {
		return s.Bedrock.MOTD.Clean
	}
	if s.Query != nil {
		return s.Query.Data["hostname"]
	}
	return ""
}

func (s Server) GetTags() []string {
	tags := make(map[string]struct{})

	// 1. Check Query Data
	if s.Query != nil {
		if gt, ok := s.Query.Data["gametype"]; ok && gt != "" {
			tags[strings.ToLower(gt)] = struct{}{}
		}
		// Extract server software from plugins string (e.g. "Paper on Bukkit 1.20.1")
		if plugins, ok := s.Query.Data["plugins"]; ok && plugins != "" {
			parts := strings.Split(plugins, " ")
			if len(parts) > 0 {
				software := strings.ToLower(strings.TrimSuffix(parts[0], ":"))
				if software != "vanilla" {
					tags[software] = struct{}{}
				}
			}
		}
	}

	// 2. Check Java ModInfo
	if s.IsJava && s.Java != nil && s.Java.ModInfo != nil {
		tags[strings.ToLower(s.Java.ModInfo.Type)] = struct{}{}
	}

	// 3. Scan MOTD for common keywords
	motd := strings.ToLower(s.MOTD())
	keywords := []string{
		"skyblock", "prison", "survival", "factions", "creative", "smp", "vanilla",
		"bedwars", "lifesteal", "cobblemon", "pixelmon", "anarchy", "towny", "hardcore",
		"minigames", "parkour", "kitpvp", "pve", "pvp", "rpg",
	}
	for _, kw := range keywords {
		if strings.Contains(motd, kw) {
			tags[kw] = struct{}{}
		}
	}

	// Convert map to slice
	var result []string
	for t := range tags {
		result = append(result, t)
	}
	sort.Strings(result)
	return result
}

func (s Server) PlayerPercent() int {
	if s.Players.Max <= 0 {
		return 0
	}
	return int(float64(s.Players.Online) / float64(s.Players.Max) * 100)
}
