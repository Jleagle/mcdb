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
				CanonicalURL: "https://" + r.Host + r.URL.Path,
				OGImage:      "https://" + r.Host + "/logo.png",
				TwitterImage: "https://" + r.Host + "/logo.png",
			},
		}

		renderTemplate(w, r, "connect.gohtml", data)
	}
}
