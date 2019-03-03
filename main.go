package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var (
	VERSION string = "UNKNOWN"
	HASH    string = "UNKNOWN"

	version, d    bool
	level uint
	wait, timeout time.Duration

	processedURLs map[string]bool
	ticker        *time.Ticker
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s [options...] [URL...]:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.BoolVar(&version, "version", false, "print the version and exit")
	flag.BoolVar(&d, "debug", false, "print debug information")
	flag.UintVar(&level, "level", 0, "specify recursion maximum depth level depth")
	flag.DurationVar(&wait, "wait", time.Millisecond*500, "wait between request")
	flag.DurationVar(&timeout, "timeout", time.Second*5, "set timeout values")
	flag.Parse()

	processedURLs = make(map[string]bool)
}

func main() {
	if version {
		fmt.Printf("spider-web %s (%s)\n", VERSION, HASH)
		os.Exit(0)
	}

	if flag.NArg() == 0 {
		flag.Usage()
	}

	debug("start with flag debug")

	for _, a := range flag.Args() {
		u, err := url.Parse(a)
		if err != nil {
			log.Printf("argument %s not url: %s", a, err)
			continue
		}
		if u.Host == "" {
			log.Printf("url %s not correct: %s", a, err)
			continue
		}

		urlFormat(u)

		search(u, int(level))

		for k, _ := range processedURLs {
			fmt.Println(k)
		}
	}
}

func urlFormat(u *url.URL) {
	u.Host = strings.TrimPrefix(u.Host, "www.")
	u.Fragment = ""
}

func search(parentUrl *url.URL, lvl int) {
	if lvl < 0 {
		return
	}
	if processedURLs[parentUrl.String()] {
		return
	}
	processedURLs[parentUrl.String()] = true

	debugf("processing %s:\n", parentUrl)

	body, err := getPage(parentUrl)
	if err != nil {
		debug(err)
		return
	}

	chieldUrls, err := getUrls(body)
	if err != nil {
		debug(err)
		return
	}

	var urls []*url.URL

	for _, u := range chieldUrls {
		chieldUrl, err := url.Parse(u)
		if err != nil {
			debug(err)
			continue
		}

		urlFormat(chieldUrl)

		if chieldUrl.Host == "" {
			chieldUrl.Host = parentUrl.Host
		} else if chieldUrl.Host == parentUrl.Host {
			debugf("\t%s\n", chieldUrl.String())
			if !processedURLs[chieldUrl.String()] {
				urls = append(urls, chieldUrl)
			}
		}

	}

	for _, u := range urls {
		if level == 0 {
			search(u, lvl)
		} else if lvl > 0 {
			search(u, lvl-1)
		}
	}
}

func getPage(u *url.URL) (io.ReadCloser, error) {
	c := &http.Client{
		Timeout: timeout,
	}

	if u.Scheme == "https" {
		c.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	if ticker != nil {
		<-ticker.C
	}

	debugf("request to %s", u.String())
	resp, err := c.Get(u.String())
	if err != nil {
		return nil, err
	}

	if wait.Seconds() != 0 {
		ticker = time.NewTicker(wait)
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

func searchUrl(n *html.Node) (urls []string) {
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

func debug(v ...interface{}) {
	if d {
		log.Print(v...)
	}
}

func debugf(f string, v ...interface{}) {
	if d {
		log.Printf(f, v...)
	}
}
