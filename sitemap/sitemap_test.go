package sitemap

import (
	"net/url"
	"testing"

	"github.com/kn100/charlotte/fetch"
)

func TestSiteMapString(t *testing.T) {
	baseURL, _ := url.Parse("https://kn100.me/")
	root := Node{URL: baseURL}

	// There's gotta be a better way of doing multiline strings!
	expected := `https://kn100.me/
`
	actual := root.String()
	if actual != expected {
		t.Errorf("The string output did not match what was expected.\n Expected: \n %s\n Actual:\n %s\n", expected, actual)
	}
}

func TestJSON(t *testing.T) {
	baseURL, _ := url.Parse("https://kn100.me/")
	root := Node{URL: baseURL}
	sm := SiteMap{RootNode: &root, Depth: 0, CreatedAt: 31989300, FinishedAt: 31989300}
	expected := `{"RootNode":{"URL":{"Scheme":"https","Opaque":"","User":null,"Host":"kn100.me","Path":"/","RawPath":"","ForceQuery":false,"RawQuery":"","Fragment":""},"CreatedAt":0,"LinksTo":null},"EffectiveTldPlusOne":"","Depth":0,"CreatedAt":31989300,"FinishedAt":31989300}`
	actual := sm.JSON()
	if actual != expected {
		t.Errorf("The JSON output did not match what was expected.\n Expected: \n %s\n Actual:\n %s\n", expected, actual)
	}
}

func TestAddLeafValidUnseen(t *testing.T) {
	baseURL, _ := url.Parse("https://kn100.me/")
	leafURL, _ := url.Parse("https://kn100.me/leaf")
	sm := SiteMap{RootNode: nil, Depth: 0, CreatedAt: 31989300, FinishedAt: 31989300}
	sm.SetRootNode(baseURL)
	added, err := sm.AddLeaf(baseURL, leafURL)
	if err != nil {
		t.Errorf("No error should have occured. Err: %s", err)
	}
	if added != true {
		t.Errorf("Should have added %s", leafURL.String())
	}
}

func TestAddLeafValidSeen(t *testing.T) {
	baseURL, _ := url.Parse("https://kn100.me/")
	dupeURL, _ := url.Parse("https://kn100.me/")
	sm := SiteMap{RootNode: nil, Depth: 0, CreatedAt: 31989300, FinishedAt: 31989300}
	sm.SetRootNode(baseURL)
	added, err := sm.AddLeaf(baseURL, dupeURL)
	if err != nil {
		t.Errorf("No error should have occured. Err: %s", err)
	}
	if added == true {
		t.Errorf("Should NOT have added %s", dupeURL.String())
	}
}

func TestAddLeafNoRootNode(t *testing.T) {
	baseURL, _ := url.Parse("https://kn100.me/")
	leafURL, _ := url.Parse("https://kn100.me/leaf")
	sm := SiteMap{RootNode: nil, Depth: 0, CreatedAt: 31989300, FinishedAt: 31989300}
	added, err := sm.AddLeaf(baseURL, leafURL)
	if err == nil {
		t.Errorf("Error should have occured, since there was no root node.")
	}
	if added == true {
		t.Errorf("Should NOT have added %s", leafURL.String())
	}
}

func TestAddLeafNotSeenFromNode(t *testing.T) {
	baseURL, _ := url.Parse("https://kn100.me/")
	fromURL, _ := url.Parse("https://monzo.com/")
	toURL, _ := url.Parse("https://monzo.com/about")
	sm := SiteMap{RootNode: nil, Depth: 0, CreatedAt: 31989300, FinishedAt: 31989300}
	sm.SetRootNode(baseURL)
	added, err := sm.AddLeaf(fromURL, toURL)
	if err == nil {
		t.Errorf("Error should have occured, since there the from node was not in the tree.")
	}
	if added == true {
		t.Errorf("Should NOT have added %s", toURL.String())
	}
}

