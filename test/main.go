package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"tiedot/db"
	"tiedot/dberr"
)

func sameMap(m1 map[string]interface{}, m2 map[string]interface{}) bool {
	if str1, err := json.Marshal(m1); err != nil {
		panic(err)
	} else if str2, err := json.Marshal(m2); err != nil {
		panic(err)
	} else if string(str1) != string(str2) {
		return false
	}
	return true
}

func main() {
	myDBDir := "../tmp/tiedot_test_embeddedExample"
	os.RemoveAll(myDBDir)
	defer os.RemoveAll(myDBDir)

	myDB, err := db.OpenDB(myDBDir)
	if err != nil {
		log.Fatal(err)
	}
	if err := myDB.Create("Feeds"); err != nil {
		log.Fatal(err)
	}
	if err := myDB.Create("Votes"); err != nil {
		log.Fatal(err)
	}
	for _, name := range myDB.AllCols() {
		if name != "Feeds" && name != "Votes" {
			log.Fatal(myDB.AllCols())
		}
	}
	if err := myDB.Rename("Votes", "Points"); err != nil {
		log.Fatal(err)
	}
	if err := myDB.Drop("Points"); err != nil {
		log.Fatal(err)
	}
	if err := myDB.Scrub("Feeds"); err != nil {
		log.Fatal(err)
	}

	feeds := myDB.Use("Feeds")
	docID, err := feeds.Insert(map[string]interface{}{
		"name": "Go 1.2 is released",
		"url":  "golang.org"})
	if err != nil {
		log.Fatal(err)
	}
	readBack, err := feeds.Read(docID)
	if err != nil {
		log.Fatal(err)
	}
	if !sameMap(readBack, map[string]interface{}{
		"name": "Go 1.2 is released",
		"url":  "golang.org"}) {
		log.Fatal(readBack)
	}
	err = feeds.Update(docID, map[string]interface{}{
		"name": "Go is very popular",
		"url":  "google.com"})
	if err != nil {
		panic(err)
	}
	feeds.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		var doc map[string]interface{}
		if json.Unmarshal(docContent, &doc) != nil {
			log.Fatal("cannot deserialize")
		}
		if !sameMap(doc, map[string]interface{}{
			"name": "Go is very popular",
			"url":  "google.com"}) {
			log.Fatal(doc)
		}
		return true
	})
	if err := feeds.Delete(docID); err != nil {
		log.Fatal(err)
	}
	if err := feeds.Delete(docID); dberr.Type(err) != dberr.ErrorNoDoc {
		log.Fatal(err)
	}

	if err := feeds.Index([]string{"author", "name", "first_name"}); err != nil {
		log.Fatal(err)
	}
	if err := feeds.Index([]string{"Title"}); err != nil {
		log.Fatal(err)
	}
	if err := feeds.Index([]string{"Source"}); err != nil {
		log.Fatal(err)
	}
	for _, path := range feeds.AllIndexes() {
		joint := strings.Join(path, "")
		if joint != "authornamefirst_name" && joint != "Title" && joint != "Source" {
			log.Fatal(feeds.AllIndexes())
		}
	}
	if err := feeds.Unindex([]string{"author", "name", "first_name"}); err != nil {
		log.Fatal(err)
	}

	// ****************** Queries ******************
	// Prepare some documents for the query
	if _, err := feeds.Insert(map[string]interface{}{"Title": "New Go release", "Source": "golang.org", "Age": 3}); err != nil {
		log.Fatal("failure")
	}
	if _, err := feeds.Insert(map[string]interface{}{"Title": "Kitkat is here", "Source": "google.com", "Age": 2}); err != nil {
		log.Fatal("failure")
	}
	if _, err := feeds.Insert(map[string]interface{}{"Title": "Good Slackware", "Source": "slackware.com", "Age": 1}); err != nil {
		log.Fatal("failure")
	}
	var query interface{}
	if json.Unmarshal([]byte(`[{"eq": "New Go release", "in": ["Title"]}, {"eq": "slackware.com", "in": ["Source"]}]`), &query) != nil {
		log.Fatal("failure")
	}
	queryResult := make(map[int]struct{}) // query result (document IDs) goes into map keys
	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
		log.Fatal(err)
	}
	// Query result are document IDs
	for id := range queryResult {
		// To get query result document, simply read it
		readBack, err := feeds.Read(id)
		if err != nil {
			panic(err)
		}
		if !sameMap(readBack, map[string]interface{}{"Title": "New Go release", "Source": "golang.org", "Age": 3}) &&
			!sameMap(readBack, map[string]interface{}{"Title": "Good Slackware", "Source": "slackware.com", "Age": 1}) {
			log.Fatal(readBack)
		}
	}

	if err := myDB.Close(); err != nil {
		log.Fatal(err)
	}

}
