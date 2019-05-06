package util

import (
	"fmt"
	"net/url"

	"golang.org/x/net/publicsuffix"
)

/*
FilterLinksByHostname removes any links from a list that are not part of a site.
For example, if the list contained kn100.me, hello.kn100.me/hi, and monzo.com,
the returned list would be kn100.me and hello.kn100.me/hi. All links passed in
are expected to be absolute. You can use MakeLinkAbsolute(linkURL, baseURL) to
ensure this.
*/
func FilterLinksByHostname(links []*url.URL, root *url.URL) []*url.URL {
	var acceptableLinks []*url.URL
	for i := 0; i < len(links); i++ {
		if LinkPartOfSite(links[i], root) {
			acceptableLinks = append(acceptableLinks, links[i])
		}

	}
	return acceptableLinks
}

/*
CleanURLs will process a list of pointers to url.URLs, removing anchors and
query parameters.
*/
func CleanURLS(links []*url.URL) {
	for i := 0; i < len(links); i++ {
		CleanURL(links[i])
	}
}

/*
LinkPartOfSite checks whether a given link exists under the domain of a site. It
returns true if so, false if not. It makes use of the publicsuffix library for
this, which is a database of tlds. It's possible this could be out of date if
you're having trouble here.
*/
func LinkPartOfSite(link, root *url.URL) bool {
	// This allows us to crawl subdomains too!
	linkTLDPlusOne, err := publicsuffix.EffectiveTLDPlusOne(link.Hostname())
	if err != nil {
		fmt.Sprintln("I couldn't extract the TLD Plus One of #s. This won't be included in the sitemap.", link.Hostname())
		return false
	}
	rootTLDPlusOne, err := publicsuffix.EffectiveTLDPlusOne(root.Hostname())
	if err != nil {
		fmt.Sprintln("I couldn't extract the TLD Plus One of #s. This won't be included in the sitemap.", root.Hostname())
		return false
	}
	return rootTLDPlusOne == linkTLDPlusOne
}

/*
CleanURLs will process a pointer to a url.URL, removing anchors and
query parameters.
*/
func CleanURL(link *url.URL) {
	link.Fragment = ""
	link.RawQuery = ""
}
