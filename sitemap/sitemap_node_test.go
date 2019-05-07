package sitemap

import (
	"net/url"
	"testing"
)

func TestString(t *testing.T) {
	baseURL, _ := url.Parse("https://kn100.me/")
	leafURL, _ := url.Parse("https://kn100.me/about")
	leafURL2, _ := url.Parse("https://kn100.me/about/kevin")
	leafURL3, _ := url.Parse("https://kn100.me/test")
	root := Node{URL: baseURL}
	leaf := Node{URL: leafURL}
	leaf2 := Node{URL: leafURL2}
	leaf3 := Node{URL: leafURL3}
	root.AddLeaf(&leaf)
	leaf.AddLeaf(&leaf2)
	root.AddLeaf(&leaf3)

	// There's gotta be a better way of doing multiline strings!
	expected := `https://kn100.me/
  https://kn100.me/about
    https://kn100.me/about/kevin
  https://kn100.me/test
`
	actual := root.String()
	if actual != expected {
		t.Errorf("The string output did not match what was expected.\n Expected: \n %s\n Actual:\n %s\n", expected, actual)
	}

}
