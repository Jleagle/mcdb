package seeder

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func StartMinecraftServerList(store Storage) {
	page := 1
	client := &http.Client{Timeout: 30 * time.Second}
	userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	for {
		pageURL := "https://minecraft-server-list.com/"
		if page > 1 {
			pageURL = fmt.Sprintf("https://minecraft-server-list.com/page/%d/", page)
		}

		fmt.Printf("Crawling %s...\n", pageURL)

		doc, err := fetchDocument(client, pageURL, userAgent)
		if err != nil {
			log.Printf("Failed to fetch %s: %v", pageURL, err)
			break
		}

		serverURLs := extractMinecraftServerListServerURLs(doc, pageURL)
		if len(serverURLs) == 0 {
			fmt.Println("No server detail links found, stopping.")
			break
		}

		found := 0
		for _, serverURL := range serverURLs {
			ip, err := fetchMinecraftServerListIP(client, serverURL, userAgent)
			if err != nil {
				log.Printf("Failed to fetch server details %s: %v", serverURL, err)
				continue
			}
			if ip == "" {
				log.Printf("No Java IP found on %s", serverURL)
				continue
			}

			err = store.SaveIP(ip)
			if err != nil {
				log.Printf("Failed to save IP %s: %v", ip, err)
			} else {
				fmt.Printf("Seeded: %s\n", ip)
				found++
			}
		}

		fmt.Printf("Found and saved %d IPs from page %d\n", found, page)

		if found == 0 {
			fmt.Println("No more IPs found, stopping.")
			break
		}

		page++
		time.Sleep(1 * time.Second)
	}
}

func extractMinecraftServerListServerURLs(doc *goquery.Document, pageURL string) []string {
	baseURL, err := url.Parse(pageURL)
	if err != nil {
		return nil
	}

	seen := make(map[string]bool)
	var serverURLs []string
	serverPath := regexp.MustCompile(`^/server/[0-9]+/?$`)

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		link, err := url.Parse(strings.TrimSpace(href))
		if err != nil {
			return
		}

		absolute := baseURL.ResolveReference(link)
		if absolute.Host != "minecraft-server-list.com" || !serverPath.MatchString(absolute.Path) {
			return
		}

		absolute.RawQuery = ""
		absolute.Fragment = ""
		normalized := absolute.String()
		if seen[normalized] {
			return
		}

		seen[normalized] = true
		serverURLs = append(serverURLs, normalized)
	})

	return serverURLs
}

func fetchMinecraftServerListIP(client *http.Client, serverURL, userAgent string) (string, error) {
	doc, err := fetchDocument(client, serverURL, userAgent)
	if err != nil {
		return "", err
	}

	return extractMinecraftServerListIP(doc), nil
}

func extractMinecraftServerListIP(doc *goquery.Document) string {
	for _, attr := range []string{"data-clipboard-text", "data-ip", "data-copy"} {
		if ip := firstAttributeIP(doc, attr); ip != "" {
			return ip
		}
	}

	text := normalizeWhitespace(doc.Text())
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)Java\s+Server\s+IP\s*[^a-z0-9.:-]*\s*([a-z0-9][a-z0-9.-]*(?::[0-9]{1,5})?)`),
		regexp.MustCompile(`(?i)Java\s+IP\s*[^a-z0-9.:-]*\s*([a-z0-9][a-z0-9.-]*(?::[0-9]{1,5})?)`),
	}

	for _, pattern := range patterns {
		match := pattern.FindStringSubmatch(text)
		if len(match) == 2 {
			return cleanIP(match[1])
		}
	}

	return ""
}

func firstAttributeIP(doc *goquery.Document, attr string) string {
	var ip string
	doc.Find("[" + attr + "]").EachWithBreak(func(i int, s *goquery.Selection) bool {
		value, _ := s.Attr(attr)
		ip = cleanIP(value)
		return ip == ""
	})
	return ip
}

func fetchDocument(client *http.Client, pageURL, userAgent string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	if strings.Contains(strings.ToLower(doc.Find("title").Text()), "just a moment") {
		return nil, fmt.Errorf("received Cloudflare challenge page")
	}

	return doc, nil
}

func cleanIP(ip string) string {
	ip = strings.TrimSpace(ip)
	ip = strings.Trim(ip, " \t\r\n.,;()[]{}")
	return ip
}

func normalizeWhitespace(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
