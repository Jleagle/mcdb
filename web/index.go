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
	Servers           []ServerWithDistance
	Page              int
	HasPrev           bool
	HasNext           bool
	PrevPage          int
	NextPage          int
	TotalPages        int
	TotalResults      int64
	CurrentSort       string
	UserLat           float64
	UserLon           float64
	UserLocationName  string
	CurrentIPSearch   string
	CurrentNameSearch string
	CurrentTagsSearch string
	UblockLink        string // New field for the uBlock Origin link
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

		limit := int64(20)
		offset := int64(page-1) * limit

		opts := storage.ListOptions{
			Limit:  limit,
			Offset: offset,
			Sort:   sort,
			IP:     ipSearch,
			Name:   nameSearch,
			Tags:   tagsSearch,
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
		var serversWithDistance []ServerWithDistance
		if sort == "location" && opts.Lat != 0 && opts.Lon != 0 {
			serversWithDistance = make([]ServerWithDistance, len(servers))
			for i, s := range servers {
				sd := ServerWithDistance{Server: s} // Initialize with the scanner.Server
				if s.Location != nil && s.Location.Lat != 0 && s.Location.Lon != 0 {
					sd.DistanceKM = haversineDistance(opts.Lat, opts.Lon, s.Location.Lat, s.Location.Lon)
				} else {
					sd.DistanceKM = -1.0 // Indicate no distance could be calculated
				}
				serversWithDistance[i] = sd
			}
		} else {
			// If not sorting by location, just wrap the existing servers
			serversWithDistance = make([]ServerWithDistance, len(servers))
			for i, s := range servers {
				serversWithDistance[i] = ServerWithDistance{Server: s}
			}
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

		ublockLink := "https://ublockorigin.com/" // Default fallback
		userAgent := r.Header.Get("User-Agent")
		if strings.Contains(userAgent, "Firefox") {
			ublockLink = "https://addons.mozilla.org/en-US/firefox/addon/ublock-origin/"
		} else if strings.Contains(userAgent, "Edg") {
			ublockLink = "https://microsoftedge.microsoft.com/addons/detail/ublock-origin/odfafepnkmbhccpbejgmiehpchacaeak"
		} else if strings.Contains(userAgent, "Chrome") {
			ublockLink = "https://chromewebstore.google.com/detail/ublock-origin-lite/ddkjiahejlhfcafbddmgiahcphecmpfh"
		}

		data := IndexTemplateData{
			BasePageData: BasePageData{
				Title:        "Minecraft Server List",
				Description:  "Browse all Minecraft servers, sorted by " + sort + ".",
				CanonicalURL: "http://" + r.Host + r.URL.RequestURI(),
				OGImage:      "http://" + r.Host + "/logo.png",
				TwitterImage: "http://" + r.Host + "/logo.png",
			},
			Servers:           serversWithDistance,
			Page:              page,
			HasPrev:           page > 1,
			HasNext:           offset+limit < total,
			PrevPage:          page - 1,
			NextPage:          page + 1,
			TotalPages:        totalPages,
			TotalResults:      total,
			CurrentSort:       sort,
			UserLat:           opts.Lat,
			UserLon:           opts.Lon,
			UserLocationName:  userLocationName,
			CurrentIPSearch:   ipSearch,
			CurrentNameSearch: nameSearch,
			CurrentTagsSearch: tagsSearch,
			UblockLink:        ublockLink, // Pass the determined link
		}

		renderTemplate(w, r, "index.gohtml", data)
	}
}
