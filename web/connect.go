package web

import (
	"net/http"
)

// ConnectTemplateData holds data for the connect page.
type ConnectTemplateData struct {
	BasePageData
}

func connectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := ConnectTemplateData{
			BasePageData: BasePageData{
				Title:        "How to Connect to a Minecraft Server",
				Description:  "Learn how to connect to any Minecraft server on Java and Bedrock editions. A step-by-step guide for new and experienced players.",
				CanonicalURL: "http://" + r.Host + r.URL.Path,
				OGImage:      "http://" + r.Host + "/logo.png",
				TwitterImage: "http://" + r.Host + "/logo.png",
			},
		}

		renderTemplate(w, r, "connect.gohtml", data)
	}
}
