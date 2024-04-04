package comment_parser

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type db interface {
	InsertNewSource(hostname, repository, typename string, occurrence int)
	InsertNewAliasEntry(repository, url, typename string, withAlias, isClosed bool, seconds float64)
}

type CommentParser struct {
	db db
}

func New(db db) *CommentParser {
	return &CommentParser{
		db: db,
	}
}

func (c *CommentParser) countSourceOccurrences(links []string) map[string]int {
	sourceOccurrences := make(map[string]int)

	for _, link := range links {
		parsedURL, err := url.Parse(link)
		if err != nil {
			fmt.Printf("Error parsing URL: %v\n", err)
			continue
		}

		host := parsedURL.Hostname()

		if host == "github.com" {
			continue
		}

		sourceOccurrences[host]++
	}

	return sourceOccurrences
}

func (c *CommentParser) fetchHTML(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error fetching URL: %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	return string(body), nil
}

func (c *CommentParser) checkTimeByAlias(repository, url, typename string, withAlias bool, createdAt, closedAt *time.Time) {
	if closedAt == nil {
		c.db.InsertNewAliasEntry(repository, url, typename, withAlias, false, time.Now().Sub(*createdAt).Seconds())
		return
	}

	seconds := closedAt.Sub(*createdAt).Seconds()
	c.db.InsertNewAliasEntry(repository, url, typename, withAlias, true, seconds)
}

func (c *CommentParser) parseHTMLLinks(htmlContent string, repository, u, typename string, createdAt, closedAt *time.Time) []string {
	var links []string
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}

	var extractLinks func(*html.Node)
	commentNo := 1
	withAlias := false
	extractLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && len(n.Attr) != 0 && n.Attr[0].Key == "dir" && n.Attr[0].Val == "auto" {
			for comment := n.FirstChild; comment != nil; comment = comment.NextSibling {
				if comment.Type == html.ElementNode && len(comment.Attr) != 0 && comment.Attr[0].Key == "href" {
					link := comment.Attr[0].Val
					links = append(links, link)

					// checking if we have aliases for users in the comment
					if commentNo == 1 {
						parsedURL, err := url.Parse(link)
						if err != nil {
							log.Fatalf("Parse User URL")
						}

						host := parsedURL.Hostname()
						if host == "github.com" {
							// in this case - it is alias for the user or org
							withAlias = strings.Count(link, "/") == 3
						}
					}
				}
			}
		}

		if commentNo == 1 {
			c.checkTimeByAlias(repository, u, typename, withAlias, createdAt, closedAt)
		}
		commentNo++

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractLinks(c)
		}
	}
	extractLinks(doc)

	return links
}

func (c *CommentParser) HandleCommentsByURL(repository, url, typename string, createdAt, closedAt *time.Time) {
	log.Println("parsing", repository, url, typename)

	htmlContent, err := c.fetchHTML(url)
	if err != nil {
		log.Fatal(err)
	}

	links := c.parseHTMLLinks(htmlContent, repository, url, typename, createdAt, closedAt)

	occurrences := c.countSourceOccurrences(links)

	for hostname, occurrence := range occurrences {
		c.db.InsertNewSource(hostname, repository, typename, occurrence)
	}

	log.Println("number of external links:", occurrences)
}
