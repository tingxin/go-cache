package go_cache

import (
	"time"
)

type Object interface {
}

type CFetcher func(arguments ...Object) (value Object, err error)

type cache_item struct {
	value Object
	err   error
}

type valueRequest struct {
	key      string
	response chan<- cache_item // the client wants a single result
}

type setRequest struct {
	key   string
	value Object
}

type expiresRequest struct {
	key      string
	expires  time.Duration
	response chan<- error
}

type fetcherRequest struct {
	key              string
	f                CFetcher
	fetcherArguments []Object
}

type cacheItem struct {
	value         Object
	expiresTag    *expiresExtra
	keyExpiresTag *expiresExtra
	fetcherTag    *fetcherExtra
}

type expiresExtra struct {
	valueTime time.Time
	expires   time.Duration
}

type fetcherExtra struct {
	fetcher          CFetcher
	fetcherArguments []Object
	err              error
	ready            chan struct{}
	calculated       bool
}

func (p *cacheItem) calculateValue() {
	value, err := p.fetcherTag.fetcher(p.fetcherTag.fetcherArguments...)
	if err == nil {
		p.value = value
		if p.expiresTag !=nil{
			p.expiresTag.valueTime = time.Now()
		}
	} else {
		p.fetcherTag.err = err
	}
	close(p.fetcherTag.ready)
}

func (p *cacheItem) deliverValue(response chan<- cache_item) {
	<-p.fetcherTag.ready
	response <- cache_item{value: p.value, err: p.fetcherTag.err}
}

func (p *expiresExtra) isExpires() (expires bool, dur time.Duration) {
	timeExpires := time.Now().Sub(p.valueTime)
	if timeExpires > p.expires {
		expires = true
	}
	dur = timeExpires - p.expires
	return
}
