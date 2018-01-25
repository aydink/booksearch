package main

import (
	"html/template"
)

var funcMap template.FuncMap

func Increment(i int) int {
	return i + 1
}

func Add(a, b int) int {
	return a + b
}

func ToHtml(s string) template.HTML {
	return template.HTML(s)
}

func init() {
	funcMap = template.FuncMap{
		"inc":    Increment,
		"add":    Add,
		"tohtml": ToHtml,
	}
}
