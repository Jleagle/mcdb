package scanner

import (
	"context"
	"log"
	"time"
)

type Storage interface {
	SaveServer(s Server) error
	GetOldestServer() (Server, error)
}

func Updater(store Storage) {
	for {
		server, err := store.GetOldestServer()
		if err != nil {
			log.Printf("Background updater: failed to get oldest server: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		status, err := Probe(ctx, server.IP, server.Location)
		cancel()

		if err != nil {
			log.Printf("Background updater: failed to probe %s: %v", server.IP, err)
			// Mark as updated anyway so we don't loop on it
			server.IsOnline = false
			store.SaveServer(server)
		} else {
			store.SaveServer(*status)
			if status.IsOnline {
				log.Printf("Background updater: updated %s (Online)", server.IP)
			} else {
				log.Printf("Background updater: updated %s (Offline)", server.IP)
			}
		}

		time.Sleep(3 * time.Second)
	}
}
