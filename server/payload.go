package main

import (
	"bufio"
	"bytes"
	"log"
	"unicode"

	"encoding/gob"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger"
	"golang.org/x/net/html"
)

type BBox struct {
	XMin int16
	YMin int16
	XMax int16
	YMax int16
}

//var db *kv.DB
var db *badger.DB

var dict map[string]string
var replacer *strings.Replacer

var f = func(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsNumber(c)
}

func int16ToBytes(x int16) []byte {
	var buf [2]byte
	buf[0] = byte(x >> 0)
	buf[1] = byte(x >> 8)
	return buf[:]
}

func bytesToInt16le(b []byte) int16 {
	return int16(uint16(b[0]) | uint16(b[1])<<8)
}

func Stem(token string) string {
	// lowercase string and replace "â", "a", "î", "i", "û", "u" accented characters
	//token = strings.ToLowerSpecial(unicode.TurkishCase, token)
	//token = replacer.Replace(token)
	// token = strings.ToLower(token)

	//parts := strings.FieldsFunc(token, f)

	/*
		token = strings.Trim(token, ".")
		token = strings.Trim(token, ",")
		token = strings.Trim(token, ":")
		token = strings.Trim(token, ";")
		token = strings.Trim(token, "\"")
		token = strings.Trim(token, "/")
		token = strings.Trim(token, "'")
		token = strings.Trim(token, "!")
		token = strings.Trim(token, "?")
		token = strings.Trim(token, "-")
		token = strings.Trim(token, ")")
		token = strings.Trim(token, "(")
		token = strings.Trim(token, "{")
		token = strings.Trim(token, "{")
		token = strings.Trim(token, "[")
		token = strings.Trim(token, "]")
		//fmt.Println(token)
		//return porterstemmer.StemString(token)
	*/

	//fmt.Println("****")
	//fmt.Println("input:", token)

	if val, ok := dict[token]; ok {
		token = val
	}
	//fmt.Println("output:", token)

	return token
}

func EncodePayload(m map[string][]BBox) []byte {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	encoder := gob.NewEncoder(w)
	err := encoder.Encode(m)
	if err != nil {
		fmt.Println(err)
	}

	return b.Bytes()
}

func DecodePayload(data []byte) map[string][]BBox {
	r := bytes.NewReader(data)

	payload := make(map[string][]BBox)

	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&payload)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(payload)

	return payload
}

func GetTokenPositions(page string, tokens []string) string {
	fmt.Println("page::::::::::", page, tokens)

	jsonStr := GetPage(page)

	allTokens := make(map[string][][4]int)
	filteredTokens := make(map[string][][4]int)

	err := json.Unmarshal(jsonStr, &allTokens)
	if err != nil {
		log.Println(err)
	}

	for _, token := range tokens {
		filteredTokens[token] = allTokens[token]
	}

	jsonString, err := json.Marshal(filteredTokens)
	if err != nil {
		fmt.Println(err)
	}

	return string(jsonString)
}

func SavePage(key, value []byte) {

	fmt.Println("INSIDE SAVE PAGE:", string(key))
	fmt.Println(DecodePayload(value))

	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		return err
	})

	if err != nil {
		fmt.Println(err)
	}

}

func GetPage(key string) []byte {
	fmt.Println("Load page:", key)

	var s []byte

	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		val, err := item.Value()
		fmt.Println("length of slice:", len(val))
		s = val
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Println(err)
		log.Println("key:", key)
	}
	fmt.Println("length of slice:", len(s))
	//fmt.Println(string(v))
	return s
}

func QueryStringTokens(page string, q string) string {
	// lowercase string and replace "â", "a", "î", "i", "û", "u" accented characters
	s := strings.ToLowerSpecial(unicode.TurkishCase, q)
	s = replacer.Replace(s)

	fmt.Println(q)

	tokens := strings.FieldsFunc(s, f)

	fmt.Println("num tokens:", len(tokens))

	fmt.Println("***********")

	for key, val := range tokens {
		fmt.Println("key:", key, "val:", val)
		tokens[key] = Stem(val)
	}

	fmt.Println("***********")

	return GetTokenPositions(page, tokens)
}

