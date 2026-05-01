package storage

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestStorage(t *testing.T) {
	InitDB()

	ctx := context.Background()
	serversCol.DeleteOne(ctx, bson.M{"_id": "1.2.3.4"})
	stateCol.DeleteOne(ctx, bson.M{"_id": "test_last_ip"})

	s := Server{
		IP: "1.2.3.4",
	}
	s.Version.Name = "1.20.1"

	err := SaveServer(s)
	if err != nil {
		t.Fatalf("Failed to save server: %v", err)
	}

	servers, err := ListServers(ListOptions{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("Failed to list servers: %v", err)
	}

	found := false
	for _, srv := range servers {
		if srv.IP == "1.2.3.4" {
			found = true
			if srv.Version.Name != "1.20.1" {
				t.Errorf("Expected version 1.20.1, got %s", srv.Version.Name)
			}
			break
		}
	}
	if !found {
		t.Errorf("Expected to find server 1.2.3.4 in the list")
	}

	count, err := CountServers()
	if err != nil {
		t.Fatalf("Failed to count servers: %v", err)
	}
	if count < 1 {
		t.Errorf("Expected count to be at least 1, got %d", count)
	}

	err = SaveLastIP("1.2.3.5")
	if err != nil {
		t.Fatalf("Failed to save last IP: %v", err)
	}

	lastIP := LoadLastIP()
	if lastIP != "1.2.3.5" {
		t.Errorf("Expected last IP 1.2.3.5, got %s", lastIP)
	}
}
