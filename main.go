package main

import (
	"bufio"
	"encoding/json"
	"fmt"
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

var logFile *os.File

func main() {

	mcnet.DefaultDialer = mcnet.Dialer{Dialer: &net.Dialer{}}

	var err error
	logFile, err = os.OpenFile("servers.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer log.Fatal(logFile.Close())

	guard := make(chan struct{}, maxGoroutines)
	bar := pb.StartNew(1 << 32)
	wg := sync.WaitGroup{}
	save := load()

	prefix, err := netip.ParsePrefix("0.0.0.0/0")
	if err != nil {
		log.Fatal(err)
	}

	for addr := prefix.Addr(); prefix.Contains(addr); addr = addr.Next() {

		wg.Add(1)
		guard <- struct{}{}
		go func(addr netip.Addr) {

			defer func() {
				bar.Increment()
				wg.Done()
				<-guard
				if addr.Compare(save) == 1 {
					_, err := logFile.WriteString(addr.String())
					if err != nil {
						log.Println(err)
					}
				}
			}()

			if addr.IsPrivate() || !addr.IsValid() {
				return
			}

			if save.Compare(addr) == 1 {
				return
			}

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

func load() netip.Addr {

	// Get last line
	var last string
	scanner := bufio.NewScanner(logFile)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			last = line
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	addr, err := netip.ParseAddr(last)
	if err != nil {
		log.Println(err)
	}

	return addr
}
