package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	VERSION string = "UNKNOWN"
	HASH    string = "UNKNOWN"

	version, debug bool
	level, tries uint
)

func init() {
	flag.BoolVar(&version, "version", false, "print the version and exit")
	flag.BoolVar(&debug, "debug", false, "print debug information")
	flag.UintVar(&level, "level", 1, "specify recursion maximum depth level depth")
	flag.UintVar(&tries, "tries", 0, "set number of tries to number")
	flag.Parse()
}

func main() {
	if version {
		fmt.Printf("version: %s (%s)\n", VERSION, HASH)
		os.Exit(0)
	}

	for _, a := range flag.Args() {
		u, err := url.Parse(a)
		if err != nil {
			fmt.Printf("error URL %s - %s", a, err)
			os.Exit(1)
		}
		if u.Host == "" {
			fmt.Printf("error URL %s - not correct hostname", a)
			os.Exit(1)
		}

		search(u)
	}
}

func search(parentUrl *url.URL) {
	parentUrl.Host = strings.TrimPrefix(parentUrl.Host, "www.")
	body, err := getPage(parentUrl)
	if err != nil {
		fmt.Println(err)
	}

	chieldUrls, err := getUrls(body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, u := range chieldUrls {
		chieldUrl, _ := url.Parse(u)
		chieldUrl.Host = strings.TrimPrefix(chieldUrl.Host, "www.")
		if chieldUrl.Host == parentUrl.Host {
			fmt.Println(u)
		}
	}
}

func getPage(u *url.URL) (io.ReadCloser, error) {
	var c *http.Client
	if u.Scheme == "https" {
		c = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	} else {
		c = &http.Client{}
	}
	resp, err := c.Get(u.String())
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func getUrls(body io.ReadCloser) ([]string, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return []string{}, err
	}

	return searchUrl(doc), nil
}

func searchUrl(n *html.Node) (urls []string){
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				urls = append(urls, attr.Val)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		for _, u := range searchUrl(c) {
			urls = append(urls, u)
		}
	}
	return urls
}
