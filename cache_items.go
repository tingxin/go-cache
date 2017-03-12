package go_cache

import (
	"time"
)

var KeyNotFoundError error

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

type fetcherRequest struct {
	key              string
	f                CFetcher
	fetcherArguments []Object
}

type cacheItemDetail struct {
	value            Object
	fetcher          CFetcher
	fetcherArguments []Object
	valueTime        time.Time
	expireSeconds    float32
	valueType        int
	storedPath       string
	frequency        int
	err              error
	ready            chan struct{}
	calculating      bool
}

func (p *cacheItemDetail) calculateValue() {
	value, err := p.fetcher(p.fetcherArguments...)
	if err == nil {
		p.value = value
	} else {
		p.err = err
	}
	close(p.ready)
}

func (p *cacheItemDetail) deliverValue(response chan<- cache_item) {
	<-p.ready
	response <- cache_item{value: p.value, err: p.err}
}
