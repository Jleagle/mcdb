package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"net/netip"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/google/uuid"
)

const maxGoroutines = 10

func main() {

	guard := make(chan struct{}, maxGoroutines)

	prefix, err := netip.ParsePrefix("0.0.0.0/0")
	if err != nil {
		panic(err)
	}

	for addr := prefix.Addr(); prefix.Contains(addr); addr = addr.Next() {

		guard <- struct{}{}
		go func(addr netip.Addr) {

			resp, delay, err := bot.PingAndList(addr.String())
			if err != nil {
				fmt.Printf("Ping and list server fail: %v", err)
				os.Exit(1)
			}

			var s status
			err = json.Unmarshal(resp, &s)
			if err != nil {
				fmt.Print("Parse json response fail:", err)
				os.Exit(1)
			}
			s.Delay = delay

			fmt.Println(s.String())

			<-guard
		}(addr)
	}
}

type status struct {
	Description chat.Message
	Players     struct {
		Max    int
		Online int
		Sample []struct {
			ID   uuid.UUID
			Name string
		}
	}
	Version struct {
		Name     string
		Protocol int
	}
	Favicon Icon
	Delay   time.Duration
}

// Icon should be a PNG image that is Base64 encoded
// (without newlines: \n, new lines no longer work since 1.13)
// and prepended with "data:image/png;base64,".
type Icon string

func (i Icon) ToImage() (icon image.Image, err error) {
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(string(i), prefix) {
		return nil, fmt.Errorf("server icon should prepended with %q", prefix)
	}
	base64png := strings.TrimPrefix(string(i), prefix)
	r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64png))
	icon, err = png.Decode(r)
	return
}

var outTemp = template.Must(template.New("output").Parse(`
	Version: [{{ .Version.Protocol }}] {{ .Version.Name }}
	Description: 
{{ .Description }}
	Delay: {{ .Delay }}
	Players: {{ .Players.Online }}/{{ .Players.Max }}{{ range .Players.Sample }}
	- [{{ .Name }}] {{ .ID }}{{ end }}
`))

func (s *status) String() string {
	var sb strings.Builder
	err := outTemp.Execute(&sb, s)
	if err != nil {
		panic(err)
	}
	return sb.String()
}
