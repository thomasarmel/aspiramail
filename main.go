package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sync"
)

const URLDONESIZE = 10000
var urlDone [URLDONESIZE]url.URL
var posArr=0
var nbThreads=0

func webRecursive(url url.URL, newWg *sync.WaitGroup, threaded bool) {
	if(threaded) {
		defer newWg.Done()
	}
	query := ""
	if(url.RawQuery != "") {
		query = "?"+url.RawQuery
	}
	urlStr := url.Scheme+"://"+url.Host+path.Clean(url.Path)+query
	resp, err := http.Get(urlStr);
	if err != nil {
		return
	}
	if(resp.ContentLength > 10000000) {
		return
	}
	defer resp.Body.Close();
	if(resp.Status[0] != '2') {
		return
	}
	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	var result string
	for scanner.Scan() {
		result+=scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		return
	}
	var getUrlRegex, _ = regexp.Compile("href=\\\"((http(s?)://((([a-z0-9]*)\\.)?)([a-z0-9]*)\\.([a-z0-9]*))?)(([a-zA-Z0-9/\\.\\?=&:;\\%_#]*)?)\\\"")
	var getEmailRegex, _ = regexp.Compile("[\\w-\\.]+(@|&#x40;|&#64;|%40|\\\\u0040)([\\w-]+\\.)+[\\w-]{2,4}")
	resultURLRegex := getUrlRegex.FindAllStringSubmatch(result, -1)
	resultEmailRegex := getEmailRegex.FindAllStringSubmatch(result, -1)
	for _, s := range resultEmailRegex {
		fmt.Println(s[0])
	}

	var wg sync.WaitGroup

	for _, s := range resultURLRegex {
		runes:=[]rune(s[0])
		newUrlStr:=string(runes[6:len(s[0])-1])

		if(len(newUrlStr) == 0) {
			continue
		}

		newUrl, _ :=url.Parse(newUrlStr)
		if(newUrl.Scheme=="") {
			newUrl.Scheme=url.Scheme
		}
		if(newUrl.Host=="") {
			newUrl.Host=url.Host
		}

		if(!deriveContainsFoo(urlDone, *newUrl)) {
			urlDone[posArr]=*newUrl
			posArr++
			if(posArr>=URLDONESIZE) {
				posArr=posArr%URLDONESIZE
			}
			if(nbThreads < 100) {
				wg.Add(1)
				go webRecursive(*newUrl, &wg, true)
				nbThreads++
			} else {
				webRecursive(*newUrl, &wg, false)
			}
		}
	}
	wg.Wait()
	if(threaded) {
		nbThreads--
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	firstURL, _ := url.Parse("https://www.google.fr/")
	nbThreads++
	go webRecursive(*firstURL, &wg, true)
	wg.Wait()
}

func deriveContainsFoo(list [URLDONESIZE]url.URL, item url.URL) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}