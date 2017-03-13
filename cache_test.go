package go_cache

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"testing"
)

var testKeys = [4]string{"beijing", "Big_data", "Cloud", "Meteorology"}
var availableKeywords []string = []string{}

func init() {
	for i := 0; i < len(testKeys); i++ {
		Default().SetWithFetcher(testKeys[i], func(arguments ...Object) (content Object, err error) {
			key, _ := arguments[0].(string)
			if googleContent, checker := httpGet("https://en.wikipedia.org/wiki/"+key, ""); checker {
				content = string(googleContent)
				err = nil
			} else {
				err = fmt.Errorf("get error when call %s", key)
			}
			return
		}, testKeys[i])
	}

}

func TestCache_Get(t *testing.T) {
	runtime.GOMAXPROCS(4)
	for i := 0; i < 100; i++ {
		index := rand.Int() % len(testKeys)
		key := testKeys[index]

		if content, err := Default().Get(key); err == nil {
			if content == nil {
				t.Logf("key %s and get message is : %s", key, "failed")
			} else {
				t.Logf("key %s and get message is : %s", key, "successful")
			}
		} else {
			t.Logf("%v", err)
			t.Fail()
		}
	}
}

func TestCache_GetString(t *testing.T) {
	runtime.GOMAXPROCS(4)
	for i := 0; i < 100; i++ {
		index := rand.Int() % len(testKeys)
		key := testKeys[index]

		if content, err := Default().GetString(key); err == nil {
			if content == "" {
				t.Logf("key %s and get message is : %s", key, "failed")
			} else {
				t.Logf("key %s and get message is : %s", key, content)
			}
		} else {
			t.Logf("%v", err)
			t.Fail()
		}
	}
}

func httpGet(uri string, token string) ([]byte, bool) {
	if req, err := http.NewRequest("GET", uri, nil); err == nil {
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		client := &http.Client{}

		resp, err := client.Do(req)

		if err == nil {
			defer resp.Body.Close()
			if body, err := ioutil.ReadAll(resp.Body); err == nil {
				return body, true
			}
		}
		log.Printf("Failed to fetch data from %s - %v", uri, err)
	}
	return nil, false
}
