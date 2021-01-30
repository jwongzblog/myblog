package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var url = flag.String("url", "", "ip:port")
var db_name = flag.String("db_name", "", "db name")
var col_name = flag.String("col_name", "", "collection name")
var data_path = flag.String("data_path", "", "json data")
var sleep_time = flag.Int("sleep_time", 5, "sleep 5 minutes")
var delete_num = flag.Int("delete_num", 10, "delete thread count")
var insert_num = flag.Int("insert_num", 5, "insert thread count")
var workerNum int32 = 0

func delete(client *mongo.Client) {
	time.Sleep(*sleep_time * time.Minutes)

	for {
		collection := client.Database(*db_name).Collection(*col_name)
		opts := options.Delete().SetCollation(&options.Collation{
			Locale:    "en_US",
			Strength:  1,
			CaseLevel: false,
		})

		_, err := collection.DeleteOne(context.TODO(), bson.D{{"bucket_id", "59197b73-eb59-481e-a9f7-e96e0d23b3a1"}}, opts)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func insertNoBlock(client *mongo.Client) {
	file, err := os.Open(*data_path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		Json := scanner.Bytes()

		go func (client *mongo.Client, Json byte[]) {
			var bdoc interface{}
			err = bson.UnmarshalExtJSON(Json, false, &bdoc)
			if err != nil {
				log.Fatal(err)
				continue
			}
	
			collection := client.Database(*db_name).Collection(*col_name)
			_, err := collection.InsertOne(context.TODO(), &bdoc)
			if err != nil {
				log.Fatal(err)
			}
		}(client, Json)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	os.Exit()
}

func insertOp(client *mongo.Client, Json byte[], c chan string) {
	var bdoc interface{}
	err = bson.UnmarshalExtJSON(Json, false, &bdoc)
	if err != nil {
		log.Fatal(err)
		continue
	}

	collection := client.Database(*db_name).Collection(*col_name)
	_, err := collection.InsertOne(context.TODO(), &bdoc)
	if err != nil {
		log.Fatal(err)
	}
	
	// lock
	var mutex sync.Mutex
	mutex.Lock()
	atomic.AddInt32(&workerNum, -1)
	//whether send chan
	if workerNum == *insert_num - 1 {
		c <- "go on"
	}
	// unlock
	mutex.Unlock()
}

func insertBlock(client *mongo.Client) {
	file, err := os.Open(*data_path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	c := make(chan string, 1)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		atomic.AddInt32(&workerNum, 1)
		Json := scanner.Bytes()

		if *insert_num > workerNum {
			go insertOp(client, Json, c)
			continue
		}

		for {
			select {
			case  <- c:
				go insertOp(client, Json, c)	
			}			
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	os.Exit()
}

func hang() {
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s", *url))
	clientOptions = clientOptions.SetMaxPoolSize(*insert_num + *delete_num)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic(err)
	}

	go insertNoBlock(client)

	for i := 0; i < *delete_num; i++ {
		go delete(client)
	}

	hang()

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}
