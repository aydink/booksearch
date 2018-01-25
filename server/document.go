package main

type Document struct {
	Id               string
	Title            string   `json:"title"`
	Content          string   `json:"content"`
	Serial           string   `json:"serial"`
	Department       string   `json:"department"`
	Genre            string   `json:"genre"`
	Category         []string `json:"category"`
	Year             int      `json:"year"`
	Page             int      `json:"page"`
	TitleHighlight   string
	ContentHighlight string
}
