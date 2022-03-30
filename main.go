package main

import (
	"encoding/json"
	"fmt"
	_ "image/jpeg"
	"net/netip"
	"os"
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
