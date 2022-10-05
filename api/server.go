package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Article ...
type Article struct {
	Title  string `json:"Title"`
	Author string `json:"author"`
	Link   string `json:"link"`
}

// Articles ...
var Articles []Article = []Article{
	{Title: "Python Intermediate and Advanced 101",
		Author: "Arkaprabha Majumdar",
		Link:   "https://www.amazon.com/dp/B089KVK23P"},
	{Title: "R programming Advanced",
		Author: "Arkaprabha Majumdar",
		Link:   "https://www.amazon.com/dp/B089WH12CR"},
	{Title: "R programming Fundamentals",
		Author: "Arkaprabha Majumdar",
		Link:   "https://www.amazon.com/dp/B089S58WWG"},
}

func Serve() error {
	return handleRequests()
}

func handleRequests() error {
	http.HandleFunc("/", homePage)
	// add our articles route and map it to our
	// returnAllArticles function like so
	http.HandleFunc("/articles", returnAllArticles)
	return http.ListenAndServe(":8000", nil)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func returnAllArticles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllArticles")
	json.NewEncoder(w).Encode(Articles)
}
