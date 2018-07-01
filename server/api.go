package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func pseudo_uuid() (uuid string) {

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	//fmt.Println(uuid)

	return
}

func ApiIndexFile(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))

	if r.Method == "GET" {

		data := make(map[string]interface{})
		t.ExecuteTemplate(w, "upload", data)

	} else {

		errorMap, err := processUploadedPdf(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			for key, val := range errorMap {
				fmt.Fprintf(w, "%s: %s\n", key, val)
			}
			return
		}
		data := make(map[string]interface{})
		data["message"] = "Dosya başarıyla yüklendi."
		t.ExecuteTemplate(w, "upload", data)
	}
}

func processUploadedPdf(r *http.Request) (map[string]string, error) {

	// max file size is 200 mb --> 209715200 bytes
	r.ParseMultipartForm(209715200)

	formErrors := make(map[string]string)

	serial := strings.TrimSpace(r.PostFormValue("serial"))
	title := strings.TrimSpace(r.PostFormValue("title"))
	if len(title) < 2 {
		formErrors["title"] = "Kitap adı 2 karakterden kısa olamaz!"
	}
	department := strings.TrimSpace(r.PostFormValue("department"))
	if len(department) < 1 {
		formErrors["department"] = "Yönergenin sahibi komutanlık seçmelisiniz!"
	}
	genre := strings.TrimSpace(r.PostFormValue("genre"))
	if len(genre) < 1 {
		formErrors["genre"] = "Yayın türünü seçmelisiniz!"
	}
	category := r.PostForm["category"]
	yearString := strings.TrimSpace(r.PostFormValue("year"))

	year, err := strconv.Atoi(yearString)
	if err != nil {
		formErrors["year"] = "Basım yılı geçerli değil!"
	}

	book := Book{}
	book.Serial = serial
	book.Title = title
	book.Department = department
	book.Genre = genre
	book.Category = category
	book.Year = year

	if len(formErrors) > 0 {
		log.Printf("/api/addbook errors:%s\n", formErrors)
		return formErrors, errors.New("uploaded form has errors")
		//fmt.Printf("%+v", book)
	}

	file, _, err := r.FormFile("file")
	defer file.Close()
	if err != nil {
		fmt.Println("FormFile:", err)
		return formErrors, err
	}

	// Create a buffer to store the header of the file in
	fileHeader := make([]byte, 512)

	// Copy the headers into the FileHeader buffer
	if _, err := file.Read(fileHeader); err != nil {
		return formErrors, err
	}

	// set position back to start.
	if _, err := file.Seek(0, 0); err != nil {
		return formErrors, err
	}

	contentType := http.DetectContentType(fileHeader)

	if contentType == "application/pdf" {

		tempFileName := pseudo_uuid()

		f, err := os.Create("books/" + tempFileName)
		if err != nil {
			fmt.Println("OpenFile:", err)
			return formErrors, err
		}

		h := md5.New()
		multiWriter := io.MultiWriter(f, h)

		io.Copy(multiWriter, file)

		hashInBytes := h.Sum(nil)
		//Convert the bytes to a string
		md5string := hex.EncodeToString(hashInBytes)

		// close the temp file and rename it using md5 hash of the file
		f.Close()
		err = os.Rename("books/"+tempFileName, "books/"+md5string+".pdf")
		if err != nil {
			fmt.Println("File rename failed:", err)
			return formErrors, err
		}

		book.Hash = md5string
		//fmt.Printf("%+v\n", book)
		processPdfFile(book)
	} else {
		log.Printf("Content-Type not supported, expecting application/pdf found %s\n", contentType)
		return formErrors, fmt.Errorf("Content-Type not supported, expecting application/pdf found %s\n", contentType)
	}

	return formErrors, nil
}

func processPdfFile(book Book) error {

	//_, err := exec.Command("pdftocairo", "-png", "-singlefile", "-f", page, "-l", page, fileMap[hash], "static/images/"+hash+"-"+page).Output()
	output, err := exec.Command("pdfinfo", "books/"+book.Hash+".pdf").Output()
	if err != nil {
		fmt.Println(err)
		return err
	}

	re := regexp.MustCompile("Pages: *([0-9]+)")
	matches := re.FindStringSubmatch(string(output))
	if len(matches) == 2 {
		book.NumPages, err = strconv.Atoi(matches[1])
		if err != nil {
			log.Printf("Failed to find PDF file number of pages, file:%s.pdf, error:%s\n", err, book.Hash)
			return err
		}
	}

	if _, err := os.Stat("books/" + book.Hash + ".txt"); os.IsNotExist(err) {
		_, err = exec.Command("pdftotext", "-enc", "UTF-8", "books/"+book.Hash+".pdf", "books/"+book.Hash+".txt").Output()
		if err != nil {
			//log.Fatalln(err)
			log.Printf("PDF text extraction failed, file:%s.pdf, error:%s\n", err, book.Hash)
			return err
		}
	}

	if _, err := os.Stat("books/" + book.Hash + ".bbox.txt"); os.IsNotExist(err) {
		_, err = exec.Command("pdftotext", "-enc", "UTF-8", "-bbox", "books/"+book.Hash+".pdf", "books/"+book.Hash+".bbox.txt").Output()
		if err != nil {
			//log.Fatalln(err)
			log.Printf("PDF payload extraction failed, file:%s.pdf, error:%s\n", err, book.Hash)
			return err
		}
	}

	// insert BBOX payload data into KV store
	ProcessPayloadFile(book.Hash)

	// send book to elasticsearh
	indexBook(book)

	// create filehash.json file for pdf file
	err = saveBookMeta(book)
	if err != nil {
		return err
	}

	//fmt.Printf("%+v\n", book)
	return nil

}

func saveBookMeta(book Book) error {

	bookJson, err := json.Marshal(book)
	if err != nil {
		return err
	}

	file, err := os.Create("books/" + book.Hash + ".json")
	defer file.Close()
	if err != nil {
		return err
	}

	_, err = file.Write(bookJson)
	if err != nil {
		return err
	}

	return nil
}

func loadBookMeta(filename string) (Book, error) {

	book := Book{}

	file, err := os.Open("books/" + filename)
	defer file.Close()
	if err != nil {
		return book, err
	}

	bookJson, err := ioutil.ReadAll(file)
	if err != nil {
		return book, err
	}

	err = json.Unmarshal(bookJson, &book)
	if err != nil {
		return book, err
	}

	return book, err
}

func reindexAllFiles() {
	fileInfos, err := ioutil.ReadDir("books")
	if err != nil {
		log.Printf("opening books directory failed.")
		return
	}

	for _, file := range fileInfos {
		if filepath.Ext(file.Name()) == ".json" {
			book, err := loadBookMeta(file.Name())
			if err != nil {
				log.Printf("loading file meta from json file:%s faied\n", err)
				continue
			}
			fmt.Println(book)
			indexBook(book)

			//store payload data in elasticsearch
			ProcessPayloadFile(book.Hash)
		}
	}
}
