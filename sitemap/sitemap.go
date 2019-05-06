package sitemap

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/kn100/charlotte/fetch"
	"github.com/kn100/charlotte/util"
)

// IndentSpaces controls the number of spaces to indent string output
const IndentSpaces int = 2

// SitemapURLSIndexed is a map where the key is a URL, and the value is a pointer to its respective Node.
// It is here as an optimization to inserting into the Sitemap, avoiding having to do a search every time
// we want to append to the tree.
var SitemapURLSIndexed = make(map[string]*Node)

// SiteMap stores metadata about a sitemap as well a pointer to the root SiteMapNode
type SiteMap struct {
	RootNode   *Node `json:"rootnode"`
	Depth      int   `json:"depth"`
	CreatedAt  int64 `json:"createdat"`
	FinishedAt int64 `json:"finishedat"`
}

// String returns a human readable representation of the Sitemap.
func (s *SiteMap) String() string {
	return s.RootNode.String()
}

// JSON returns a JSON representation of the Sitemap.
func (s *SiteMap) JSON() string {
	b, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		log.Printf("Unable to marshal Sitemap into JSON. Error %s", err)
		return "{}"
	}
	return string(b)
}

// AddLeaf adds a leaf to this node (ie, a link that can be traversed from this node)
// Returns false if node already existed (ie, this branch has been traversed)
// Returns true if new node
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
		// We've already got this in the sitemap. Ignore.
		// Will be better to add this to the sitemap too but it causes an infinite loop.
		// Didn't have time to investigate.
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
	SitemapURLSIndexed[baseURL.String()] = &rootNode
	return true
}

/*
GetNodesFromDepth returns nodes at a given depth in the tree. For example:
  A-B-C
  	|	  If your sitemap had these nodes, and you asked it for nodes at
    D     depth 2, the return would be nodes C and D. If you asked it for
    |     nodes at depth 3, the return would be E.
    E
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
MakeSiteMap returns a sitemap, indexed from the seed up to the depth specified.
*/
func MakeSiteMap(seed string, depth int) *SiteMap {
	sm := SiteMap{}
	sm.CreatedAt = time.Now().Unix()
	sm.Depth = depth
	seedurl, err := url.Parse(seed)
	if err != nil {
		log.Printf("The seed URL (%s) didn't parse. Error was %s\n", seed, err)
		return &sm
	}

	sm.SetRootNode(seedurl)
	fillSiteMap(&sm)
	return &sm
}

func fillSiteMap(sm *SiteMap) {
	for currDepth := 0; currDepth < sm.Depth; currDepth++ {
		nodes := sm.GetNodesFromDepth(currDepth)
		uris := getURLsFromNodeSlice(nodes)
		jobResults := fetch.Links(uris)
		for i := 0; i < len(jobResults); i++ {
			util.CleanURLS(jobResults[i].LinksTo)
			jobResults[i].LinksTo = util.FilterLinksByHostname(jobResults[i].LinksTo, sm.RootNode.URL)
		}
		addToSiteMap(sm, jobResults)
	}
	sm.FinishedAt = time.Now().Unix()
}

func getURLsFromNodeSlice(nodes []*Node) []*url.URL {
	var urls []*url.URL
	for i := 0; i < len(nodes); i++ {
		urls = append(urls, nodes[i].URL)
	}
	return urls
}

func addToSiteMap(sitemap *SiteMap, jobResults []fetch.JobResult) {
	for i := 0; i < len(jobResults); i++ {
		fromNode := jobResults[i].FromURL
		for j := 0; j < len(jobResults[i].LinksTo); j++ {
			sitemap.AddLeaf(fromNode, jobResults[i].LinksTo[j])
		}
	}
}
