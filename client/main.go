package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

type Book struct {
	Id         string   `json:"id"`
	Title      string   `json:"title"`
	Author     string   `json:"author"`
	Serial     string   `json:"serial"`
	Department string   `json:"department"`
	Genre      string   `json:"genre"`
	Category   []string `json:"category"`
	Year       string   `json:"year"`
	NumPages   string   `json:"num_pages"`
	Hash       string   `json:"hash"`
}

func Upload(url, file string, book Book) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// Add your image file
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()
	fw, err := w.CreateFormFile("file", file)
	if err != nil {
		return
	}
	if _, err = io.Copy(fw, f); err != nil {
		return
	}
	// Add serial fields
	if fw, err = w.CreateFormField("serial"); err != nil {
		return
	}
	if _, err = fw.Write([]byte(book.Serial)); err != nil {
		return
	}
	// Add title field
	if fw, err = w.CreateFormField("title"); err != nil {
		return
	}
	if _, err = fw.Write([]byte(book.Title)); err != nil {
		return
	}
	// Add genre field
	if fw, err = w.CreateFormField("genre"); err != nil {
		return
	}
	if _, err = fw.Write([]byte(book.Genre)); err != nil {
		return
	}

	// Add genre field
	if fw, err = w.CreateFormField("department"); err != nil {
		return
	}
	if _, err = fw.Write([]byte("Mevzuat")); err != nil {
		return
	}

	// Add year field
	if fw, err = w.CreateFormField("year"); err != nil {
		return
	}
	if _, err = fw.Write([]byte(book.Year)); err != nil {
		return
	}

	for _, cat := range book.Category {
		// Add genre field
		if fw, err = w.CreateFormField("category"); err != nil {
			return
		}
		if _, err = fw.Write([]byte(cat)); err != nil {
			return
		}
	}

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())
	//req.Header.Set("Content-Type", "application/pdf")

	// Submit the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

func uploadBooks(csvFile string) {
	file, err := os.Open(csvFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	r := csv.NewReader(file)
	r.Comma = ';'
	r.Comment = '#'

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		book := Book{}
		book.Serial = record[1]
		book.Title = record[2]
		book.Year = record[3]
		book.Genre = record[4]

		fmt.Println(record[2])

		Upload("http://localhost:8080/api/addbook", "/Users/aydink/Downloads/kanunlar/"+record[0], book)
	}
}

func main() {
	csvFilename := os.Args[1]
	uploadBooks(csvFilename)
	//Upload("http://localhost:8080/api/addbook", "/Users/aydink/Desktop/anayasa_2011.pdf")
	//Upload("http://localhost:8080/api/addbook", "/Users/aydink/Desktop/PostGIS in Action.pdf")
}
