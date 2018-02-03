package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func parseTextFile(hash string) ([]string, error) {

	content, err := ioutil.ReadFile("books/" + hash + ".txt")

	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	pages := strings.Split(string(content), "\f")
	return pages, nil
}

func indexBook(book Book) {

	var buf bytes.Buffer
	bookJson, err := json.Marshal(book)
	if err != nil {
		log.Fatalln(err)
		return
	}

	buf.WriteString("{ \"index\" : { \"_index\" : \"ray\", \"_type\" : \"book\", \"_id\": \"" + book.Hash + "\" } }")
	buf.WriteString("\n")
	buf.WriteString(string(bookJson))
	buf.WriteString("\n")

	postToElasticsearch(buf.Bytes())

	// clear buffer
	buf.Reset()

	pages, err := parseTextFile(book.Hash)

	if err != nil {
		log.Fatalln(err)
		return
	}

	numPages := len(pages)

	for i := 0; i < numPages; i++ {

		doc := Document{}
		doc.Title = book.Title
		doc.Content = pages[i]
		doc.Page = i + 1
		doc.Department = book.Department
		doc.Genre = book.Genre
		doc.Year = book.Year
		doc.Category = book.Category
		doc.NumPages = numPages

		b, err := json.Marshal(doc)
		if err != nil {
			log.Fatalln(err)
			return
		}
		fmt.Println(string(b))

		buf.WriteString("{ \"index\" : { \"_index\" : \"book\", \"_type\" : \"novel\", \"_id\": \"" + book.Hash + "-" + strconv.Itoa(doc.Page) + "\" } }")
		buf.WriteString("\n")
		buf.WriteString(string(b))
		buf.WriteString("\n")
		//fmt.Printf("%v", doc)
	}

	//fmt.Print(buf.String())
	postToElasticsearch(buf.Bytes())

}

func postToElasticsearch(buffer []byte) {

	url := "http://localhost:9200/_bulk"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(buffer))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	//body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("response Body:", string(body))
}
