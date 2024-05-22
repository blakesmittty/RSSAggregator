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
	PubDate     string `xml:"pubDate"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

func createHTMLHeader(title string, file *os.File) {
	header := `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>` + title + `</title></head>`
	file.WriteString(header)
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

func createFeedPage(rss RSS, feed Feed) {
	pageName := feed.File
	page, err := os.Create(pageName)
	if err != nil {
		fmt.Printf("Error creating page: %v", err)
		return
	}
	defer page.Close()
	createHTMLHeader(pageName, page)
	header := `<body><h1><a href="` + rss.Channel.Link + `">` + rss.Channel.Title + `</a></h1><p>` + rss.Channel.Description + `</p>`
	tableHeader := `<table border="1"><tbody><tr><th>Date</th><th>Source</th><th>News</th></tr>`
	page.WriteString(header)
	page.WriteString(tableHeader)
	for i := 0; i < len(rss.Channel.Items); i++ {
		item := rss.Channel.Items[i]
		page.WriteString("<tr>")
		tableDate := `<td>` + item.PubDate + `</td>`
		tableSource := `<td><a href="` + feed.URL + `">` + rss.Channel.Title + `</a></td>`
		tableNews := `<td><a href="` + item.Link + `">` + item.Description + `</a></td>`
		page.WriteString(tableDate)
		page.WriteString(tableSource)
		page.WriteString(tableNews)
		page.WriteString("</tr>")
	}
	page.WriteString("</tbody></table></body></html>")
}

func processAndCreateFeedPages(feeds Feeds) {
	for i := 0; i < len(feeds.Feeds); i++ {
		var rss RSS
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
			return
		}
		createFeedPage(rss, feed)
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

	createHTMLHeader(feeds.Title, file)
	createIndexBody(feeds, file)
	processAndCreateFeedPages(feeds)

}
