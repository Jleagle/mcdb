package main

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"strings"
)

const pngPrefix = "data:image/png;base64,"

// Icon should be a PNG image that is Base64 encoded
// (without newlines: \n, new lines no longer work since 1.13)
// and prepended with "data:image/png;base64,".
type Icon string

func (i Icon) ToImage() (icon image.Image, err error) {

	if !strings.HasPrefix(string(i), pngPrefix) {
		return nil, fmt.Errorf("server icon should prepended with %q", pngPrefix)
	}

	base64png := strings.TrimPrefix(string(i), pngPrefix)
	r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64png))

	return png.Decode(r)
}
