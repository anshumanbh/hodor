package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type nmapscan struct {
	IPAddress string
	Port      string
	Protocol  string
	Service   string
	State     string
}

func main() {

	proj := os.Getenv("PROJECT_ID")
	subscriptionName := os.Getenv("SUBSCRIPTION_NAME")
	datasetName := os.Getenv("DATASET_NAME")
	tableName := os.Getenv("TABLE_NAME")
	bucketName := os.Getenv("BUCKET_NAME")

	ctx := context.Background()

	psclient, err := pubsub.NewClient(ctx, proj)
	check(err)

	subscription := psclient.Subscription(subscriptionName)

	bqclient, err := bigquery.NewClient(ctx, proj)
	check(err)

	ds := bqclient.Dataset(datasetName)

	for {
		msgit, err := subscription.Pull(ctx)
		check(err)
		msg, err := msgit.Next()
		if err == iterator.Done {
			fmt.Println("Done")
		}
		check(err)

		fmt.Printf("Got message: %q\n", string(msg.ID))

		rbytes := msg.Data

		fmt.Println("------------")
		fmt.Println(string(rbytes))
		fmt.Println("------------")

		schema, err := bigquery.InferSchema(nmapscan{})
		check(err)

		//TODO: check if already exits or not
		ds.Table(tableName).Create(ctx, schema)

		var gcsname bytes.Buffer
		gcsname.WriteString("gs://")
		gcsname.WriteString(bucketName)
		gcsname.WriteString("/")
		gcsname.WriteString(string(rbytes))
		fmt.Println(gcsname.String())

		gcsRef := bigquery.NewGCSReference(gcsname.String())
		loader := ds.Table(tableName).LoaderFrom(gcsRef)

		job, err := loader.Run(ctx)
		check(err)

		fmt.Printf("Job for data load operation: %+v\n", job)
		fmt.Printf("Waiting for job to complete.\n")

		for range time.Tick(5 * time.Second) {
			status, err := job.Status(ctx)
			if err != nil {
				fmt.Printf("Failure determining status: %v", err)
				break
			}
			if !status.Done() {
				continue
			}
			if err := status.Err(); err == nil {
				fmt.Printf("Success\n")
			} else {
				fmt.Printf("Failure: %+v\n", err)
			}
			break
		}

		time.Sleep(10 * time.Second)
	}

}
