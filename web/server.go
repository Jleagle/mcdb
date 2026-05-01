package web

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Jleagle/mcdb/storage"
)

// ServerTemplateData holds data for the server detail page.
type ServerTemplateData struct {
	BasePageData
	storage.Server
	RawJSON string
}

func serverHandler(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ip := strings.TrimPrefix(r.URL.Path, "/server/")
		if ip == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		server, err := store.GetServer(ip)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		raw, err := json.MarshalIndent(server, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling server data:", err)
		}

		description := fmt.Sprintf("Details for Minecraft server %s running version %s with %d/%d players.",
			server.IP, server.Version.Name, server.Players.Online, server.Players.Max)
		if server.Location != nil {
			description += fmt.Sprintf(" Located in %s, %s.", server.Location.City, server.Location.Country)
		}

		title := fmt.Sprintf("%s - Minecraft Server (%s) | %d/%d Players",
			server.IP, server.Version.Name, server.Players.Online, server.Players.Max)

		data := ServerTemplateData{
			BasePageData: BasePageData{
				Title:        title,
				Description:  description,
				CanonicalURL: "https://" + r.Host + r.URL.Path,
				OGImage:      cmp.Or(string(server.Favicon), "https://"+r.Host+"/logo.png"),
				TwitterImage: cmp.Or(string(server.Favicon), "https://"+r.Host+"/logo.png"),
			},
			Server:  server,
			RawJSON: string(raw),
		}

		renderTemplate(w, r, "server.gohtml", data)
	}
}
