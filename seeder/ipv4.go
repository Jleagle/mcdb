package seeder

import (
	"context"
	"fmt"
	"log"
	"net/netip"
	"sync"
	"time"

	"github.com/Jleagle/mcdb/scanner"
	"github.com/cheggaaa/pb/v3"
)

const maxGoroutines = 100

func StartIPv4(store Storage) {
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

			status, err := scanner.Probe(ctx, addr.String(), nil)
			if err != nil {
				return
			}
			if !status.IsOnline {
				return
			}

			err = store.SaveIP(addr.String())
			if err != nil {
				log.Printf("Failed to save IP %s: %v", addr.String(), err)
			} else {
				fmt.Printf("Found and seeded Minecraft server: %s\n", addr.String())
			}

			store.SaveLastIP(addr.String())

		}(addr)
	}

	wg.Wait()
	bar.Finish()
}
