package web

import (
	"net/http"
)

// SearchTemplateData holds data for the search page.
type SearchTemplateData struct {
	BasePageData
	IP      string
	Name    string
	Tags    string
	Version string
	Country string
	Privacy string
}

func searchHandler(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := SearchTemplateData{
			BasePageData: BasePageData{
				Title:        "Advanced Minecraft Server Search",
				Description:  "Search Minecraft servers by IP, name, version, or tags using advanced filters.",
				CanonicalURL: "https://" + r.Host + r.URL.Path,
				OGImage:      "https://" + r.Host + "/logo.png",
				TwitterImage: "https://" + r.Host + "/logo.png",
			},
			IP:      r.URL.Query().Get("ip"),
			Name:    r.URL.Query().Get("name"),
			Tags:    r.URL.Query().Get("tags"),
			Version: r.URL.Query().Get("version"),
			Country: r.URL.Query().Get("country"),
			Privacy: r.URL.Query().Get("privacy"),
		}

		renderTemplate(w, r, "search.gohtml", data)
	}
}
