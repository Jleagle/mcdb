package web

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Jleagle/mcdb/scanner"
	"github.com/Jleagle/mcdb/storage"
)

type ServerWithDistance struct {
	scanner.Server
	DistanceKM float64
}

// IndexTemplateData holds data for the server listing page.
type IndexTemplateData struct {
	BasePageData
	Servers              []ServerWithDistance
	Page                 int
	HasPrev              bool
	HasNext              bool
	PrevPage             int
	NextPage             int
	TotalPages           int
	TotalResults         int64
	CurrentSort          string
	UserLat              float64
	UserLon              float64
	UserLocationName     string
	CurrentIPSearch      string
	CurrentNameSearch    string
	CurrentTagsSearch    string
	CurrentVersionSearch string
	CurrentCountrySearch string
	CurrentPrivacySearch string
	UblockLink           string // New field for the uBlock Origin link
}

func indexHandler(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}

		sort := r.URL.Query().Get("sort")
		if sort == "" {
			sort = "players"
		}

		ipSearch := r.URL.Query().Get("ip")
		nameSearch := r.URL.Query().Get("name")
		tagsSearch := r.URL.Query().Get("tags")
		versionSearch := r.URL.Query().Get("version")
		countrySearch := r.URL.Query().Get("country")
		privacySearch := r.URL.Query().Get("privacy")

		limit := int64(20)
		offset := int64(page-1) * limit

		opts := storage.ListOptions{
			Limit:   limit,
			Offset:  offset,
			Sort:    sort,
			IP:      ipSearch,
			Name:    nameSearch,
			Tags:    tagsSearch,
			Version: versionSearch,
			Country: countrySearch,
			Privacy: privacySearch,
		}

		var userLocationName string
		if sort == "location" {
			// Get user IP
			ip := r.Header.Get("X-Forwarded-For")
			if ip != "" {
				ip = strings.Split(ip, ",")[0]
				ip = strings.TrimSpace(ip)
			}
			if ip == "" {
				ip = strings.Split(r.RemoteAddr, ":")[0]
			}
			// Get location of user
			loc, err := scanner.GetLocation(ip)
			if err == nil {
				opts.Lat = loc.Lat
				opts.Lon = loc.Lon
				userLocationName = loc.City + ", " + loc.Country
			}
		}

		servers, err := store.ListServers(opts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Calculate distances if sorting by location
		var serversWithDistance = make([]ServerWithDistance, len(servers))
		for i, s := range servers {
			sd := ServerWithDistance{Server: s}
			if sort == "location" && opts.Lat != 0 && opts.Lon != 0 {
				if s.Location != nil && s.Location.Lat != 0 && s.Location.Lon != 0 {
					sd.DistanceKM = haversineDistance(opts.Lat, opts.Lon, s.Location.Lat, s.Location.Lon)
				} else {
					sd.DistanceKM = -1.0
				}
			}
			serversWithDistance[i] = sd
		}

		total, err := store.CountServersWithOptions(opts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		totalPages := int((total + limit - 1) / limit)
		if totalPages < 1 {
			totalPages = 1
		}

		data := IndexTemplateData{
			BasePageData: BasePageData{
				Title:        "Minecraft Server List",
				Description:  "Browse all Minecraft servers, sorted by " + sort + ".",
				CanonicalURL: "https://" + r.Host + r.URL.RequestURI(),
				OGImage:      "https://" + r.Host + "/logo.png",
				TwitterImage: "https://" + r.Host + "/logo.png",
			},
			Servers:              serversWithDistance,
			Page:                 page,
			HasPrev:              page > 1,
			HasNext:              offset+limit < total,
			PrevPage:             page - 1,
			NextPage:             page + 1,
			TotalPages:           totalPages,
			TotalResults:         total,
			CurrentSort:          sort,
			UserLat:              opts.Lat,
			UserLon:              opts.Lon,
			UserLocationName:     userLocationName,
			CurrentIPSearch:      ipSearch,
			CurrentNameSearch:    nameSearch,
			CurrentTagsSearch:    tagsSearch,
			CurrentVersionSearch: versionSearch,
			CurrentCountrySearch: countrySearch,
			CurrentPrivacySearch: privacySearch,
		}

		userAgent := r.Header.Get("User-Agent")
		switch {
		case strings.Contains(userAgent, "Firefox"):
			data.UblockLink = "https://addons.mozilla.org/en-US/firefox/addon/ublock-origin/"
		case strings.Contains(userAgent, "Edg"):
			data.UblockLink = "https://microsoftedge.microsoft.com/addons/detail/ublock-origin/odfafepnkmbhccpbejgmiehpchacaeak"
		case strings.Contains(userAgent, "Chrome"):
			data.UblockLink = "https://chromewebstore.google.com/detail/ublock-origin-lite/ddkjiahejlhfcafbddmgiahcphecmpfh"
		default:
			data.UblockLink = "https://ublockorigin.com/"
		}

		renderTemplate(w, r, "index.gohtml", data)
	}
}
