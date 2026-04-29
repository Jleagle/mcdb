package seeder

import (
	"reflect"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractMinecraftServerListServerURLs(t *testing.T) {
	doc := mustDocument(t, `
		<a href="/server/411920/">Complex Gaming</a>
		<a href="https://minecraft-server-list.com/server/12345/">Another Server</a>
		<a href="/server/411920/?ref=duplicate">Duplicate Server</a>
		<a href="/servers/survival/">Survival</a>
		<a href="https://example.com/server/999/">External</a>
	`)

	got := extractMinecraftServerListServerURLs(doc, "https://minecraft-server-list.com/page/2/")
	want := []string{
		"https://minecraft-server-list.com/server/411920/",
		"https://minecraft-server-list.com/server/12345/",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}

func TestExtractMinecraftServerListIP(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "java server ip text",
			html: `<body>Java Server IP &#10233;mslc.mc-complex.com Bedrock IP:Port &#10233; mslc.mc-complex.com:25565</body>`,
			want: "mslc.mc-complex.com",
		},
		{
			name: "clipboard attribute",
			html: `<button data-clipboard-text="play.example.net:25565">Copy IP</button>`,
			want: "play.example.net:25565",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := mustDocument(t, tt.html)
			if got := extractMinecraftServerListIP(doc); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func mustDocument(t *testing.T, html string) *goquery.Document {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("failed to parse document: %v", err)
	}

	return doc
}
