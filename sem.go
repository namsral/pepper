package main

import (
    "fmt"
    "errors"
    "strings"
)

// A SearchEngine defines parameters for a keyword defined search engine
type SearchEngine struct {
    URL string
    Title string
    Keyword string
}

// A SEManager defines a list of search engines and methods to manipulate them
type SEManager struct {
    Engines []SearchEngine
    Default SearchEngine
}

// Add a new search engine
func (sem SEManager) Add(se SearchEngine) error {
    sem.Engines = append(sem.Engines, se)
    return nil
}

// Get the search engine which handles `keyword`
func (sem SEManager) Get(keyword string) (SearchEngine, error) {
    if len(keyword) < 1 {
        return SearchEngine{}, errors.New("Keyword too short")
    }
    for _, se := range sem.Engines {
        if keyword == se.Keyword {
            return se, nil
        }
    }
    return SearchEngine{}, errors.New(fmt.Sprintf("No search engine with keyword %s found", keyword))
}

// Return URL string based on `query`
func (sem SEManager) Query(query string) (string, error) {
    words := strings.Split(query, " ")

    switch len(words){
    case 0:
        return "", errors.New("Query strings does not contain any words")
    case 1:
        return fmt.Sprintf(sem.Default.URL, words[0]), nil
    }

    se, err := sem.Get(words[0])
    if err != nil {
        return fmt.Sprintf(sem.Default.URL, strings.Join(words, " ")), nil  
    }
    return fmt.Sprintf(se.URL, strings.Join(words[1:], " ")), nil
}

// NewSEManager allocates and returns a new SEManager with a default search engine
func NewSEManager() *SEManager {
    return &SEManager{
        Engines: []SearchEngine{},
        Default: SearchEngine{
            URL: "https://duckduckgo.com/?q=%s",
            Title: "DuckDuckGo",
            Keyword: "duck",
        },
    }
}