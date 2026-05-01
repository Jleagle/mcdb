package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Jleagle/mcdb/scanner"
	"github.com/Jleagle/mcdb/seeder"
	"github.com/Jleagle/mcdb/storage"
	"github.com/Jleagle/mcdb/web"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mcdb",
		Short: "Minecraft Server Database",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			storage.InitDB()
		},
	}

	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the web server",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting web server and background updater...")
			go scanner.Updater(storageInterface{})
			web.Start(storageInterface{})
		},
	}

	var seedCmd = &cobra.Command{
		Use:   "seed",
		Short: "Seed the database from known Minecraft server lists",
	}

	var seedIPv4Cmd = &cobra.Command{
		Use:   "ipv4",
		Short: "Seed the database by scanning the IPv4 space",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting IPv4 seeder...")
			seeder.StartIPv4(storageInterface{})
		},
	}

	var seedMinecraftMPCmd = &cobra.Command{
		Use:   "minecraft-mp",
		Short: "Seed the database by crawling minecraft-mp.com",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting minecraft-mp seeder...")
			seeder.StartMinecraftMP(storageInterface{})
		},
	}

	var seedMinecraftServerListCmd = &cobra.Command{
		Use:   "minecraft-server-list",
		Short: "Seed the database by crawling minecraft-server-list.com",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting minecraft-server-list seeder...")
			seeder.StartMinecraftServerList(storageInterface{})
		},
	}
	seedCmd.AddCommand(seedIPv4Cmd, seedMinecraftMPCmd, seedMinecraftServerListCmd)

	var probeCmd = &cobra.Command{
		Use:   "probe [host]",
		Short: "Probe a specific server for Java, Bedrock and Query data",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			host := args[0]
			fmt.Printf("Probing %s...\n", host)

			status, err := scanner.Probe(context.Background(), host, nil)
			if err != nil {
				fmt.Printf("Error probing: %v\n", err)
				return
			}

			if status == nil {
				fmt.Println("No Minecraft server found at this address.")
				return
			}

			err = storage.SaveServer(*status)
			if err != nil {
				fmt.Printf("Error saving to DB: %v\n", err)
				return
			}

			fmt.Printf("Successfully probed and saved %s\n", host)
			out, _ := json.MarshalIndent(status, "", "  ")
			fmt.Println(string(out))
		},
	}

	rootCmd.AddCommand(serveCmd, seedCmd, probeCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// storageInterface adapts our storage package to the interfaces required by scanner, web and seeder
type storageInterface struct{}

func (s storageInterface) SaveServer(status storage.Server) error {
	return storage.SaveServer(status)
}

func (s storageInterface) SaveIP(ip string) (bool, error) {
	return storage.SaveIP(ip)
}

func (s storageInterface) SaveLastIP(ip string) error {
	return storage.SaveLastIP(ip)
}

func (s storageInterface) LoadLastIP() string {
	return storage.LoadLastIP()
}

func (s storageInterface) ListServers(opts storage.ListOptions) ([]storage.Server, error) {
	return storage.ListServers(opts)
}

func (s storageInterface) GetOldestServer() (storage.Server, error) {
	return storage.GetOldestServer()
}

func (s storageInterface) GetServer(ip string) (storage.Server, error) {
	return storage.GetServer(ip)
}

func (s storageInterface) GetServerIPs() ([]storage.IPWithDate, error) {
	return storage.GetServerIPs()
}

func (s storageInterface) CountServers() (int64, error) {
	return storage.CountServers()
}

func (s storageInterface) CountServersWithOptions(opts storage.ListOptions) (int64, error) {
	return storage.CountServersWithOptions(opts)
}

func (s storageInterface) CountPlayersOnline() (int64, error) {
	return storage.CountPlayersOnline()
}

func (s storageInterface) GetTags() ([]storage.TagCount, error) {
	return storage.GetTags()
}

func (s storageInterface) GetCountries() ([]storage.CountryCount, error) {
	return storage.GetCountries()
}

func (s storageInterface) GetVersions() ([]storage.VersionCount, error) {
	return storage.GetVersions()
}
