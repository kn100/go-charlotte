package fetch

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/net/html"
)

/*
JobResult stores the result of one link retrieval.
*/
type JobResult struct {
	FromURL *url.URL
	LinksTo []*url.URL
}

/*
Links returns a list of JobResults - each one containing the results for one queue entry
*/
func Links(queue []*url.URL) []JobResult {
	var jobResults []JobResult
	done := make(chan JobResult)
	var wg sync.WaitGroup
	for len(queue) > 0 {
		wg.Add(1)
		toProcess := queue[0]
		queue = queue[1:]
		go getLinksForSingleURL(toProcess, done)
	}
	go linkConsumer(done, &jobResults, &wg)
	wg.Wait()
	return jobResults
}

/*
getLinksForSingleURL is the 'job' that Links runs. It returns the JobResult via the channel
*/
func getLinksForSingleURL(url *url.URL, done chan JobResult) {
	links := JobResult{FromURL: url, LinksTo: nil}

	resp, err := http.Get(url.String())
	if err != nil {
		log.Printf("Loading failed for link %s. Pretending it has no links.\n", url.String())
		done <- links
		return
	}
	defer resp.Body.Close()
	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			err := z.Err()
			if err == io.EOF {
				//end of the file, break out of the loop
				done <- links
				return
			}
			// There's been an error. We should probably deal with this more gracefully, but for now log and return what we did get.
			log.Println("There was an error parsing the html.", err)
			done <- links
			return

		case tt == html.StartTagToken:
			t := z.Token()

			isAnchor := t.Data == "a"
			if isAnchor {
				link := getHref(t)
				if link != "" {
					foundLink, err := url.Parse(link)
					if err != nil {
						log.Printf("Wasn't able to parse %s. Error %s\n", link, err)
						return
					}
					links.LinksTo = append(links.LinksTo, foundLink)
				}
			}
		}
	}
}

/*
getHref will when given a html.Token, find the href key and returns it.
boo
*/
func getHref(t html.Token) string {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == "href" {
			return a.Val
		}
	}
	return ""
}

func linkConsumer(j chan JobResult, results *[]JobResult, wg *sync.WaitGroup) {
	for s := range j {
		*results = append(*results, s)
		wg.Done()
	}
}
