package web

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

type URL struct {
	XMLName    xml.Name `xml:"url"`
	Loc        string   `xml:"loc"`
	LastMod    string   `xml:"lastmod,omitempty"`
	ChangeFreq string   `xml:"changefreq,omitempty"`
	Priority   float64  `xml:"priority,omitempty"`
}

type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

func sitemapHandler(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scheme := "https"
		if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
			scheme = "http"
		}
		host := r.Host
		baseURL := fmt.Sprintf("%s://%s", scheme, host)

		urls := []URL{
			{Loc: baseURL + "/", ChangeFreq: "daily", Priority: 1.0},
			{Loc: baseURL + "/servers", ChangeFreq: "daily", Priority: 0.9},
			{Loc: baseURL + "/search", ChangeFreq: "monthly", Priority: 0.7},
			{Loc: baseURL + "/connect", ChangeFreq: "monthly", Priority: 0.7},
		}

		ips, err := store.GetServerIPs()
		if err == nil {
			for _, entry := range ips {
				urls = append(urls, URL{
					Loc:        fmt.Sprintf("%s/server/%s", baseURL, entry.IP),
					LastMod:    entry.UpdatedAt.Format("2006-01-02"),
					ChangeFreq: "weekly",
					Priority:   0.5,
				})
			}
		}

		sitemap := URLSet{
			XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
			URLs:  urls,
		}

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(xml.Header))
		xml.NewEncoder(w).Encode(sitemap)
	}
}
