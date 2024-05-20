package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Feeds struct {
	XMLName xml.Name `xml:"feeds"`
	Title   string   `xml:"title,attr"`
	Feeds   []Feed   `xml:"feed"`
}

type Feed struct {
	URL  string `xml:"url,attr"`
	Name string `xml:"name,attr"`
	File string `xml:"file,attr"`
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

func createIndexHeader(title string, index *os.File) {
	header := `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>` + title + `</title></head>`
	index.WriteString(header)
}

func createIndexBody(feeds Feeds, index *os.File) {
	index.WriteString("<body><ul>")
	title := "<h1>" + feeds.Title + "</h1>"
	index.WriteString(title)
	for i := 0; i < len(feeds.Feeds); i++ {
		feed := feeds.Feeds[i]
		content := `<li><a href="` + feed.File + `">` + feed.Name + `</a></li>`
		index.WriteString(content)
	}
	index.WriteString("</body></ul></html>")
}

func createItemPage(feeds Feeds) {
	var rss RSS
	for i := 0; i < len(feeds.Feeds); i++ {
		feed := feeds.Feeds[i]
		response, httpErr := http.Get(feed.URL)
		if httpErr != nil {
			fmt.Printf("Error fetching feed: %v\n", httpErr)
			return
		}
		defer response.Body.Close()
		body, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			fmt.Printf("Error reading response body: %v\n", readErr)
			return
		}
		parseErr := xml.Unmarshal(body, &rss)
		if parseErr != nil {
			fmt.Printf("Error parsing RSS document: %v", parseErr)
		}

	}
}

func main() {
	var input string
	fmt.Print("Enter URL to feeds: ")
	fmt.Scan(&input)
	response, httpErr := http.Get(input)
	if httpErr != nil {
		fmt.Printf("Error fetching feeds: %v\n", httpErr)
		return
	}
	defer response.Body.Close()

	body, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		fmt.Printf("Error reading response body: %v\n", readErr)
		return
	}

	var feeds Feeds
	parseErr := xml.Unmarshal(body, &feeds)
	if parseErr != nil {
		fmt.Printf("Error parsing feeds: %v", parseErr)
		return
	}
	fmt.Print("\n")

	var fileName string
	fmt.Print("Enter the name of the index file: ")
	fmt.Scan(&fileName)
	fileName += ".html"
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("Error in file creation: %v\n", err)
		return
	}
	defer file.Close()

	createIndexHeader(feeds.Title, file)
	createIndexBody(feeds, file)

}
