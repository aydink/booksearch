package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"unicode"
)

type Definition struct {
	definition string
	pos        string
}

func createFile() {

	dict := make(map[string][]Definition)

	file, err := os.Open("sozluk.txt")
	if err != nil {
		log.Fatalln(err)
		return
	}

	r := strings.NewReplacer("â", "a", "î", "i", "û", "u")

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		line := strings.Split(scanner.Text(), "\t")
		eng := strings.ToLower(line[0])
		tur := r.Replace(strings.ToLowerSpecial(unicode.TurkishCase, line[1]))
		pos := line[2]

		if _, ok := dict[eng]; !ok {
			defs := make([]Definition, 0)
			defs = append(defs, Definition{tur, pos})
			dict[eng] = defs
		} else {
			dict[eng] = append(dict[eng], Definition{tur, pos})
		}

	}

	/*
		for k, v := range dict {
			fmt.Println("Key:", k)

			prevPos := ""

			for i := 0; i < len(v); i++ {
				if prevPos != v[i].pos {
					fmt.Print(v[i].pos, " ")
					prevPos = v[i].pos
				}

				fmt.Print(v[i].definition, ", ")
			}
			fmt.Println("")
		}
	*/

	var buf bytes.Buffer

	for k, v := range dict {

		fmt.Println("{\"index\": {} }")
		fmt.Printf("{\"eng\": \"%s\", \"tur\":\"", k)

		fmt.Fprintln(&buf, "{\"index\": {} }")
		fmt.Fprintf(&buf, "{\"eng\": \"%s\", \"tur\":\"", k)

		prevPos := ""

		for i := 0; i < len(v); i++ {
			if prevPos != v[i].pos {
				fmt.Print(v[i].pos, " ")
				fmt.Fprint(&buf, v[i].pos, " ")
				prevPos = v[i].pos
			}

			// if we are printing last definition, do not print ", "
			if i < len(v)-1 {
				fmt.Print(v[i].definition, ", ")
				fmt.Fprint(&buf, v[i].definition, ", ")
			} else {
				fmt.Print(v[i].definition)
				fmt.Fprint(&buf, v[i].definition)
			}

		}
		fmt.Println("\"}")
		fmt.Fprintln(&buf, "\"}")
	}

	postDictionary(buf.Bytes())

}

func postDictionary(buffer []byte) {

	buffer = bytes.Replace(buffer, []byte("{N}"), []byte("<br><em>isim</em> "), -1)
	buffer = bytes.Replace(buffer, []byte("{V}"), []byte("<br><em>fiil</em> "), -1)
	buffer = bytes.Replace(buffer, []byte("{A}"), []byte("<br><em>sıfat</em> "), -1)
	buffer = bytes.Replace(buffer, []byte("{ADV}"), []byte("<br><em>zarf</em> "), -1)
	buffer = bytes.Replace(buffer, []byte("{ID}"), []byte("<br><em>deyim</em> "), -1)

	//fmt.Println(string(buffer))

	url := "http://localhost:9200/dictionary/entry/_bulk"

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

// before the senver accept connection load file mapping from disk
func main() {
	//showStemFrequency()
	createFile()
}
