package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/net/html"
)

// Map of meta rel to value
var (
	rels = []string{
		"icon",
		"shortcut icon",
		"apple-touch-icon",
	}

	furl = flag.String("url", "", "The URL to fetch icons for")
)

func main() {
	flag.Parse()

	c := &http.Client{
		Timeout: 15 * time.Second,
	}

	if *furl == "" {
		fmt.Println("You must provide a url")
		os.Exit(1)
	}

	resp, err := c.Get(*furl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ct := resp.Header.Get("Content-Type")
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if mt != "text/html" {
		fmt.Println("unsupported media type:", mt)
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	hrefs := extractIconUrls(body, rels)

	for _, v := range hrefs {
		u, err := url.Parse(*furl + v)
		if err != nil {
			fmt.Println("err")
			continue
		}

		resp, err := c.Get(u.String())
		if err != nil {
			fmt.Println("failed to download get:", v)
		}
		defer resp.Body.Close()
		// Decode image
	}
}

// Takes a html byte array parses it extracting the href
// of link tags if their rel matches one of the whitelist
func extractIconUrls(body []byte, whitelist []string) []string {
	var hrefs []string

	tk := html.NewTokenizer(bytes.NewReader(body))
	for {
		tt := tk.Next()
		switch {
		case tt == html.ErrorToken:
			return hrefs
		case tt == html.StartTagToken || tt == html.SelfClosingTagToken:
			t := tk.Token()
			isLink := t.Data == "link"
			if isLink {
				rel := getAttrVal("rel", t)
				if rel != "" && stringInSlice(rel, whitelist) {
					href := getAttrVal("href", t)
					if href != "" {
						hrefs = append(hrefs, href)
					}
				}
			}
		}
	}

	return []string{}
}

func getAttrVal(key string, t html.Token) string {
	for _, attr := range t.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func stringInSlice(s string, sl []string) bool {
	for _, v := range sl {
		if v == s {
			return true
		}
	}
	return false
}
