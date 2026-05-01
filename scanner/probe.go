package scanner

import (
	"context"
	"log"
	"time"

	"github.com/Jleagle/mcdb/storage"
	"github.com/mcstatus-io/mcutil/v2"
)

func Probe(ctx context.Context, host string, loc *storage.Location) (*storage.Server, error) {
	s := &storage.Server{
		IP:        host,
		UpdatedAt: time.Now(),
		IsOnline:  false,
	}

	found := false

	// 1. Try Java Status
	javaStatus, err := mcutil.Status(ctx, host, 25565)
	if err == nil {
		s.IsJava = true
		s.Java = javaStatus
		found = true

		// Map to legacy fields for web UI compatibility
		s.Version.Name = javaStatus.Version.NameClean
		s.Version.Protocol = int(javaStatus.Version.Protocol)
		if javaStatus.Players.Online != nil {
			s.Players.Online = int(*javaStatus.Players.Online)
		}
		if javaStatus.Players.Max != nil {
			s.Players.Max = int(*javaStatus.Players.Max)
		}
		if javaStatus.Favicon != nil {
			s.Favicon = storage.Icon(*javaStatus.Favicon)
		}
		s.Delay = javaStatus.Latency
	}

	// 2. Try Bedrock Status
	bedrockStatus, err := mcutil.StatusBedrock(ctx, host, 19132)
	if err == nil {
		s.IsBedrock = true
		s.Bedrock = bedrockStatus
		found = true
	}

	// 3. Try Query (GS4)
	queryStatus, err := mcutil.FullQuery(ctx, host, 25565)
	if err == nil {
		s.HasQuery = true
		s.Query = queryStatus
		found = true
	} else {
		// Try Bedrock port for query
		queryStatus, err = mcutil.FullQuery(ctx, host, 19132)
		if err == nil {
			s.HasQuery = true
			s.Query = queryStatus
			found = true
		}
	}

	s.IsOnline = found
	s.Tags = s.GetTags()

	// Try to get location if online and not provided
	if s.IsOnline && loc == nil {
		newLoc, err := GetLocation(host)
		if err == nil {
			s.Location = newLoc
		} else {
			log.Printf("Failed to get location for %s: %v", host, err)
		}
	} else if loc != nil {
		s.Location = loc
	}

	return s, nil
}
