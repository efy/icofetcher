package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/biessek/golang-ico"
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

type Icon struct {
	ImageConfig *image.Config
	Image       *image.Image
	URL         *url.URL
	Rel         string
	Mime        string
}

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

	if len(hrefs) == 0 {
		hrefs = append(hrefs, "/favicon.ico")
	}

	var icons []Icon

	for _, v := range hrefs {
		var u *url.URL
		var err error
		icon := Icon{}

		// Need more exhaustive checking of the returned href
		if strings.HasPrefix(v, "http") || strings.HasPrefix(v, "https") {
			u, err = url.Parse(v)
		} else {
			u, err = url.Parse(*furl + v)
		}
		if err != nil {
			fmt.Println(err)
			continue
		}

		resp, err := c.Get(u.String())
		if err != nil {
			fmt.Println("failed to download:", u.String())
			os.Exit(1)
		}

		ct := resp.Header.Get("Content-Type")
		mt, _, err := mime.ParseMediaType(ct)
		if err != nil {
			fmt.Println(err)
			continue
		}

		icon.URL = u
		icon.Mime = mt

		switch {
		case mt == "image/x-icon" || mt == "image/vnd.microsoft.icon":
			icoConf, err := ico.DecodeConfig(resp.Body)
			if err != nil {
				fmt.Println(err)
				continue
			}
			icoImg, err := ico.Decode(resp.Body)
			if err != nil {
				fmt.Println(err)
				continue
			}
			icon.ImageConfig = &icoConf
			icon.Image = &icoImg
			icons = append(icons, icon)
		case mt == "image/png":
			pngImage, err := png.DecodeConfig(resp.Body)
			if err != nil {
				fmt.Println(err)
				continue
			}
			icon.ImageConfig = &pngImage
			icons = append(icons, icon)
		default:
			fmt.Println("cannot handle mimetype:", mt)
		}

		resp.Body.Close()
	}

	for _, icon := range icons {
		fmt.Printf("%v width: %v, height: %v\n", icon.URL, icon.ImageConfig.Width, icon.ImageConfig.Height)
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
