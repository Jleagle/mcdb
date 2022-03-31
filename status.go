package main

import (
	"strings"
	"text/template"
	"time"

	"github.com/Tnze/go-mc/chat"
	"github.com/google/uuid"
)

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

var outTemp = template.Must(template.New("output").Parse(`
	Version: [{{ .Version.Protocol }}] {{ .Version.Name }}
	Description: {{ .Description }}
	Delay: {{ .Delay }}
	Players: {{ .Players.Online }}/{{ .Players.Max }}{{ range .Players.Sample }}
	- [{{ .Name }}] {{ .ID }}{{ end }}
`))

func (s *status) String() (string, error) {

	var sb strings.Builder
	err := outTemp.Execute(&sb, s)
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}
