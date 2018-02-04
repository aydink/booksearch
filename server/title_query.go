package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/olivere/elastic"
)

//func query(keywords string, filterMap map[string]string) map[string]interface{} {
func titleQuery(keywords string, start int, filters [][3]string) map[string]interface{} {

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
	boolQuery = boolQuery.Must(elastic.NewMatchQuery("titlefull", keywords).Operator("AND"))
	//boolQuery = boolQuery.MinimumShouldMatch("1")

	for i := 0; i < len(filters); i++ {
		boolQuery = boolQuery.Filter(elastic.NewTermQuery(filters[i][0], filters[i][2]))
		//fmt.Println("filtre eklendi")
	}

	printQuery(boolQuery)

	higlight := elastic.NewHighlight()
	higlight = higlight.Field("title")
	higlight = higlight.FragmentSize(200)
	higlight = higlight.NumOfFragments(1)

	printQuery(higlight)

	aggsYear := elastic.NewTermsAggregation()
	aggsYear = aggsYear.Field("year")

	aggsGenre := elastic.NewTermsAggregation()
	aggsGenre = aggsGenre.Field("genre")

	aggsDepartment := elastic.NewTermsAggregation()
	aggsDepartment = aggsDepartment.Field("department")

	aggsCategory := elastic.NewTermsAggregation()
	aggsCategory = aggsCategory.Field("category")

	/*
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
	*/

	//printQuery(aggsGenre)
	//printQuery(aggsYear)

	/*
		postFilter := elastic.NewBoolQuery()
		postFilter = postFilter.Must(elastic.NewTermQuery("year", 2000))
		postFilter = postFilter.Must(elastic.NewTermQuery("genre", "Korku"))

		printQuery(postFilter)
	*/

	search := client.Search().
		Index("ray").                           // search in index "twitter"
		Query(boolQuery).                       // specify the query
		From(start).Size(10).                   // take documents 0-9
		Pretty(true).                           // pretty print request and response JSON
		Highlight(higlight).                    // Highlight results
		Aggregation("Basım yılı", aggsYear).    // Aggregation basım yılı
		Aggregation("Türü", aggsGenre).         // Aggregation Genre
		Aggregation("Kuvveti", aggsDepartment). // Aggregation Department
		Aggregation("Kategori", aggsCategory)   // Aggregation Category
		//Suggester(suggester)
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

	books := make([]Book, 0, 10)
	yearFacet := make([]Facet, 0)
	genreFacet := make([]Facet, 0)
	departmentFacet := make([]Facet, 0)
	categoryFacet := make([]Facet, 0)

	// Iterate through results
	for _, hit := range searchResult.Hits.Hits {
		// hit.Index contains the name of the index

		// Deserialize hit.Source into a Tweet (could also be just a map[string]interface{}).
		b := Book{}
		err := json.Unmarshal(*hit.Source, &b)
		if err != nil {
			// Deserialization failed
			fmt.Println(err)
		}

		b.Id = hit.Id
		//t.ContentHighlight = strings.Join(hit.Highlight["title"], "<br>")
		books = append(books, b)

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
	if agg, found := searchResult.Aggregations.Terms("Kuvveti"); found {
		for _, bucket := range agg.Buckets {
			fmt.Println(bucket.Key, bucket.DocCount)
			departmentFacet = append(departmentFacet, Facet{bucket.Key.(string), bucket.DocCount})
		}
	}

	// Deserialize aggregations
	if agg, found := searchResult.Aggregations.Terms("Kategori"); found {
		for _, bucket := range agg.Buckets {
			fmt.Println(bucket.Key, bucket.DocCount)
			categoryFacet = append(categoryFacet, Facet{bucket.Key.(string), bucket.DocCount})
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
	data["books"] = books
	data["yearFacet"] = yearFacet
	data["genreFacet"] = genreFacet
	data["departmentFacet"] = departmentFacet
	data["categoryFacet"] = categoryFacet
	data["filters"] = filters
	data["hasResult"] = hasResult

	/*
		if len(searchResult.Suggest["phrase_suggestion"][0].Options) > 0 {
			data["suggest_text"] = searchResult.Suggest["phrase_suggestion"][0].Options[0].Text
			data["suggest_hl"] = searchResult.Suggest["phrase_suggestion"][0].Options[0].Highlighted
		}
	*/
	data["pages"] = paginate(start, 10, int(searchResult.TotalHits()))

	return data
}
