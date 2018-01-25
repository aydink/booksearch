package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"strconv"
	"strings"

	"github.com/olivere/elastic"
)

type Facet struct {
	Key      string
	DocCount int64
}

type DictionaryEntry struct {
	English string        `json:"eng"`
	Turkish template.HTML `json:"tur"`
}

func printQuery(query elastic.Query) {
	src, err := query.Source()
	if err != nil {
		panic(err)
	}
	data, err := json.Marshal(src)
	if err != nil {
		panic(err)
	}
	s := string(data)
	fmt.Println(s)
}

//func query(keywords string, filterMap map[string]string) map[string]interface{} {
func query(keywords string, start int, filters [][3]string) map[string]interface{} {

	fmt.Println("keywords:", keywords)
	fmt.Println("filters:", filters)

	ctx := context.Background()

	url := "http://127.0.0.1:9200"

	//Create an Elasticsearch client
	client, err := elastic.NewClient(elastic.SetURL(url), elastic.SetSniff(true))
	if err != nil {
		log.Fatal(err)
	}

	boolQuery := elastic.NewBoolQuery()
	boolQuery = boolQuery.Must(elastic.NewMatchQuery("content", keywords))
	boolQuery = boolQuery.Should(elastic.NewMatchQuery("content.bigrammed", keywords))
	//boolQuery = boolQuery.MinimumShouldMatch("1")

	for i := 0; i < len(filters); i++ {
		boolQuery = boolQuery.Filter(elastic.NewTermQuery(filters[i][0], filters[i][2]))
		fmt.Println("filtre eklendi")
	}

	printQuery(boolQuery)

	higlight := elastic.NewHighlight()
	higlight = higlight.Field("content")
	higlight = higlight.FragmentSize(200)
	higlight = higlight.NumOfFragments(2)

	//printQuery(higlight)

	aggsYear := elastic.NewTermsAggregation()
	aggsYear = aggsYear.Field("year")

	aggsGenre := elastic.NewTermsAggregation()
	aggsGenre = aggsGenre.Field("genre")

	generator := elastic.NewDirectCandidateGenerator("content.trigram")
	generator = generator.SuggestMode("always")
	generator = generator.MinWordLength(3)

	suggester := elastic.NewPhraseSuggester("phrase_suggestion")
	suggester = suggester.Field("content.trigram")
	suggester = suggester.Size(1)
	suggester = suggester.GramSize(3)
	suggester = suggester.CandidateGenerator(generator)
	suggester = suggester.Text(keywords)
	suggester = suggester.Highlight("<em>", "</em>")

	//printQuery(aggsGenre)
	//printQuery(aggsYear)

	/*
		postFilter := elastic.NewBoolQuery()
		postFilter = postFilter.Must(elastic.NewTermQuery("year", 2000))
		postFilter = postFilter.Must(elastic.NewTermQuery("genre", "Korku"))

		printQuery(postFilter)
	*/

	search := client.Search().
		Index("book").                       // search in index "twitter"
		Query(boolQuery).                    // specify the query
		From(start).Size(10).                // take documents 0-9
		Pretty(true).                        // pretty print request and response JSON
		Highlight(higlight).                 // Highlight results
		Aggregation("Basım yılı", aggsYear). // Aggregation basım yılı
		Aggregation("Türü", aggsGenre).      // Aggregation basım yılı
		Suggester(suggester)
		//PostFilter(postFilter)               // Apply Post_filter

	searchResult, err := search.Do(ctx) // execute
	if err != nil {
		// Handle error
		//panic(err)
		fmt.Println(err)
	}

	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from Elasticsearch.
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

	fmt.Println(searchResult.TotalHits())

	data := make(map[string]interface{}, 0)
	data["TotalHits"] = searchResult.TotalHits()

	docs := make([]Document, 0, 10)
	yearFacet := make([]Facet, 0)
	genreFacet := make([]Facet, 0)

	// Iterate through results
	for _, hit := range searchResult.Hits.Hits {
		// hit.Index contains the name of the index

		// Deserialize hit.Source into a Tweet (could also be just a map[string]interface{}).
		t := Document{}
		err := json.Unmarshal(*hit.Source, &t)
		if err != nil {
			// Deserialization failed
			fmt.Println(err)
		}

		t.Id = hit.Id
		t.ContentHighlight = strings.Join(hit.Highlight["content"], "<br>")
		docs = append(docs, t)

		//fmt.Println(hit.Highlight["content"])

		// Work with tweet
		//fmt.Printf("%s\n%s\n%d\n----\n", t.Title, t.Genre, t.Page)
	}

	// Deserialize aggregations
	if agg, found := searchResult.Aggregations.Terms("Türü"); found {
		for _, bucket := range agg.Buckets {
			fmt.Println(bucket.Key, bucket.DocCount)
			genreFacet = append(genreFacet, Facet{bucket.Key.(string), bucket.DocCount})
		}
	}

	// Deserialize aggregations
	if agg, found := searchResult.Aggregations.Terms("Basım yılı"); found {
		for _, bucket := range agg.Buckets {
			fmt.Println(bucket.Key, bucket.DocCount)
			yearFacet = append(yearFacet, Facet{strconv.FormatFloat(bucket.Key.(float64), 'f', 0, 64), bucket.DocCount})
		}
	}

	//fmt.Printf("%+v", searchResult.Suggest["phrase_suggestion"])

	hasResult := true
	if searchResult.TotalHits() == 0 {
		hasResult = false
	}

	data["q"] = keywords
	data["docs"] = docs
	data["yearFacet"] = yearFacet
	data["genreFacet"] = genreFacet
	data["filters"] = filters
	data["hasResult"] = hasResult

	if len(searchResult.Suggest["phrase_suggestion"][0].Options) > 0 {
		data["suggest_text"] = searchResult.Suggest["phrase_suggestion"][0].Options[0].Text
		data["suggest_hl"] = searchResult.Suggest["phrase_suggestion"][0].Options[0].Highlighted
	}
	data["pages"] = paginate(start, 10, int(searchResult.TotalHits()))

	return data
}

//func query(keywords string, filterMap map[string]string) map[string]interface{} {
func queryDictionary(keywords string) (DictionaryEntry, bool) {

	ctx := context.Background()
	url := "http://127.0.0.1:9200"

	//Create an Elasticsearch client
	client, err := elastic.NewClient(elastic.SetURL(url), elastic.SetSniff(true))
	if err != nil {
		log.Fatal(err)
	}

	entry := DictionaryEntry{}

	termQuery := elastic.NewTermQuery("eng", keywords)

	printQuery(termQuery)

	search := client.Search().
		Index("dictionary"). // search in index "twitter"
		Query(termQuery).    // specify the query
		From(0).Size(1)      // take documents 0-9

	searchResult, err := search.Do(ctx) // execute
	if err != nil {
		fmt.Println(err)
		return entry, false
	}

	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from Elasticsearch.
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

	fmt.Println("TotalHits:", searchResult.TotalHits())

	hasResult := false

	if searchResult.TotalHits() >= 1 {
		fmt.Println("evet buldum")
		err := json.Unmarshal(*searchResult.Hits.Hits[0].Source, &entry)
		fmt.Println(entry)

		if err != nil {
			// Deserialization failed
			fmt.Println(err)
			return entry, hasResult
		}

		hasResult = true
	}

	fmt.Println(entry)

	return entry, hasResult
}
