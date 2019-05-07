// Package fetch can asynchronously request links and parse the links in the
// pages out. It's intended use is to spider a given set of pages
// asynchronously.
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
func Links(client *http.Client, queue []*url.URL) []JobResult {

	var jobResults []JobResult
	// This channel is used for communication between producers and the consumer.
	done := make(chan JobResult)
	var producerWaitGroup sync.WaitGroup
	var consumerWaitGroup sync.WaitGroup
	for len(queue) > 0 {
		producerWaitGroup.Add(1)

		// This is effectively a dequeue operation
		toProcess := queue[0]
		queue = queue[1:]

		// give the work to a goroutine to do it!
		go getLinksForSingleURL(client, toProcess, done, &producerWaitGroup)
	}
	consumerWaitGroup.Add(1)
	go linkConsumer(done, &jobResults, &consumerWaitGroup)
	// We cannot proceed until every producer has finished.
	producerWaitGroup.Wait()
	// Nothing else will be written to this channel, so close it. This will
	// trigger the consumer to quit after it's finished what it is doing.
	close(done)
	consumerWaitGroup.Wait()
	return jobResults
}

/*
getLinksForSingleURL is the 'job' that Links runs. It returns the JobResult via the channel
*/
func getLinksForSingleURL(client *http.Client, url *url.URL, done chan JobResult, wg *sync.WaitGroup) {
	links := JobResult{FromURL: url, LinksTo: nil}

	resp, err := client.Get(url.String())
	if err != nil {
		// We could implement some retry logic here. I didn't though!
		log.Printf("Loading failed for link %s. Pretending it has no links. Err: %s\n", url.String(), err)
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
				// End of the file, break out of the loop
				done <- links
				wg.Done()
				return
			}
			// There's been an error. We should probably deal with this more
			// gracefully, but for now log and return the links we did get.
			log.Println("There was an error parsing the html.", err)
			done <- links
			return

		case tt == html.StartTagToken:
			t := z.Token()

			if t.Data == "a" {
				// We've found <a>!
				link := getHref(t)
				if link != "" {
					foundLink, err := url.Parse(link)
					if err != nil {
						// Looks like garbage in the href tag. Leave it out.
						log.Printf("Wasn't able to parse %s. Ignoring. Error %s\n", link, err)
					} else {
						links.LinksTo = append(links.LinksTo, foundLink)
					}
				}
			}
		}
	}
}

/*
getHref will when given a html.Token, find the href key and returns it.
If it cannot find a href, it returns the empty string.
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

/*
linkConsumer ranges on a channel, appending the data it gets from it onto job
results array you passed it. Since this is one single goroutine, this is
threadsafe (but kinda cheaty)
*/
func linkConsumer(j chan JobResult, results *[]JobResult, wg *sync.WaitGroup) {
	for s := range j {
		*results = append(*results, s)

	}
	wg.Done()
}
