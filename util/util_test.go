package util

import (
	"fmt"
	"net/url"
	"testing"
)

func TestFilterLinksByHostname(t *testing.T) {
	baseURL, _ := url.Parse("https://kn100.me/")
	urlSubPage, _ := url.Parse("https://kn100.me/about")
	urlSubDomain, _ := url.Parse("https://hire.kn100.me/")
	otherDomain, _ := url.Parse("https://monzo.com/")
	var urls []*url.URL
	urls = append(urls, urlSubPage, urlSubDomain, otherDomain)
	filteredURLS := FilterLinksByHostname(urls, baseURL)
	for i := 0; i < len(filteredURLS); i++ {
		if filteredURLS[i] == otherDomain {
			t.Errorf("%s was not filtered from resultant addresses", otherDomain.String())
		}
	}
}

func TestCleanURLS(t *testing.T) {
	queryAndAnchor, _ := url.Parse("https://kn100.me/test#hello?test=monzo")
	expectedResult := "https://kn100.me/test"
	var urls []*url.URL
	urls = append(urls, queryAndAnchor)
	CleanURLS(urls)
	if urls[0].String() != expectedResult {
		t.Errorf("URL wasn't filtered correctly. Expected %s, Got: %s", expectedResult, urls[0].String())
	}
}

func TestLinkPartOfSite(t *testing.T) {
	baseURL, _ := url.Parse("https://kn100.me/")
	partOfSiteURL, _ := url.Parse("https://kn100.me/about")
	otherDomain, _ := url.Parse("https://monzo.com/test")
	if LinkPartOfSite(partOfSiteURL, baseURL) != true {
		t.Errorf("URL %s is part of %s", partOfSiteURL.String(), baseURL.String())
	}
	if LinkPartOfSite(partOfSiteURL, otherDomain) != false {
		t.Errorf("URL %s is NOT part of %s", otherDomain.String(), baseURL.String())
	}
}

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
