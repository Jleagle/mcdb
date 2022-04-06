package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/netip"
	"os"
	"sync"

	"github.com/Tnze/go-mc/bot"
	mcnet "github.com/Tnze/go-mc/net"
	"github.com/cheggaaa/pb/v3"
)

const maxGoroutines = 10

func main() {

	// Get where we left off
	f, err := os.OpenFile("start.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = f.Close(); err != nil {
			log.Println(err)
		}
	}()

	b, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	savedAddress, err := netip.ParseAddr(string(b))
	if err != nil {
		log.Println(err)
	}

	prefix, err := netip.ParsePrefix("0.0.0.0/0")
	if err != nil {
		log.Fatal(err)
	}

	var latest netip.Addr

	defer func() {
		_, err = f.WriteString(latest.String())
		if err != nil {
			log.Println(err)
		}
	}()

	mcnet.DefaultDialer = mcnet.Dialer{Dialer: &net.Dialer{}}

	bar := pb.StartNew(1 << 32)
	guard := make(chan struct{}, maxGoroutines)
	wg := sync.WaitGroup{}

	for addr := prefix.Addr(); prefix.Contains(addr); addr = addr.Next() {

		if addr.IsPrivate() || !addr.IsValid() {
			continue
		}

		if savedAddress.Compare(addr) == 1 {
			continue
		}

		wg.Add(1)
		guard <- struct{}{}
		go func(addr netip.Addr) {

			defer func() {
				bar.Increment()
				wg.Done()
				<-guard
				if addr.Compare(latest) == 0 {
					latest = addr
				}
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
