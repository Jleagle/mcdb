package scanner

import (
	"context"
	"log"
	"time"

	"github.com/Jleagle/mcdb/storage"
)

type Storage interface {
	SaveServer(s storage.Server) error
	GetOldestServer() (storage.Server, error)
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
			log.Printf("\033[31mBackground updater: failed to probe %s: %v\033[0m", server.IP, err)
			// Mark as updated anyway so we don't loop on it
			server.IsOnline = false
			if err := store.SaveServer(server); err != nil {
				log.Printf("Background updater: failed to save %s: %v", server.IP, err)
			}
		} else {
			if err := store.SaveServer(*status); err != nil {
				log.Printf("Background updater: failed to save %s: %v", server.IP, err)
			}

			if status.IsOnline {
				log.Printf("\033[32mBackground updater: updated %s (Online)\033[0m", server.IP)
			} else {
				log.Printf("\033[31mBackground updater: updated %s (Offline)\033[0m", server.IP)
			}
		}

		time.Sleep(1 * time.Second)
	}
}
