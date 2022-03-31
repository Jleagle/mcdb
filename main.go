package main

import (
	"encoding/json"
	"fmt"
	_ "image/jpeg"
	"net"
	"net/netip"
	"sync"

	"github.com/Tnze/go-mc/bot"
	mcnet "github.com/Tnze/go-mc/net"
	"github.com/cheggaaa/pb/v3"
)

const maxGoroutines = 10

func main() {

	prefix, err := netip.ParsePrefix("0.0.0.0/0")
	if err != nil {
		fmt.Printf("Failed to parse prefix: %v", err)
		return
	}

	mcnet.DefaultDialer = mcnet.Dialer{Dialer: &net.Dialer{}}

	bar := pb.StartNew(1 << 32)
	guard := make(chan struct{}, maxGoroutines)
	wg := sync.WaitGroup{}

	for addr := prefix.Addr(); prefix.Contains(addr); addr = addr.Next() {

		if addr.IsPrivate() || !addr.IsValid() {
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

			resp, delay, err := bot.PingAndList(fmt.Sprintf("%s:%d", addr.String(), 25565))
			if err != nil {

				if _, ok := err.(bot.LoginErr); ok {
					return
				}

				fmt.Println("Ping and list server fail: ", err)
				return
			}

			var s status
			err = json.Unmarshal(resp, &s)
			if err != nil {
				fmt.Println("Parse json response fail:", err)
				return
			}

			s.Delay = delay

			fmt.Println(addr.String() + " " + s.Description.Text)

		}(addr)
	}

	wg.Wait()
	bar.Finish()
}
