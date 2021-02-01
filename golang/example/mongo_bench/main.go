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
var delete_num = flag.Int("delete_num", 5, "delete thread count")
var insert_num = flag.Int("insert_num", 10, "insert thread count")
var insertOpCount uint64 = 0
var deleteOpCount uint64 = 0

func delete(client *mongo.Client) {
	time.Sleep(time.Duration(*sleep_time) * time.Minute)

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
		atomic.AddUint64(&deleteOpCount, 1)
	}
}

func insertOp(client *mongo.Client, Json []byte, c chan int) {
	var bdoc interface{}
	err := bson.UnmarshalExtJSON(Json, false, &bdoc)
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database(*db_name).Collection(*col_name)
	_, err = collection.InsertOne(context.TODO(), &bdoc)
	if err != nil {
		log.Fatal(err)
	}
	atomic.AddUint64(&insertOpCount, 1)

	<-c
}

func insertBlock(client *mongo.Client) {
	file, err := os.Open(*data_path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	c := make(chan int, *insert_num)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		Json := []byte(scanner.Text())
		go insertOp(client, Json, c)
		c <- 1
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	log.Print("exit")
}

func statisticFunc() {
	callTicker := time.NewTicker(10 * time.Second)
	defer callTicker.Stop()

	var insertNum uint64 = 0
	var deleteNum uint64 = 0

	for {
		select {
		case <-callTicker.C:

			fmt.Printf("insert %d ops, delete %d ops, per 10s \n", insertOpCount-insertNum, deleteOpCount-deleteNum)
			insertNum = insertOpCount
			deleteNum = deleteOpCount
		}
	}
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
	clientOptions = clientOptions.SetMaxPoolSize(uint64(*insert_num + *delete_num))
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic(err)
	}

	go insertBlock(client)

	for i := 0; i < *delete_num; i++ {
		go delete(client)
	}

	go statisticFunc()

	hang()

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}
