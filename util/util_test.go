package util

import (
	"net/url"
	"testing"
)

func TestFilterLinksByHostname(t *testing.T) {
	urlSubPage, _ := url.Parse("https://kn100.me/about")
	urlSubDomain, _ := url.Parse("https://hire.kn100.me/")
	otherDomain, _ := url.Parse("https://monzo.com/")
	var urls []*url.URL
	urls = append(urls, urlSubPage, urlSubDomain, otherDomain)
	filteredURLS := FilterLinksByHostname(urls, "https://kn100.me/")
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
	url, _ := url.Parse("https://kn100.me/about")
	if LinkPartOfSite(url, "kn100.me") != true {
		t.Errorf("URL %s is part of %s", url.String(), "kn100.me")
	}
	if LinkPartOfSite(url, "monzo.com") != false {
		t.Errorf("URL %s is NOT part of %s", url.String(), "monzo.com")
	}
}

func TestInvalidDataLinkPartOfSite(t *testing.T) {
	goodURL, _ := url.Parse("http://kn100.me/")
	badURL, _ := url.Parse("https://kn100.man.")
	if LinkPartOfSite(badURL, "kn100.me") != false {
		t.Errorf("Invalid domain %s should NOT be considered part of the site.", badURL.String())
	}
	if LinkPartOfSite(goodURL, "kn100.man") != false {
		t.Errorf("Invalid domain %s should NOT be considered part of the site.", badURL.String())
	}
}
