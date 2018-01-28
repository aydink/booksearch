package main

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/olivere/elastic"
)

func main() {
	err := OpenConnection()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = createIndex("book", "schema_book.json")
	if err != nil {
		fmt.Println(err)
	}
}

var client *elastic.Client

func OpenConnection() error {
	url := "http://127.0.0.1:9200"

	//Create an Elasticsearch client
	c, err := elastic.NewClient(elastic.SetURL(url), elastic.SetSniff(true))
	if err != nil {
		return err
	}
	// set global client object
	client = c
	return nil
}

func createIndex(indexName, schemaFile string) error {

	ctx := context.Background()

	b, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		return err
	}

	createIndex, err := client.CreateIndex(indexName).BodyString(string(b)).Do(ctx)
	if err != nil {
		// Handle error
		return err
	}
	if !createIndex.Acknowledged {
		// Not acknowledged
		err = fmt.Errorf("failed to create '%s' index", indexName)
		return err
	}

	return nil
}

func deleteIndex(indexName string) error {

	ctx := context.Background()

	// Delete an index.
	deleteIndex, err := client.DeleteIndex(indexName).Do(ctx)
	if err != nil {
		// Handle error
		return err
	}
	if !deleteIndex.Acknowledged {
		// Not acknowledged
		err = fmt.Errorf("failed to delete '%s' index", indexName)
	}

	return nil
}
