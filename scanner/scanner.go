package scanner

import (
	"context"
	"fmt"
	"log"
	"net/netip"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
)

const maxGoroutines = 100

type Storage interface {
	SaveServer(s Server) error
	SaveLastIP(ip string) error
	LoadLastIP() string
	GetOldestServer() (Server, error)
}

func Start(store Storage) {
	guard := make(chan struct{}, maxGoroutines)
	bar := pb.StartNew(1 << 32)
	wg := sync.WaitGroup{}

	lastIPStr := store.LoadLastIP()
	var lastIP netip.Addr
	if lastIPStr != "" {
		lastIP, _ = netip.ParseAddr(lastIPStr)
	}

	prefix, err := netip.ParsePrefix("0.0.0.0/0")
	if err != nil {
		log.Fatal(err)
	}

	for addr := prefix.Addr(); prefix.Contains(addr); addr = addr.Next() {

		if addr.IsPrivate() || !addr.IsValid() || !addr.Is4() {
			bar.Increment()
			continue
		}

		if lastIP.IsValid() && addr.Compare(lastIP) <= 0 {
			bar.Increment()
			continue
		}

		wg.Add(1)
		guard <- struct{}{}
		go func(addr netip.Addr) {
			defer func() {
				bar.Increment()
				wg.Done()
				<-guard
			}()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			status, err := Probe(ctx, addr.String(), nil)
			if err != nil {
				return
			}
			if !status.IsOnline {
				return
			}

			err = store.SaveServer(*status)
			if err != nil {
				log.Printf("Failed to save server %s: %v", status.IP, err)
			} else {
				fmt.Printf("Found Minecraft server: %s\n", status.IP)
			}

			store.SaveLastIP(addr.String())

		}(addr)
	}

	wg.Wait()
	bar.Finish()
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
