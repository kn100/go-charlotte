package sitemap

import (
	"fmt"
	"net/url"
	"strings"
)

// Node stores metadata about a given link as well as a slice pointing to SiteMapNodes that it links to.
type Node struct {
	URL       *url.URL `json:"url"`
	CreatedAt int64    `json:"createdat"`
	LinksTo   []*Node  `json:"linksto"`
}

// String() returns a nice readable representation of this node and all it links to recursively.
// It hides the initial depth value used for recursion from the consumer. Just a bit nicer to work with!
func (s *Node) String() string {
	return s.string(0)
}

func (s *Node) string(depth int) string {
	output := ""
	output = output + fmt.Sprintf("%s%s\n", s.indent(depth), s.URL)
	depth++
	for _, node := range s.LinksTo {
		output = output + node.string(depth)
	}
	return output
}

// AddLeaf adds a leaf to this node (ie, a link that can be traversed from this node)
func (s *Node) AddLeaf(siteMapNode *Node) {
	s.LinksTo = append(s.LinksTo, siteMapNode)
}

func (s *Node) indent(depth int) string {
	return strings.Repeat(" ", depth*IndentSpaces)
}
