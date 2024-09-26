package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Feeds represents the structure XML document containing the feeds the user wants
type Feeds struct {
	XMLName xml.Name `xml:"feeds"`
	Title   string   `xml:"title,attr"`
	Feeds   []Feed   `xml:"feed"`
}

// Feed represents an individual feed from the Feeds document
type Feed struct {
	URL  string `xml:"url,attr"`
	Name string `xml:"name,attr"`
	File string `xml:"file,attr"`
}

// RSS represents the structure of an RSS document; contains only the elements needed
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

// Item represents an individual <item> element from RSS channel
type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// Channel represents the structure of the <channel> element in an RSS document
type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

/**
 * createHTMLHeader creates the needed header for each document created with its corresponding title
 *
 * param title string - the attribute value from "name" in the feeds document
 * param file *os.File - the file that is being written to
 */
func createHTMLHeader(title string, file *os.File) {
	header := `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>` + title + `</title></head>`
	file.WriteString(header)
}

/**
 * createIndexBody creates the body of the index file
 *
 * param - feeds Feeds - the feeds data used to create links to each of their pages
 * param - index *os.File - the index file being written to
 */
func createIndexBody(feeds Feeds, index *os.File) {
	index.WriteString("<body><ul>")
	title := "<h1>" + feeds.Title + "</h1>"
	index.WriteString(title)
	// iterate to create a link to every feed page
	for i := 0; i < len(feeds.Feeds); i++ {
		feed := feeds.Feeds[i]
		content := `<li><a href="` + feed.File + `">` + feed.Name + `</a></li>`
		index.WriteString(content)
	}
	index.WriteString("</body></ul></html>")
}

/**
 * createFeedPage creates the page for a singular RSS feed passed into it
 *
 * param rss RSS - the corresponding RSS feed structure
 * param feed Feed - the feed element from feeds that contains the corresponding RSS document
 */
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
	// iterate to find every item and write a table row for it
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
	page.WriteString("</tbody></table></body></html>") // write appropriate closing tags once the table is made
}

/**
 * processAndCreateFeedPages creates a page for every <feed> element in <feeds> from the users XML document
 *
 * param feeds Feeds - the <feeds> element that contains the individual <feed>s
 */
func processAndCreateFeedPages(feeds Feeds) {
	for i := 0; i < len(feeds.Feeds); i++ {
		var rss RSS // declared within loop in order to prevent data accumulation
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
		createFeedPage(rss, feed) // create page with the rss document processed along with its corresponding <feed> element from feeds
	}
}

/**
 * handleAndProcessUserInput prompts the user for the URL to their feeds XML document containing the feeds
 * they want to aggregate. It then processes the document into Feeds and creates the index page with
 * each feed
 *
 * return feeds Feeds - the processed feeds XML document
 * return file *os.File - the index file that is created
 */
func handleAndProcessUserInput() (feeds Feeds, file *os.File) {
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
	return feeds, file
}

func main() {
	feeds, file := handleAndProcessUserInput()
	if file != nil {
		defer file.Close()
		createHTMLHeader(feeds.Title, file)
		createIndexBody(feeds, file)
		processAndCreateFeedPages(feeds)
	}
}