func ProcessPayloadFile(hash string) {

	var pageNumber int
	file, err := os.Open("books/" + hash + ".bbox.txt")
	if err != nil {
		fmt.Println(err)
	}

	z := html.NewTokenizer(file)

	var bbox [4]int
	var tokens map[string][][4]int

	insideWord := false

	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			return

		case html.StartTagToken:
			t := z.Token()

			if t.Data == "page" {
				pageNumber++
				fmt.Println(pageNumber, "------------------------------")

				tokens = make(map[string][][4]int)
			}

			if t.Data == "word" {

				//bbox = BBox{}
				bbox = [4]int{}

				for _, w := range t.Attr {
					n, err := strconv.ParseFloat(w.Val, 64)
					if err != nil {
						fmt.Println(err)
					}
					n = math.Floor(n + 0.5)
					coor := int(n)

					switch w.Key {
					case "xmin":
						//bbox.XMin = coor
						bbox[0] = coor
					case "ymin":
						//bbox.YMin = coor
						bbox[1] = coor
					case "xmax":
						//bbox.XMax = coor
						bbox[2] = coor
					case "ymax":
						//bbox.YMax = coor
						bbox[3] = coor
					}
				}

				insideWord = true
			} else {
				insideWord = false
			}

		case html.TextToken:
			if insideWord {
				token := strings.TrimSpace(z.Token().Data)

				// lowercase string and replace "â", "a", "î", "i", "û", "u" accented characters
				token = strings.ToLowerSpecial(unicode.TurkishCase, token)
				token = replacer.Replace(token)

				parts := strings.FieldsFunc(token, f)

				for i := 0; i < len(parts); i++ {
					token := Stem(parts[i])
					if len(token) > 0 {
						tokens[token] = append(tokens[token], bbox)
					}
				}
			}

		case html.EndTagToken: // </tag>
			t := z.Token()
			if t.Data == "page" {
				//fmt.Println("end page:", pageNumber)
				//fmt.Println(len(tokens))

				// insert Payloads into KV store. Use md5 hash and page number as key
				key := hash + "-" + strconv.Itoa(pageNumber)
				//fmt.Println(key)
				//fmt.Println(tokens)

				jsonStr, err := json.Marshal(tokens)
				if err != nil {
					log.Fatalln(jsonStr)
				}
				//SavePage([]byte(key), EncodePayload(tokens))
				SavePage([]byte(key), jsonStr)
			}
		}
	}
}

func loadTurkishStems() map[string]string {
	file, err := os.Open("data/turkish_synonym.txt")
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	dict := make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), "=>")
		dict[line[0]] = line[1]
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return dict
}

func init() {

	dict = loadTurkishStems()
	fmt.Println("stemmer dictionart loaded:", len(dict), "items")

	replacer = strings.NewReplacer("â", "a", "î", "i", "û", "u")

	//db = OpenDatabase()

	var err error

	opts := badger.DefaultOptions
	opts.Dir = "data/badger"
	opts.ValueDir = "data/badger"
	db, err = badger.Open(opts)

	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close()

}

func main2() {

	/*
			book := flag.String("book", "", "name of the book")
			page := flag.String("page", "1", "page numbers")

			htmlDir := flag.String("html", "", "directory containing html files with bbox values")

			flag.Parse()

			//db = OpenDatabase()

			//db.BeginTransaction()



				if *htmlDir != "" {
					ProcessFiles(*htmlDir)
				} else {
					fmt.Println(*book + "-" + *page)
					GetPage(*book + "-" + *page)
				}

				k, v, e := db.Last()

				fmt.Println(string(k), string(v), e)

				//db.Commit()

				//GetPage("test-10")
				db.Close()

				//ExampleRemove(s)
				//SavePayload()


		db.BeginTransaction()

		ProcessBBOXFile("test.html")

		db.Commit()

		//GetPage("test-40")
		//fmt.Println(GetTokenPositions("test-40", []string{"yer", "iç"}))

		s := "kimse insan haysîyetiyle bağdaşmayan bir cezaya veya muameleye TÂBÎ tutulamaz"
		sonuc := QueryStringTokens("test-3", s)
		fmt.Println(sonuc)

		db.Close()
	*/

	ProcessPayloadFile("34a5e226ff1d8bcd402ee23c716ac40b")

	page := "34a5e226ff1d8bcd402ee23c716ac40b-24"
	q := "bakanlar"
	fmt.Println(QueryStringTokens(page, q))
}
