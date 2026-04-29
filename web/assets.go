package web

import (
	"net/http"
	"os"
)

// RegisterAssetHandler registers the handler for static assets.
func RegisterAssetHandler(mux *http.ServeMux) {
	entries, err := os.ReadDir("assets")
	if err != nil {
		return
	}

	fs := http.FileServer(http.Dir("assets"))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		mux.Handle("/"+entry.Name(), fs)
	}
}
