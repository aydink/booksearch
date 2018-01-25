package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Document struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Serial     string   `json:"serial"`
	Department string   `json:"department"`
	Genre      string   `json:"genre"`
	Category   []string `json:"category"`
	Year       int      `json:"year"`
	Page       int      `json:"page"`
}

func getFileHash(filePath string) (string, error) {
	var md5string string

	file, err := os.Open(filePath)
	defer file.Close()

	if err != nil {
		return md5string, err
	}

	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return md5string, err
	}

	hashInBytes := hash.Sum(nil)
	//Convert the bytes to a string
	md5string = hex.EncodeToString(hashInBytes)

	return md5string, nil
}

func parseTextFile(filename string) ([]string, error) {

	content, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	pages := strings.Split(string(content), "\f")
	return pages, nil
}

func processFiles() {
	file, err := os.Open("file_list.txt")
	if err != nil {
		log.Fatalln(err)
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		md5hash := line[0:32]
		filename := line[33:]

		// remove ".pdf" extension from file name
		baseName := filename[0 : len(filename)-4]
		fmt.Println(md5hash, baseName)

		indexFile(md5hash, filename)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func indexFile(md5hash, filename string) {

	pages, err := parseTextFile("text/" + md5hash + ".txt")

	if err != nil {
		log.Fatalln(err)
		return
	}

	departments := []string{"KARA", "HAVA", "DENİZ", "GENEL", "MSB"}
	department := departments[rand.Intn(len(departments))]

	genres := []string{"Talimname", "Tez", "Teknik Doküman", "Klavuz", "Broşür"}
	genre := genres[rand.Intn(len(genres))]

	year := 1990 + rand.Intn(27)

	categories := []string{"Piyade", "Tank", "Top", "Ortak Konular", "Muhabere", "İkmal", "Bakım"}
	category := categories[0 : 1+rand.Intn(len(categories)-1)]

	for i := 0; i < len(pages)-1; i++ {

		doc := Document{}
		doc.Title = filename[0 : len(filename)-4]
		doc.Content = pages[i] + pages[i+1]
		doc.Page = i + 1
		doc.Department = department
		doc.Genre = genre
		doc.Year = year
		doc.Category = category

		b, err := json.Marshal(doc)
		if err != nil {
			log.Fatalln(err)
			return
		}
		fmt.Println(string(b))

		//fmt.Printf("%v", doc)
	}
}

func processBooks() {
	filepath.Walk("/Users/aydink/Downloads/E-Books/", func(path string, f os.FileInfo, err error) error {

		pathSeperator := "/"
		if runtime.GOOS == "windows" {
			pathSeperator = "\\"
		}

		if !f.IsDir() {

			if strings.ToLower(filepath.Ext(path)) == ".pdf" {
				//fmt.Println(path)
				md5hash, err := getFileHash(path)
				if err != nil {
					return err
				}

				//fmt.Println(md5hash)

				parent := filepath.Dir(path)
				parts := strings.Split(parent, pathSeperator)

				//fmt.Println(parent)
				author := parts[len(parts)-1]
				//fmt.Println("author:", author)

				bookName := f.Name()[0 : len(f.Name())-4]
				//fmt.Println("bookName", bookName)

				_, err = exec.Command("pdftotext", "-enc", "UTF-8", path, "text/"+md5hash+".txt").Output()
				if err != nil {
					//log.Fatalln(err)
					log.Println(err, path)
					return nil
				}

				//fmt.Println(out)

				indexNovel(md5hash, bookName, author)
			}
		}
		return nil
	})
}

func processTextFiles() {
	filepath.Walk("/Users/aydink/Downloads/E-Books/", func(path string, f os.FileInfo, err error) error {

		pathSeperator := "/"
		if runtime.GOOS == "windows" {
			pathSeperator = "\\"
		}

		if !f.IsDir() {

			if strings.ToLower(filepath.Ext(path)) == ".pdf" {
				//fmt.Println(path)
				md5hash, err := getFileHash(path)
				if err != nil {
					return err
				}

				//fmt.Println(md5hash)

				parent := filepath.Dir(path)
				parts := strings.Split(parent, pathSeperator)

				//fmt.Println(parent)
				author := parts[len(parts)-1]
				//fmt.Println("author:", author)

				bookName := f.Name()[0 : len(f.Name())-4]
				//fmt.Println("bookName", bookName)

				if _, err := os.Stat("text/" + md5hash + ".txt"); os.IsNotExist(err) {
					_, err = exec.Command("pdftotext", "-enc", "UTF-8", path, "text/"+md5hash+".txt").Output()
					if err != nil {
						//log.Fatalln(err)
						log.Println(err, path)
						return nil
					}
				}

				//fmt.Println(out)
				indexNovel(md5hash, bookName, author)
			}
		}
		return nil
	})
}

func indexNovel(md5hash, bookName, author string) {
	var buf bytes.Buffer

	pages, err := parseTextFile("text/" + md5hash + ".txt")

	if err != nil {
		log.Fatalln(err)
		return
	}

	genres := []string{"Bilim kurgu", "Polisiye", "Romantik", "Korku", "Tarih"}
	genre := genres[rand.Intn(len(genres))]

	departments := []string{"Yayın evi", "Kedi", "Epsilon", "Ötüken", "Test", "Tübitak", "MEB"}
	department := departments[rand.Intn(len(departments))]

	year := 1990 + rand.Intn(27)

	for i := 0; i < len(pages); i++ {

		doc := Document{}
		doc.Title = bookName
		doc.Content = pages[i]
		doc.Page = i + 1
		doc.Department = department
		doc.Genre = genre
		doc.Year = year
		doc.Category = []string{"novel"}

		b, err := json.Marshal(doc)
		if err != nil {
			log.Fatalln(err)
			return
		}
		//fmt.Println(string(b))

		buf.WriteString("{ \"index\" : { \"_index\" : \"book\", \"_type\" : \"novel\", \"_id\": \"" + md5hash + "-" + strconv.Itoa(doc.Page) + "\" } }")
		buf.WriteString("\n")
		buf.WriteString(string(b))
		buf.WriteString("\n")
		//fmt.Printf("%v", doc)
	}

	//fmt.Print(buf.String())
	postToElasticsearch(buf.Bytes())

}

func createFileList() {

	file, err := os.OpenFile("file_list.txt", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	defer file.Close()

	if err != nil {
		fmt.Println(err)
		return
	}

	filepath.Walk("/Users/aydink/Downloads/E-Books/", func(path string, f os.FileInfo, err error) error {

		if !f.IsDir() {

			if strings.ToLower(filepath.Ext(path)) == ".pdf" {
				//fmt.Println(path)
				md5hash, err := getFileHash(path)
				if err != nil {
					return err
				}

				file.WriteString(md5hash + "\t" + path + "\n")

			}
		}
		return nil
	})
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
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	rand.Seed(time.Now().Unix())

	//processFiles()

	//parseTextFile("text/1b1b0f1f65739f897d95849dfe11c600.txt")

	//processBooks()

	// process and index pdf files to Elasticsearh
	processTextFiles()

	// create file_list.txt which has tab seperated hash and path of pdf files
	//createFileList()
}
