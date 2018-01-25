package main

import (
	"fmt"
	"strings"
)

var filterFullNames = map[string]string{
	"year":     "Basım yılı",
	"genre":    "Türü",
	"category": "Kategori",
}

// getFullFilterName return full name of the filter
// eg. "year": "Yıl", "genre": "Türü"
func getFullFilterName(key string) string {
	if value, found := filterFullNames[key]; found {
		return value
	}
	return key
}

// getFilterMap converts url path to key value filter map
func getFilters(url string) [][3]string {
	fmt.Println("url path:", url)

	// for consistency strip trailing "/" path seperator at the end of the path
	if strings.HasSuffix(url, "/") {
		url = url[0 : len(url)-1]
	}

	filters := make([][3]string, 0)

	parts := strings.Split(url, "/")

	if len(parts)%2 == 0 {

		for i := 2; i < len(parts)-1; i = i + 2 {
			if (strings.TrimSpace(parts[i]) != "") && (strings.TrimSpace(parts[i+1]) != "") {
				filters = append(filters, [3]string{parts[i], getFullFilterName(parts[i]), parts[i+1]})
			} else {
				fmt.Println("hata")
			}
		}
	}
	return filters
}
