package go_cache

import (
	"fmt"
	"sync"
	"time"
)

func init() {
	__default = newCache()
}

var __default *cache

func Cache() *cache {
	return __default
}

func newCache() *cache {
	ca := &cache{}
	ca.get_requests = make(chan valueRequest)
	ca.set_requests = make(chan setRequest)
	ca.fetcher_requests = make(chan fetcherRequest)
	ca.expires_requests = make(chan expiresRequest)
	ca.deleteKey_requests = make(chan string, 5)

	ca.keyExpiresMap = make(map[string]*expiresExtra)

	ca.DefaultExpired = time.Hour
	go ca.requestForward()
	go ca.keyExpiresCheck()
	return ca
}

type cache struct {
	get_requests       chan valueRequest
	set_requests       chan setRequest
	expires_requests   chan expiresRequest
	fetcher_requests   chan fetcherRequest

	deleteKey_requests chan string

	keyExpiresMap   map[string]*expiresExtra
	keyExpiresMutex sync.Mutex
	DefaultExpired  time.Duration
}

func (p *cache) Get(key string) (Object, error) {
	response := make(chan cache_item)
	p.get_requests <- valueRequest{key: key, response: response}
	res := <-response
	return res.value, res.err
}

func (sender *cache) GetString(key string) (result string, err error) {
	object_value, get_err := sender.Get(key)
	err = get_err
	if err == nil {
		if typedValue, isTypeValue := object_value.(string); isTypeValue {
			result = typedValue
		}
	}
	return
}

func (sender *cache) GetInt(key string) (result int, err error) {
	object_value, get_err := sender.Get(key)
	err = get_err
	if err == nil {
		if typedValue, isTypeValue := object_value.(int); isTypeValue {
			result = typedValue
		}
	}
	return
}

func (p *cache) GetFloat64(key string) (result float64, err error) {
	object_value, get_err := p.Get(key)
	err = get_err
	if err == nil {
		if typedValue, isTypeValue := object_value.(float64); isTypeValue {
			result = typedValue
		}
	}
	return
}

func (p *cache) GetBool(key string) (result bool, err error) {
	object_value, get_err := p.Get(key)
	err = get_err
	if err == nil {
		if typedValue, isTypeValue := object_value.(bool); isTypeValue {
			result = typedValue
		}
	}
	return
}

func (p *cache) Set(key string, value Object) {
	p.set_requests <- setRequest{key: key, value: value}
}

func (p *cache) SetKeyExpires(key string, expires time.Duration) {
	p.keyExpiresMutex.Lock()
	defer p.keyExpiresMutex.Unlock()
	p.keyExpiresMap[key] = &expiresExtra{valueTime: time.Now(), expires: expires}
}

func (p *cache) SetValueExpires(key string, expires time.Duration) error {
	res := make(chan error)
	p.expires_requests <- expiresRequest{key: key, expires: expires, response: res}
	return <-res
}

func (p *cache) SetWithFetcher(key string, fetcher CFetcher, args ...Object) {
	p.fetcher_requests <- fetcherRequest{f: fetcher, key: key, fetcherArguments: args}
}

func (p *cache) Delete(key string) {
	p.deleteKey_requests <- key
	p.keyExpiresMutex.Lock()
	defer p.keyExpiresMutex.Unlock()
	_, ok := p.keyExpiresMap[key]
	if ok {
		delete(p.keyExpiresMap, key)
	}
}

func (p *cache) keyExpiresCheck() {

	for {
		time.Sleep(time.Second)
		p.keyExpiresMutex.Lock()
		for key, value := range p.keyExpiresMap {
			if ok,_:=value.isExpires();ok {
				p.deleteKey_requests <- key
				delete(p.keyExpiresMap, key)
			}
		}
		p.keyExpiresMutex.Unlock()
	}
}

func (p *cache) requestForward() {
	contents := make(map[string]*cacheItem)
	for {
		select {
		case get_request := <-p.get_requests:
			getRequestHandler(&get_request, contents)
		case set_request := <-p.set_requests:
			setRequestHandler(&set_request, contents)
		case fetcher_request := <-p.fetcher_requests:
			fetcherRequestHandler(&fetcher_request, contents)
		case expires_request := <-p.expires_requests:
			expiresRequestHandler(&expires_request, contents)
		case deleteKey_request := <-p.deleteKey_requests:
			 deleteKeyRequestHandler(deleteKey_request, contents)
		}
	}
}

func getRequestHandler(get_request *valueRequest, contents map[string]*cacheItem) {
	detail, ok := contents[get_request.key]
	if !ok {
		get_request.response <- cache_item{value: nil, err: fmt.Errorf("Didn't found the key %s", get_request.key)}
		return
	}

	expires := false
	if detail.expiresTag != nil {
		expires, _ = detail.expiresTag.isExpires()
	}

	if detail.fetcherTag == nil {
		if expires {
			get_request.response <- cache_item{value: nil, err: fmt.Errorf("%s value exipred\n", get_request.key)}
		} else {
			get_request.response <- cache_item{value: detail.value, err: nil}
		}
		return
	}

	if expires {
		detail.fetcherTag.calculated = false
		detail.fetcherTag.ready = make(chan struct{})
	}

	if !detail.fetcherTag.calculated {
		detail.fetcherTag.calculated = true
		go detail.calculateValue()
	}
	go detail.deliverValue(get_request.response)
}

func setRequestHandler(set_request *setRequest, contents map[string]*cacheItem) {
	if detail, ok := contents[set_request.key]; ok {
		detail.value = set_request.value
	} else {

		ca := &cacheItem{
			value:      set_request.value,
			fetcherTag: nil,
			expiresTag: nil,
		}
		contents[set_request.key] = ca
	}
}

func fetcherRequestHandler(fetcher_request *fetcherRequest, contents map[string]*cacheItem) {
	fetcherTag := &fetcherExtra{
		fetcher:          fetcher_request.f,
		fetcherArguments: fetcher_request.fetcherArguments,
		ready:            make(chan struct{}),
		err:              nil,
	}
	contents[fetcher_request.key] = &cacheItem{
		value:      nil,
		fetcherTag: fetcherTag,
	}
}

func expiresRequestHandler(expires_request *expiresRequest, contents map[string]*cacheItem) {
	detail, ok := contents[expires_request.key]
	if !ok {
		expires_request.response <- fmt.Errorf("Didn't found the key %s", expires_request.key)
		return
	}
	if detail.expiresTag == nil {
		detail.expiresTag = &expiresExtra{}
	}
	detail.expiresTag.expires = expires_request.expires
	detail.expiresTag.valueTime = time.Now()
	expires_request.response <- nil
}

func deleteKeyRequestHandler(key string, contents map[string]*cacheItem) {
	_, ok := contents[key]
	if ok {
		delete(contents, key)
	}

}
