package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Jleagle/mcdb/scanner"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	client     *mongo.Client
	db         *mongo.Database
	serversCol *mongo.Collection
	stateCol   *mongo.Collection
)

func InitDB() {
	connStr := os.Getenv("MCDB_MONGO")
	if connStr == "" {
		log.Fatal("MCDB_MONGO environment variable not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(options.Client().ApplyURI(connStr))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		if strings.Contains(err.Error(), "tls: internal error") || strings.Contains(err.Error(), "context deadline exceeded") {
			log.Fatalf("Could not connect to MongoDB. This usually means your IP is not whitelisted in MongoDB Atlas.\nError: %v", err)
		}
		log.Fatal(err)
	}

	db = client.Database("mcdb")
	serversCol = db.Collection("servers")
	stateCol = db.Collection("state")

	// Create indexes
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = serversCol.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: "data.location.geo", Value: "2dsphere"}},
		})
	}()
}

type mongoServer struct {
	ID        string         `bson:"_id"` // IP
	Data      scanner.Server `bson:"data"`
	UpdatedAt time.Time      `bson:"updated_at"`
}

func SaveServer(s scanner.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"data":       s,
			"updated_at": time.Now(),
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := serversCol.UpdateOne(ctx, bson.M{"_id": s.IP}, update, opts)
	return err
}

func SaveIP(ip string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Only insert if it doesn't exist
	update := bson.M{
		"$setOnInsert": bson.M{
			"data":       scanner.Server{IP: ip},
			"updated_at": time.Now(),
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := serversCol.UpdateOne(ctx, bson.M{"_id": ip}, update, opts)
	return err
}

type ListOptions struct {
	Limit  int64
	Offset int64
	Sort   string
	Lat    float64
	Lon    float64
	IP     string
	Name   string
	Tags   string
}

func ListServers(opts ListOptions) ([]scanner.Server, error) {
	ctx := context.Background()

	filter := listFilter(opts)

	total, err := CountServersWithOptions(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to count servers: %w", err)
	}

	var results []mongoServer

	if opts.Sort == "random" {
		sampleSize := int64(1000)
		if total < sampleSize {
			sampleSize = total
		}
		if sampleSize < opts.Limit {
			sampleSize = opts.Limit
		}
		if sampleSize == 0 && opts.Limit > 0 {
			sampleSize = opts.Limit
		}

		pipeline := mongo.Pipeline{
			bson.D{{Key: "$match", Value: filter}},
			bson.D{{Key: "$sample", Value: bson.M{"size": sampleSize}}},
			bson.D{{Key: "$skip", Value: opts.Offset}},
			bson.D{{Key: "$limit", Value: opts.Limit}},
		}

		cursor, err := serversCol.Aggregate(ctx, pipeline)
		if err != nil {
			return nil, fmt.Errorf("failed to aggregate servers for random sort: %w", err)
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &results); err != nil {
			return nil, fmt.Errorf("failed to decode aggregated servers: %w", err)
		}
	} else {
		findOpts := options.Find().
			SetLimit(opts.Limit).
			SetSkip(opts.Offset)

		switch opts.Sort {
		case "players":
			findOpts.SetSort(bson.D{{Key: "data.players.online", Value: -1}})
		case "max_players":
			findOpts.SetSort(bson.D{{Key: "data.players.max", Value: -1}})
		case "location":
			if opts.Lat != 0 || opts.Lon != 0 {
				filter["data.location.geo"] = bson.M{
					"$exists": true,
					"$near": bson.M{
						"$geometry": bson.M{
							"type":        "Point",
							"coordinates": []float64{opts.Lon, opts.Lat},
						},
					},
				}
			} else {
				filter["data.location.geo"] = bson.M{"$exists": true}
			}
		case "added_recently":
			findOpts.SetSort(bson.D{{Key: "data.updated_at", Value: -1}})
		default:
			findOpts.SetSort(bson.D{{Key: "data.updated_at", Value: -1}})
		}

		cursor, err := serversCol.Find(ctx, filter, findOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to find servers: %w", err)
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &results); err != nil {
			return nil, fmt.Errorf("failed to decode found servers: %w", err)
		}
	}

	var servers []scanner.Server
	for _, r := range results {
		if r.Data.IP == "" {
			r.Data.IP = r.ID
		}
		servers = append(servers, r.Data)
	}
	return servers, nil
}

func CountServers() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return serversCol.CountDocuments(ctx, bson.M{})
}

func CountServersWithOptions(opts ListOptions) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return serversCol.CountDocuments(ctx, listFilter(opts))
}

func CountPlayersOnline() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$data.players.online"},
		}}},
	}

	cursor, err := serversCol.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Total int64 `bson:"total"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return 0, err
	}
	if len(results) == 0 {
		return 0, nil
	}

	return results[0].Total, nil
}

func listFilter(opts ListOptions) bson.M {
	var andFilters []bson.M
	if opts.IP != "" {
		ipFilter := bson.M{"_id": bson.M{"$regex": opts.IP, "$options": "i"}}
		andFilters = append(andFilters, ipFilter)
	}
	if opts.Name != "" {
		nameFilter := bson.M{"data.version.name": bson.M{"$regex": opts.Name, "$options": "i"}}
		andFilters = append(andFilters, nameFilter)
	}
	if opts.Tags != "" {
		tags := strings.Split(opts.Tags, ",")
		tagFilters := make([]bson.M, len(tags))
		for i, tag := range tags {
			tagFilters[i] = bson.M{"data.tags": bson.M{"$regex": strings.TrimSpace(tag), "$options": "i"}}
		}
		andFilters = append(andFilters, bson.M{"$and": tagFilters})
	}

	if len(andFilters) == 0 {
		return bson.M{}
	}

	return bson.M{"$and": andFilters}
}

func GetOldestServer() (scanner.Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.FindOne().SetSort(bson.M{"updated_at": 1})
	var result mongoServer
	err := serversCol.FindOne(ctx, bson.M{}, opts).Decode(&result)
	if err != nil {
		return scanner.Server{}, err
	}

	if result.Data.IP == "" {
		result.Data.IP = result.ID
	}

	return result.Data, nil
}

func GetServer(ip string) (scanner.Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result mongoServer
	err := serversCol.FindOne(ctx, bson.M{"_id": ip}).Decode(&result)
	if err != nil {
		return scanner.Server{}, err
	}

	if result.Data.IP == "" {
		result.Data.IP = result.ID
	}

	return result.Data, nil
}

func SaveLastIP(ip string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Update().SetUpsert(true)
	_, err := stateCol.UpdateOne(ctx, bson.M{"_id": "last_ip"}, bson.M{"$set": bson.M{"value": ip}}, opts)
	return err
}

func LoadLastIP() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result lastIPState
	err := stateCol.FindOne(ctx, bson.M{"_id": "last_ip"}).Decode(&result)
	if err != nil {
		return ""
	}
	return result.Value
}

type TagCount struct {
	Name  string `bson:"_id"`
	Count int    `bson:"count"`
}

type lastIPState struct {
	Value string `bson:"value"`
}

func GetTags() ([]TagCount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$unwind", Value: "$data.tags"}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":   "$data.tags",
			"count": bson.M{"$sum": 1},
		}}},
		bson.D{{Key: "$sort", Value: bson.M{"count": -1}}},
	}

	cursor, err := serversCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []TagCount
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
