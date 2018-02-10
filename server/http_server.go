package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("").Funcs(funcMap).ParseFiles("templates/test.html"))

	data := make(map[string]interface{})
	data["sayi"] = 2
	data["raw"] = "<h3>Level 3 Header</h3>"
	data["userid"] = r.Header.Get("userid")

	t.ExecuteTemplate(w, "test.html", data)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")

	if err != nil {
		fmt.Fprintf(w, "Hata: %s!", err)
	}
	t.Execute(w, nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {

	//t, err := template.ParseFiles("templates/search.html")
	//t := template.Must(template.New("").Funcs(funcMap).ParseFiles("templates/search.html", "templates/partial_facet.html", "templates/partial_pagination.html", "templates/partial_definition.html"))
	t := template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))

	keywords := r.URL.Query().Get("q")
	searchType := r.URL.Query().Get("w")
	start := r.URL.Query().Get("start")
	startInt, err := strconv.Atoi(start)
	//fmt.Println("start:", startInt)

	if err != nil {
		//fmt.Println("error parsing 'start' parameter")
		startInt = 0
	}

	data := make(map[string]interface{})
	templateName := "search"

	if searchType == "title" {
		data = titleQuery(keywords, startInt, getFilters(r.URL.Path))
		templateName = "title"
	} else {
		data = query(keywords, startInt, getFilters(r.URL.Path))
	}

	// show dictionary definion on only first page
	if startInt == 0 {
		data["definition"], data["hasDefinition"] = queryDictionary(keywords)
	}

	err = t.ExecuteTemplate(w, templateName, data)
	if err != nil {
		fmt.Println(err)
	}
}

func filterHandler(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	if strings.HasSuffix(url, "/") {
		fmt.Println(url)
		url = url[0 : len(url)-1]
		fmt.Println(url)
	}
	filters := getFilters(url)

	fmt.Fprintln(w, "filters:", filters, "<br>")
	fmt.Fprintln(w, "path:", r.URL.Path, "numparts:", len(filters))

}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))

	query := r.URL.Query().Get("page")
	q := r.URL.Query().Get("q")
	parts := strings.Split(query, "-")
	hash := parts[0]
	page := parts[1]

	createImage(query)

	data := make(map[string]interface{})
	data["q"] = q
	data["image"] = query
	data["hash"] = hash
	data["page"] = page
	data["doc"] = getDocument(query)

	t.ExecuteTemplate(w, "document", data)
}

func payloadHandler(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	q := r.URL.Query().Get("q")

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, QueryStringTokens(page, q))
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("page")
	parts := strings.Split(query, "-")
	hash := parts[0]
	page := parts[1]

	createImage(query)

	http.ServeFile(w, r, "static/images/"+hash+"-"+page+".png")
}

func reindexHandler(w http.ResponseWriter, r *http.Request) {
	go reindexAllFiles()
	fmt.Fprint(w, "Reindeing all pdf files")
}

func createImage(query string) {

	parts := strings.Split(query, "-")
	hash := parts[0]
	page := parts[1]

	//fmt.Println("hash:", hash, "page:", page, "file:", fileMap[hash])

	if _, err := os.Stat("static/images/" + hash + "-" + page + ".png"); os.IsNotExist(err) {
		_, err := exec.Command("pdftocairo", "-png", "-singlefile", "-f", page, "-l", page, "books/"+hash+".pdf", "static/images/"+hash+"-"+page).Output()
		if err != nil {
			//log.Fatalln(err)
			log.Println(err)
		}
	} else {
		//fmt.Println("-----------------------", "using cashed image")
	}
}

// PDF file handler
// send pdf file and sets a proper title
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("book")
	if len(hash) > 32 {
		hash = hash[:32]
	}

	if len(hash) < 32 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Geçersiz bir istekte bulundunuz.")
		log.Printf("download pdf, invalid hash value:%s", hash)
		return
	}

	// check if user wants to download file
	force := r.URL.Query().Get("force")

	file, err := os.Open("books/" + hash + ".pdf")
	defer file.Close()
	if err != nil {
		log.Printf("failed to serve pdf file:%s", "books/"+hash+".pdf")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Geçersiz bir istekte bulundunuz.")
		return
	}

	book := getBook(hash)
	name := book.Serial + " " + book.Title
	name = strings.TrimSpace(name)

	// if there is an explicit url prameter "force=true" then force browser to download not try to display the pdf file
	if force == "true" {
		w.Header().Set("Content-Disposition", "attachment; filename="+name+".pdf")
	}

	w.Header().Set("Content-Type", "application/pdf")
	io.Copy(w, file)
}

func main() {

	log.SetFlags(log.Llongfile)

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/test", requireAuthMiddleware(testHandler))
	http.HandleFunc("/api/reindex", reindexHandler)

	http.HandleFunc("/search/", searchHandler)
	http.HandleFunc("/filter/", filterHandler)
	http.HandleFunc("/page", pageHandler)
	http.HandleFunc("/image", imageHandler)
	http.HandleFunc("/download/", downloadHandler)
	http.HandleFunc("/api/addbook", ApiIndexFile)
	http.HandleFunc("/api/payloads", payloadHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.ListenAndServe(":8080", nil)
}
