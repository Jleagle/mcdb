package seeder

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func StartMinecraftMP(store Storage) {
	page := 1
	userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	for {
		url := "https://minecraft-mp.com/"
		if page > 1 {
			url = fmt.Sprintf("https://minecraft-mp.com/servers/list/%d/", page)
		}

		fmt.Printf("Crawling %s...\n", url)

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("User-Agent", userAgent)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to fetch %s: %v", url, err)
			break
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Got status %d for %s", resp.StatusCode, url)
			resp.Body.Close()
			break
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("Failed to parse %s: %v", url, err)
			break
		}

		found := 0
		doc.Find("button.clipboard").Each(func(i int, s *goquery.Selection) {
			ip, exists := s.Attr("data-clipboard-text")
			if exists && ip != "" {
				isNew, err := store.SaveIP(ip)
				if err != nil {
					log.Printf("Failed to save IP %s: %v", ip, err)
				} else {
					if isNew {
						fmt.Printf("Seeded new: %s\n", ip)
					} else {
						fmt.Printf("Seeded existing: %s\n", ip)
					}
					found++
				}
			}
		})

		fmt.Printf("Found and saved %d IPs from page %d\n", found, page)

		if found == 0 {
			fmt.Println("No more IPs found, stopping.")
			break
		}

		page++
		time.Sleep(1 * time.Second) // Be nice
	}
}
