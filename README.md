# go-charlotte

A Creepy crawly crawler

go-charlotte is a web crawler that can spider a given domain to a given depth limit. It does this relatively quickly. See main.go for an example of how to use it.

The sitemap is a unique list of domains. They are shown in first seen order. This means at each depth level, the URLs seen are 'new' and have never been seen before. 

Output is available as a nice human readable tree, or nice machine readable JSON.

## Important notes:
* It pays no attention to silly things like server load/politeness. It will blast a lot of requests very quickly.
* It does not pay attention to robots.txt. Only use it on consenting domains! 
* It will traverse to subdomains.

## Running
You can run this like this 
``` 
go run main.go 
```
## Crawl strategy

I optimized for crawl speed more than anything. Every frontier (the current depth of the crawl) it will asynchronously request all the links at that frontier, parse out and filter the links, and thus making the queue for the next frontier. This does have the downside of blasting the server with possibly hundreds of requests very quickly. A naive solution to this would be to add a small delay between each request being fired.

## To implement:
* Finish tests (the remaining stuff to be tested required Mocking, and I ran out of the time I allocated towards this task).
* Make it care about robots.txt conditionally.
* Add some way of throttling the crawler (worker pool?)
* Handle http/https more nicely
* Represent cycles in the tree somehow
* Add useragent so server knows Charlotte is taking a look
* Add retry logic so that if a link fails to load, we don't just throw it away immediately.
* Should probably vendor the deps