func TestAddRootNode(t *testing.T) {
	baseURL, _ := url.Parse("https://kn100.me/")
	sm := SiteMap{RootNode: nil, Depth: 0, CreatedAt: 31989300, FinishedAt: 31989300}
	res := sm.SetRootNode(baseURL)
	if res != true {
		t.Errorf("Should have returned true as was valid root node")
	}
	if baseURL != sm.RootNode.URL {
		t.Errorf("Should have added %s as root node", baseURL.String())
	}
}

func TestAddRootNodeTwice(t *testing.T) {
	baseURL, _ := url.Parse("https://kn100.me/")
	baseURL2, _ := url.Parse("https://monzo.com/")
	sm := SiteMap{RootNode: nil, Depth: 0, CreatedAt: 31989300, FinishedAt: 31989300}
	sm.SetRootNode(baseURL)
	res := sm.SetRootNode(baseURL2)
	if res != false {
		t.Errorf("Should have returned false as root node was already set once.")
	}
}

func TestGetNodesFromDepth0(t *testing.T) {
	baseURL, _ := url.Parse("https://kn100.me/")
	sm := SiteMap{RootNode: nil, Depth: 0, CreatedAt: 31989300, FinishedAt: 31989300}
	sm.SetRootNode(baseURL)
	res := sm.GetNodesFromDepth(0)
	if res[0].URL.String() != "https://kn100.me/" || len(res) != 1 {
		t.Errorf("Should have gotten 1 node")
	}
}

func TestGetNodesFromDepth1(t *testing.T) {
	baseURL, _ := url.Parse("https://kn100.me/")
	leafURL, _ := url.Parse("https://kn100.me/leaf/")
	leafURL2, _ := url.Parse("https://kn100.me/otherleaf/")

	sm := SiteMap{RootNode: nil, Depth: 2, CreatedAt: 31989300, FinishedAt: 31989300}
	sm.SetRootNode(baseURL)
	sm.AddLeaf(baseURL, leafURL)
	sm.AddLeaf(baseURL, leafURL2)
	res := sm.GetNodesFromDepth(1)
	if len(res) != 2 {
		t.Errorf("Should have gotten 2 nodes. Got %d", len(res))
	}
}

// TODO: Testing fillSiteMap and MakeSiteMap would require mocking, and I don't have time right now
// to do this.

func TestGetURLsFromNodeSlice(t *testing.T) {
	URL0, _ := url.Parse("https://kn100.me/")
	URL1, _ := url.Parse("https://kn100.me/leaf/")
	node1 := Node{URL: URL0}
	node2 := Node{URL: URL1}
	var nodes []*Node
	nodes = append(nodes, &node1, &node2)
	urls := getURLsFromNodeSlice(nodes)
	if urls[0] != URL0 {
		t.Errorf("urls[0] was supposed to be %s", URL0)
	}
	if urls[1] != URL1 {
		t.Errorf("urls[1] was supposed to be %s", URL1)
	}
}

func TestaddToSiteMap(t *testing.T) {
	sm := SiteMap{RootNode: nil, Depth: 2, CreatedAt: 31989300, FinishedAt: 31989300}
	baseURL, _ := url.Parse("https://kn100.me/")
	sm.SetRootNode(baseURL)

	leafURL, _ := url.Parse("https://kn100.me/leaf/")
	leafURL2, _ := url.Parse("https://kn100.me/otherleaf/")
	var leafs []*url.URL
	leafs = append(leafs, leafURL, leafURL2)

	jobResult := fetch.JobResult{FromURL: baseURL, LinksTo: leafs}
	var jobResults []fetch.JobResult
	jobResults = append(jobResults, jobResult)

	addToSiteMap(&sm, jobResults)
	if sm.RootNode.LinksTo[0].URL != leafURL {
		t.Errorf("Expected the first node the root node linked to be %s, actual: %s", leafURL, sm.RootNode.LinksTo[0].URL)
	}
	if sm.RootNode.LinksTo[1].URL != leafURL {
		t.Errorf("Expected the second node the root node linked to be %s, actual: %s", leafURL2, sm.RootNode.LinksTo[1].URL)
	}
}
