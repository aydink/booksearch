package main

import (
	"fmt"
	"html/template"
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
	start := r.URL.Query().Get("start")
	startInt, err := strconv.Atoi(start)
	fmt.Println("start:", startInt)

	if err != nil {
		fmt.Println("error parsing 'start' parameter")
		startInt = 0
	}

	data := query(keywords, startInt, getFilters(r.URL.Path))

	// show dictionary definion on only first page
	if startInt == 0 {
		data["definition"], data["hasDefinition"] = queryDictionary(keywords)
	}

	err = t.ExecuteTemplate(w, "search", data)
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
	t, err := template.ParseFiles("templates/document.html")
	if err != nil {
		fmt.Fprintf(w, "Hata: %s!", err)
	}

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

	t.Execute(w, data)
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

func createImage(query string) {

	parts := strings.Split(query, "-")
	hash := parts[0]
	page := parts[1]

	fmt.Println("hash:", hash, "page:", page, "file:", fileMap[hash])

	if _, err := os.Stat("static/images/" + hash + "-" + page + ".png"); os.IsNotExist(err) {
		_, err := exec.Command("pdftocairo", "-png", "-singlefile", "-f", page, "-l", page, "books/"+hash+".pdf", "static/images/"+hash+"-"+page).Output()
		if err != nil {
			//log.Fatalln(err)
			log.Println(err, fileMap[hash])
		}
	} else {
		fmt.Println("-----------------------", "using cashed image")
	}
}

func createImage_old(query string) {

	parts := strings.Split(query, "-")
	hash := parts[0]
	page := parts[1]

	fmt.Println("hash:", hash, "page:", page, "file:", fileMap[hash])

	if _, err := os.Stat("static/images/" + hash + "-" + page + ".png"); os.IsNotExist(err) {
		_, err := exec.Command("pdftocairo", "-png", "-singlefile", "-f", page, "-l", page, fileMap[hash], "static/images/"+hash+"-"+page).Output()
		if err != nil {
			//log.Fatalln(err)
			log.Println(err, fileMap[hash])
		}
	} else {
		fmt.Println("-----------------------", "using cashed image")
	}
}

func main() {

	log.SetFlags(log.Llongfile)

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/test", requireAuthMiddleware(testHandler))
	http.HandleFunc("/search/", searchHandler)
	http.HandleFunc("/filter/", filterHandler)
	http.HandleFunc("/page", pageHandler)
	http.HandleFunc("/image", imageHandler)
	http.HandleFunc("/api/addbook", ApiIndexFile)
	http.HandleFunc("/api/payloads", payloadHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.ListenAndServe(":8080", nil)
}
