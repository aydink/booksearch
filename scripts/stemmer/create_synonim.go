package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"
)

type Stem struct {
	stem  string
	count int
}

func showStemFrequency() {

	dict := make(map[string]int)

	file, err := os.Open("stem.txt")
	if err != nil {
		log.Fatalln(err)
		return
	}

	r := strings.NewReplacer("â", "a", "î", "i", "û", "u")

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Split(r.Replace(strings.ToLowerSpecial(unicode.TurkishCase, scanner.Text())), "\t")

		//word := line[0]
		stem := line[1]

		if _, ok := dict[stem]; ok {
			dict[stem]++
		} else {
			dict[stem] = 1
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// To store the keys in slice in sorted order
	var stems []Stem
	for k := range dict {
		stems = append(stems, Stem{k, dict[k]})
	}

	sort.Slice(stems, func(i, j int) bool {
		return stems[i].count < stems[j].count
	})

	counter := 0
	// To perform the opertion you want
	for k, v := range stems {
		fmt.Println("Key:", k, "Value:", v)
		counter++
	}

	fmt.Println("toplam:", counter)
}

func createSynonimFile() {

	file, err := os.Open("stem.txt")
	if err != nil {
		log.Fatalln(err)
		return
	}

	r := strings.NewReplacer("â", "a", "î", "i", "û", "u")

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Split(r.Replace(strings.ToLowerSpecial(unicode.TurkishCase, scanner.Text())), "\t")
		word := line[0]
		stem := line[1]

		if strings.Compare(word, stem) != 0 {
			fmt.Printf("%s=>%s\n", word, stem)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

// before the senver accept connection load file mapping from disk
func main() {
	//showStemFrequency()
	createSynonimFile()
}
