package web

import (
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Jleagle/mcdb/scanner"
	"github.com/Jleagle/mcdb/storage"
)

// BasePageData contains common fields for all pages for SEO.
type BasePageData struct {
	Title        string
	Description  string
	CanonicalURL string
	OGImage      string
	TwitterImage string
}

type Storage interface {
	ListServers(opts storage.ListOptions) ([]scanner.Server, error)
	GetServer(ip string) (scanner.Server, error)
	CountServers() (int64, error)
	CountServersWithOptions(opts storage.ListOptions) (int64, error)
	CountPlayersOnline() (int64, error)
	GetTags() ([]storage.TagCount, error)
}

type templateContext struct {
	Data       interface{}
	BaseURL    string
	CurrentNav string
}

func Start(store Storage) {
	mux := http.NewServeMux()

	// Register application handlers
	mux.HandleFunc("/", homeHandler(store))
	mux.HandleFunc("/servers", indexHandler(store))
	mux.HandleFunc("/search", searchHandler(store))
	mux.HandleFunc("/server/", serverHandler(store))
	mux.HandleFunc("/connect", connectHandler())

	// Register asset handler
	RegisterAssetHandler(mux)

	log.Println("Starting web server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func renderTemplate(w http.ResponseWriter, r *http.Request, tmplName string, data interface{}) {
	funcMap := template.FuncMap{
		"safe": func(s any) template.URL {
			return template.URL(fmt.Sprint(s))
		},
		"formatNumber": formatNumber,
	}

	templateData := templateContext{
		Data:       data,
		BaseURL:    "https://" + r.Host,
		CurrentNav: currentNav(r.URL.Path),
	}

	tmpl, err := template.New("layout.gohtml").Funcs(funcMap).ParseFiles(
		filepath.Join("web", "templates", "layout.gohtml"),
		filepath.Join("web", "templates", tmplName),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "layout.gohtml", templateData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func currentNav(path string) string {
	switch {
	case path == "/":
		return "home"
	case path == "/servers" || strings.HasPrefix(path, "/server/"):
		return "servers"
	case path == "/search":
		return "search"
	case path == "/connect":
		return "connect"
	default:
		return ""
	}
}

func formatNumber(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var out []byte
	firstGroupLen := len(s) % 3
	if firstGroupLen == 0 {
		firstGroupLen = 3
	}

	out = append(out, s[:firstGroupLen]...)
	for i := firstGroupLen; i < len(s); i += 3 {
		out = append(out, ',')
		out = append(out, s[i:i+3]...)
	}

	return string(out)
}

// haversineDistance calculates the distance between two points on the Earth
// given their latitudes and longitudes. Returns distance in kilometers.
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers

	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
