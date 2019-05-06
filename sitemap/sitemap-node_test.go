package sitemap

import (
	"fmt"
	"net/url"
	"testing"
)

func TestInvalidDataLinkPartOfSite(t *testing.T) {
	goodURL, _ := url.Parse("http://kn100.me/")
	badURL, err := url.Parse("https://kn100.man.")
	if err != nil {
		fmt.Println(err)
	}
	if LinkPartOfSite(badURL, goodURL) != false {
		t.Errorf("Invalid domain %s should NOT be considered part of the site.", badURL.String())
	}
	if LinkPartOfSite(goodURL, badURL) != false {
		t.Errorf("Invalid domain %s should NOT be considered part of the site.", badURL.String())
	}
}
