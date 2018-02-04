package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("templates/upload.html")
		t.Execute(w, token)
	} else {
		// max file size is 200 mb --> 209715200 bytes
		r.ParseMultipartForm(209715200)

		formsErrors := make(map[string]string)

		serial := strings.TrimSpace(r.PostFormValue("serial"))
		title := strings.TrimSpace(r.PostFormValue("title"))
		if len(title) < 2 {
			formsErrors["title"] = "Kitap adı 2 karakterden kısa olamaz!"
		}
		department := strings.TrimSpace(r.PostFormValue("department"))
		if len(department) < 1 {
			formsErrors["department"] = "Yönergenin sahibi komutanlık seçmelisiniz!"
		}
		genre := strings.TrimSpace(r.PostFormValue("genre"))
		if len(genre) < 1 {
			formsErrors["genre"] = "Yayın türünü seçmelisiniz!"
		}
		category := r.PostForm["category"]
		yearString := strings.TrimSpace(r.PostFormValue("year"))

		year, err := strconv.Atoi(yearString)
		if err != nil {
			formsErrors["year"] = "Basım yılı geçerli değil!"
		}

		book := Book{}
		book.Serial = serial
		book.Title = title
		book.Department = department
		book.Genre = genre
		book.Category = category
		book.Year = year

		if len(formsErrors) > 0 {
			log.Printf("API addbook errors:%s\n", formsErrors)
			fmt.Printf("%+v", book)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "form errors:%s", formsErrors)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			fmt.Println("FormFile:", err)
			return
		}
		defer file.Close()

		fmt.Fprintf(w, "Headers --> %v\n", handler.Header["Content-Type"][0])

		if handler.Header["Content-Type"][0] == "application/pdf" {

			tempFileName := pseudo_uuid()

			f, err := os.OpenFile("books/"+tempFileName, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				fmt.Println("OpenFile:", err)
				return
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
			}

			book.Hash = md5string
			fmt.Printf("%+v\n", book)
			processPdfFile(book)

		} else {
			fmt.Println("Content-Type not supported. Expecting application/pdf but found", handler.Header["Content-Type"][0])
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Content-Type not supported. Expecting application/pdf but found", handler.Header["Content-Type"][0])
		}
	}
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

	fmt.Printf("%+v\n", book)
	return nil

}
