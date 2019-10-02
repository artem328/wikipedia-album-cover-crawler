package crawler

import (
	"bytes"
	"golang.org/x/net/html"
	"strconv"
	"strings"
)

type Parser struct {
}

type PageResult struct {
	urls        []string
	nextPageUrl string
}

type AlbumDetails struct {
	year     int
	coverUrl string
	artist   string
	name     string
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) parseListItems(body []byte) *PageResult {
	document, _ := html.Parse(bytes.NewReader(body))

	pages := p.getElementById(document, "mw-pages")
	category := p.getElementsByClassName(pages, "mw-category-group")

	if len(category) == 0 {
		return nil
	}

	categoryGroups := p.getElementsByClassName(category[0], "mw-category-group")

	if len(categoryGroups) == 0 {
		return nil
	}

	urls := make([]string, 0)

	for _, categoryGroup := range categoryGroups {
		items := p.getElementsByTag(categoryGroup, "a")

		for _, item := range items {
			href := p.getAttribute(item, "href")

			if href == "" {
				continue
			}

			urls = append(urls, createUrl(href))
		}
	}

	nextPage := p.getLastChildByTag(pages, "a")
	nextPageUrl := ""

	if nextPage != nil && p.getText(nextPage) == "next page" {
		href := p.getAttribute(nextPage, "href")

		if href != "" {
			nextPageUrl = createUrl(href)
		}
	}

	return &PageResult{
		urls:        urls,
		nextPageUrl: nextPageUrl,
	}
}

func (p *Parser) parseAlbumDetails(body []byte) *AlbumDetails {
	document, _ := html.Parse(bytes.NewReader(body))

	infobox := p.getElementsByClassName(document, "infobox")

	if len(infobox) == 0 {
		return nil
	}

	rows := p.getElementsByTag(infobox[0], "tr")

	if len(rows) < 3 {
		return nil
	}

	name := p.getText(rows[0])
	if name == "" {
		return nil
	}

	cover := p.getElementsByTag(rows[1], "img")

	if len(cover) == 0 {
		return nil
	}

	coverUrl := p.getAttribute(cover[0], "src")

	if coverUrl == "" {
		return nil
	}

	artist := p.getText(rows[2])

	artistText := strings.Split(artist, "by ")
	artistText = artistText[1:]

	year := 0

	if len(rows) > 3 {
		th := p.getElementsByTag(rows[3], "th")
		td := p.getElementsByTag(rows[3], "td")
		if len(th) > 0 && len(td) > 0 && p.getText(th[0]) == "Released" {
			date := strings.Trim(p.getText(td[0]), " ")
			if len(date) > 3 {
				yearString := date[len(date)-4:]

				yearNum, err := strconv.Atoi(yearString)

				if err == nil {
					year = yearNum
				}
			}
		}
	}

	return &AlbumDetails{
		year:     year,
		name:     name,
		coverUrl: createUrl(coverUrl),
		artist:   strings.Join(artistText, "by "),
	}
}

func (p *Parser) getElementById(node *html.Node, id string) *html.Node {
	all := p.getElementsByAttribute(node, "id", id)

	if len(all) == 0 {
		return nil
	}

	return all[0]
}

func (p *Parser) getAttribute(node *html.Node, attribute string) string {
	for _, attr := range node.Attr {
		if attr.Key == attribute {
			return attr.Val
		}
	}

	return ""
}

func (p *Parser) getElementsByClassName(node *html.Node, class string) []*html.Node {
	return p.getElementsByFunc(node, func(n *html.Node) bool {
		classAttr := p.getAttribute(n, "class")

		if classAttr == "" {
			return false
		}

		classes := strings.Split(classAttr, " ")

		for _, className := range classes {
			if className == class {
				return true
			}
		}

		return false
	})
}

func (p *Parser) getElementsByAttribute(node *html.Node, attribute, value string) []*html.Node {
	return p.getElementsByFunc(node, func(n *html.Node) bool {
		return p.getAttribute(n, attribute) == value
	})
}

func (p *Parser) getElementsByTag(node *html.Node, tag string) []*html.Node {
	return p.getElementsByFunc(node, func(n *html.Node) bool {
		return n.Data == tag
	})
}

func (p *Parser) getLastChildByTag(node *html.Node, tag string) *html.Node {
	for c := node.LastChild; c != nil; c = c.PrevSibling {
		if c.Type == html.ElementNode && c.Data == tag {
			return c
		}
	}

	return nil
}

func (*Parser) getElementsByFunc(node *html.Node, f func(n *html.Node) bool) []*html.Node {
	var elements []*html.Node
	var lookupFunc func(*html.Node)

	lookupFunc = func(n *html.Node) {
		if n.Type == html.ElementNode && f(n) {
			elements = append(elements, n)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			lookupFunc(c)
		}
	}

	lookupFunc(node)

	return elements
}

func (p *Parser) getText(node *html.Node) string {
	text := ""

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.ElementNode:
			text += p.getText(c)
		case html.TextNode:
			text += c.Data
		}
	}

	return text
}
