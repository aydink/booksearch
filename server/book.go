package main

type Book struct {
	Id         string   `json:"id"`
	Title      string   `json:"title"`
	Author     string   `json:"author"`
	Serial     string   `json:"serial"`
	Department string   `json:"department"`
	Genre      string   `json:"genre"`
	Category   []string `json:"category"`
	Year       int      `json:"year"`
	NumPages   int      `json:"num_pages"`
	Hash       string   `json:"hash"`
}
