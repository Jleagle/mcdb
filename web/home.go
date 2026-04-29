package web

import (
	"net/http"

	"github.com/Jleagle/mcdb/storage"
)

// HomeTemplateData holds data for the home page.
type HomeTemplateData struct {
	BasePageData
	TotalServers       int64
	TotalPlayersOnline int64
	Tags               []storage.TagCount
}

func homeHandler(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		total, _ := store.CountServers()
		playersOnline, _ := store.CountPlayersOnline()
		tags, _ := store.GetTags()

		data := HomeTemplateData{
			BasePageData: BasePageData{
				Title:        "Homepage",
				Description:  "Find the best Minecraft servers. Search by IP, name, version, or tags. Discover new servers with our comprehensive database.",
				CanonicalURL: "http://" + r.Host + r.URL.Path,
				OGImage:      "http://" + r.Host + "/logo.png",
				TwitterImage: "http://" + r.Host + "/logo.png",
			},
			TotalServers:       total,
			TotalPlayersOnline: playersOnline,
			Tags:               tags,
		}

		renderTemplate(w, r, "home.gohtml", data)
	}
}
