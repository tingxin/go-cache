package go_cache

import (
	"fmt"
)

func init() {
	KeyNotFoundError = fmt.Errorf("Didn't found key")
	__default = newCache()
}

var __default *cache

func Default() *cache {
	return __default
}

func newCache() *cache {
	ca := &cache{get_requests: make(chan valueRequest), fetcher_requests: make(chan fetcherRequest)}
	go ca.serve()
	return ca
}

type cache struct {
	get_requests     chan valueRequest
	fetcher_requests chan fetcherRequest
}

func (p *cache) Get(key string) (Object, error) {
	response := make(chan cache_item)
	p.get_requests <- valueRequest{key: key, response: response}
	res := <-response
	return res.value, res.err
}

func (p *cache) SetWithFetcher(key string, fetcher CFetcher, args ...Object) {
	p.fetcher_requests <- fetcherRequest{f: fetcher, key: key, fetcherArguments: args}
}

func (p *cache) serve() {
	contents := make(map[string]*cacheItemDetail)
	for {
		select {
		case get_request := <-p.get_requests:
			if detail, ok := contents[get_request.key]; ok {
				if !detail.calculating {
					detail.calculating = true
					go detail.calculateValue()
				}
				go detail.deliverValue(get_request.response)
			} else {
				get_request.response <- cache_item{value: nil, err: KeyNotFoundError}
			}
		case func_request := <-p.fetcher_requests:

			contents[func_request.key] = &cacheItemDetail{
				value:            nil,
				fetcher:          func_request.f,
				fetcherArguments: func_request.fetcherArguments,
				ready:            make(chan struct{}),
				err:              nil,
			}
		}
	}

}
