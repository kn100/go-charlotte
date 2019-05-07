// Package sitemap provides a method of building and storing a sitemap.
package sitemap

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/kn100/charlotte/fetch"
	"github.com/kn100/charlotte/util"
	"golang.org/x/net/publicsuffix"
)

/*
IndentSpaces controls the number of spaces to indent string output
*/
const IndentSpaces int = 2

/*
SitemapURLSIndexed is a map where the key is a URL, and the value is a pointer
to its respective Node. It is here as an optimization to inserting into the
Sitemap, avoiding having to do a search every time we want to append to the
tree.
*/
var SitemapURLSIndexed = make(map[string]*Node)

/*
SiteMap stores metadata about a sitemap as well a pointer to the root Node.
We can then traverse the entire tree from this one Node.
RootEffectiveTLDPlusOne stores the tld, plus the part to the left of the dot.
For example, blog.monzo.com's RootEffectiveTLDPlusOne becomes monzo.com
*/
type SiteMap struct {
	RootNode                *Node  `json:"RootNode"`
	RootEffectiveTLDPlusOne string `json:"EffectiveTldPlusOne"`
	Depth                   int    `json:"Depth"`
	CreatedAt               int64  `json:"CreatedAt"`
	FinishedAt              int64  `json:"FinishedAt"`
}

/*
SetRootNode sets the root node of this Sitemap.
*/
func (s *SiteMap) SetRootNode(baseURL *url.URL) bool {
	if s.RootNode != nil {
		return false
	}
	rootNode := Node{
		URL:       baseURL,
		CreatedAt: time.Now().Unix(),
	}
	s.RootNode = &rootNode

	rootTLDPlusOne, err := publicsuffix.EffectiveTLDPlusOne(baseURL.Hostname())
	if err != nil {
		fmt.Printf("I couldn't extract the TLD Plus One of baseURL %s. This means the sitemap won't be populated at all.\n", baseURL.String())
		return false
	}
	s.RootEffectiveTLDPlusOne = rootTLDPlusOne
	SitemapURLSIndexed[baseURL.String()] = &rootNode
	return true
}

/*
MakeSiteMap returns a sitemap, indexed from the seed up to the depth specified.
*/
func MakeSiteMap(seed string, depth int, httpTimeout time.Duration) *SiteMap {
	sm := SiteMap{}
	sm.CreatedAt = time.Now().Unix()
	sm.Depth = depth
	seedurl, err := url.Parse(seed)
	if err != nil {
		log.Printf("The seed URL (%s) didn't parse. Error was %s\n", seed, err)
		return &sm
	}

	sm.SetRootNode(seedurl)
	fillSiteMap(&sm, httpTimeout)
	return &sm
}

/*
String returns a human readable representation of the Sitemap.
*/
func (s *SiteMap) String() string {
	return s.RootNode.String()
}

/*
JSON returns a JSON representation of the Sitemap.
*/
func (s *SiteMap) JSON() string {
	b, err := json.Marshal(s)
	if err != nil {
		log.Printf("Unable to marshal Sitemap into JSON. Error %s", err)
		return "{}"
	}
	return string(b)
}

/*
AddLeaf adds a leaf to this node (ie, a link that is traversable from this node)
Returns false if it didn't add the node. If there is no error, then the node
already existed (ie, this branch has been traversed), so it wasn't added.
Returns true if new node
*/
func (s *SiteMap) AddLeaf(from *url.URL, to *url.URL) (bool, error) {
	if to.Hostname() == "" {
		to = s.RootNode.URL.ResolveReference(to)
	}
	if s.RootNode == nil {
		return false, errors.New("there was no root node set")
	}

	fromNode, seenFromURLBefore := SitemapURLSIndexed[from.String()]
	if !seenFromURLBefore {
		errText := fmt.Sprintf("from node %s is not in sitemap", from.String())
		return false, errors.New(errText)
	}
	_, seenToURLBefore := SitemapURLSIndexed[to.String()]
	if seenToURLBefore {
		// We've already got this in the sitemap. Ignore. Will be better to add
		// this to the sitemap too but it causes an infinite loop in the
		// traversal methods. Didn't have time to investigate.
		// TODO: Investigate solutions to this behaviour
		return false, nil
	}

	// Fresh, unseen URL. Create the Node and add it to the sitemap
	newNode := Node{
		URL:       to,
		CreatedAt: time.Now().Unix(),
	}
	fromNode.AddLeaf(&newNode)
	SitemapURLSIndexed[to.String()] = &newNode
	return true, nil
}

/*
GetNodesFromDepth returns nodes at a given depth in the tree.
*/
func (s *SiteMap) GetNodesFromDepth(depth int) []*Node {
	return getNodesFromDepth(*s.RootNode, 0, depth)
}

/*
getNodesFromDepth is a recursive function that identifies nodes at a given
depth. It does this by traversing the tree in a fan out, BFS style.
Usage: getNodesFromDepth(startNode, 0, depth)
*/
func getNodesFromDepth(startNode Node, currDepth int, depth int) []*Node {
	var nodesFound []*Node

	if currDepth == depth {
		var arr []*Node
		arr = append(arr, &startNode)
		return arr
	}
	if currDepth < depth {
		for i := 0; i < len(startNode.LinksTo); i++ {

			nodes := getNodesFromDepth(*startNode.LinksTo[i], currDepth+1, depth)

			nodesFound = append(nodesFound, nodes...)
		}
	}
	return nodesFound
}

/*
fillSiteMap traverses and fills in a given Sitemap.
*/
func fillSiteMap(sm *SiteMap, httpTimeout time.Duration) {
	client := http.Client{
		Timeout: httpTimeout,
	}
	checkDepth := 0
	for checkDepth < sm.Depth {
		nodes := sm.GetNodesFromDepth(checkDepth)
		uris := getURLsFromNodeSlice(nodes)
		jobResults := fetch.Links(&client, uris)
		for i := 0; i < len(jobResults); i++ {
			util.CleanURLS(jobResults[i].LinksTo)
			jobResults[i].LinksTo = util.FilterLinksByHostname(jobResults[i].LinksTo, sm.RootEffectiveTLDPlusOne)
		}
		seenSomethingNew := addToSiteMap(sm, jobResults)
		if !seenSomethingNew {
			break
		}
		checkDepth++
	}
	sm.FinishedAt = time.Now().Unix()
	sm.Depth = checkDepth
}

/*
addToSiteMap takes a list of JobResults, and parses through them to add new
links to the sitemap. Returns true if it added something, false if it did not
(because there were no new links to add)
*/
func addToSiteMap(sitemap *SiteMap, jobResults []fetch.JobResult) bool {
	seenSomethingNew := false
	for i := 0; i < len(jobResults); i++ {
		fromNode := jobResults[i].FromURL
		for j := 0; j < len(jobResults[i].LinksTo); j++ {
			added, err := sitemap.AddLeaf(fromNode, jobResults[i].LinksTo[j])
			if err != nil {
				log.Printf("error adding entry %s from %s to Sitemap, err: %s", jobResults[i].LinksTo[j].String(), fromNode.String(), err)
			}
			if added == true {
				seenSomethingNew = true
			}
		}
	}
	return seenSomethingNew
}

/*
getURLSFromNodeSlice takes a slice of Nodes, and extracts out the URL fields. It
then returns a slice of these URLS
*/
func getURLsFromNodeSlice(nodes []*Node) []*url.URL {
	var urls []*url.URL
	for i := 0; i < len(nodes); i++ {
		urls = append(urls, nodes[i].URL)
	}
	return urls
}
