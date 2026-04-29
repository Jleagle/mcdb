package web

import (
	"net/http"
)

// SearchTemplateData holds data for the search page.
type SearchTemplateData struct {
	BasePageData
	IP   string
	Name string
	Tags string
}

func searchHandler(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := SearchTemplateData{
			BasePageData: BasePageData{
				Title:        "Advanced Minecraft Server Search",
				Description:  "Search Minecraft servers by IP, name, version, or tags using advanced filters.",
				CanonicalURL: "http://" + r.Host + r.URL.Path,
				OGImage:      "http://" + r.Host + "/logo.png",
				TwitterImage: "http://" + r.Host + "/logo.png",
			},
			IP:   r.URL.Query().Get("ip"),
			Name: r.URL.Query().Get("name"),
			Tags: r.URL.Query().Get("tags"),
		}

		renderTemplate(w, r, "search.gohtml", data)
	}
}
