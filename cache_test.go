package go_cache

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"
	"sync"
)

const testNumber = 100

var testKeys = [4]string{"beijing", "Big_data", "Cloud", "Meteorology"}

type mockHandler struct{}

var mockService *httptest.Server

type person struct {
	Name string
	Age  int
}

func init() {
	runtime.GOMAXPROCS(4)
	mockHandlerInstance := &mockHandler{}
	mockService = httptest.NewServer(mockHandlerInstance)

}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	url := r.URL.RequestURI()

	if strings.HasPrefix(url, "/search") {
		w.WriteHeader(200)
		w.Write([]byte("search"))
	} else if strings.HasPrefix(url, "/google") {
		w.WriteHeader(200)
		w.Write([]byte("google"))
	} else if strings.HasPrefix(url, "/time") {
		w.WriteHeader(200)
		w.Write([]byte(time.Now().Format("2006-01-02 15:04:05")))
	} else {
		w.WriteHeader(400)
	}

}

func TestCache_Int(t *testing.T) {

	for i := 0; i < testNumber; i++ {
		go func(index int) {
			Cache().Set(fmt.Sprintf("TestCache_Int%d", index), index)
		}(i)
	}
	time.Sleep(time.Second)

	for i := 0; i < testNumber; i++ {
		go func(index int) {

			key := fmt.Sprintf("TestCache_Int%d", index)
			result, err := Cache().GetInt(key)
			if err != nil {
				t.Errorf("Get error %v \n", err)
				return
			}
			if index != result {
				t.Errorf("the value of key %s is wrong, expected %d, actual is %d", key, index, result)
				return
			}
		}(i)
	}
}

func TestCache_FloatValue(t *testing.T) {
	for i := 0; i < testNumber; i++ {
		go func(index int) {
			Cache().Set(fmt.Sprintf("TestCache_FloatValue%d", index), math.Sqrt(float64(index)))
		}(i)
	}
	time.Sleep(time.Second)
	for i := 0; i < testNumber; i++ {
		go func(index int) {
			key := fmt.Sprintf("TestCache_FloatValue%d", index)
			result, err := Cache().GetFloat64(key)
			if err != nil {
				t.Errorf("Get error %v \n", err)
			} else {
				expected := math.Sqrt(float64(index))
				if math.Abs(result-expected) > 0.00000001 {
					t.Errorf("the value of key %s is wrong, expected %f, actual is %f", key, expected, result)
				}
			}
		}(i)
	}
}

func TestCache_GetString(t *testing.T) {
	for i := 0; i < len(testKeys); i++ {
		Cache().SetWithFetcher(testKeys[i], func(arguments ...Object) (content Object, err error) {
			key, _ := arguments[0].(string)
			if googleContent, checker := httpGet(mockService.URL+"/search/"+key, ""); checker {
				content = string(googleContent)
				err = nil
			} else {
				err = fmt.Errorf("get error when call %s", key)
			}
			return
		}, testKeys[i])
	}
	for i := 0; i < 100; i++ {
		index := rand.Int() % len(testKeys)
		key := testKeys[index]

		if content, err := Cache().GetString(key); err == nil {
			if content == "" {
				t.Error("can not get correct result")
			}
		} else {
			t.Errorf("%v",err)
		}
	}
}

func TestCache_ValueExpires(t *testing.T) {
	testKey := "k8s"

	testValue := "kuberneters"
	v, err := Cache().Get(testKey)
	if err == nil {
		t.Errorf("shoud get error:not found key")
	}

	Cache().Set(testKey, testValue)
	Cache().SetValueExpires(testKey, time.Second*2)

	v, err = Cache().GetString(testKey)
	if err != nil {
		t.Errorf("get error %v", err)
	}

	if v != testValue {
		t.Error("Can not get same data")
	}

	time.Sleep(time.Second * 2)

	v, err = Cache().GetString(testKey)
	if err == nil {
		t.Error("should get error time exipred")
	}

	testKey = "time"
	Cache().SetWithFetcher(testKey, func(arguments ...Object) (content Object, err error) {
		key, _ := arguments[0].(string)
		if googleContent, checker := httpGet(mockService.URL+"/"+key, ""); checker {
			content = string(googleContent)
			err = nil
		} else {
			err = fmt.Errorf("get error when call %s", key)
		}
		return
	}, testKey)

	Cache().SetValueExpires(testKey, time.Second*5)
	v, err = Cache().GetString(testKey)
	if err != nil {
		t.Errorf("get error %v", err)
	}
	firstValue:=v

	if v!=time.Now().Format("2006-01-02 15:04:05"){
		t.Error("get result is failed")
	}

	time.Sleep(time.Second*2)
	v, err = Cache().GetString(testKey)

	if v!=firstValue{
		t.Error("result shouldn't be changed")
	}

	time.Sleep(time.Second*3)

	v, err = Cache().GetString(testKey)
	if err != nil {
		t.Errorf("get error %v", err)
	}

	if v!=time.Now().Format("2006-01-02 15:04:05"){
		t.Error("value shoud be update")
	}

}

func TestCache_Delete(t *testing.T) {
	for i:=0;i<testNumber;i++{
		Cache().Set("delete" + string(i), i*i)
	}

	waiter:=sync.WaitGroup{}
	for i:=0;i<testNumber;i++{
		waiter.Add(1)
		go func(index int) {
			Cache().Delete("delete"+ string(index))
			waiter.Done()
		}(i)
	}
	waiter.Wait()
	for i:=0;i<testNumber;i++{
		go func(index int) {
			waiter.Add(1)
			_, err:=Cache().Get("delete" + string(index))
			if err==nil{
				t.Errorf("should found No key error")
			}
			waiter.Done()
		}(i)
	}
	waiter.Wait()
}

func TestCache_KeyExpires(t *testing.T) {
	tingxin := person{"tingxin", 30}
	Cache().Set(tingxin.Name, tingxin)

	p, err := Cache().Get(tingxin.Name)
	if err != nil {
		t.Error("get error: %#v\n", err)
	}
	pTingxin, ok := p.(person)
	if !ok {
		t.Error("can not get back the person struct")
	}
	if pTingxin.Age != tingxin.Age {
		t.Error("Can not get same data")
	}
	Cache().SetKeyExpires(tingxin.Name, time.Second*2)
	time.Sleep(time.Second*3)
	_,dErr:=Cache().Get(tingxin.Name)

	if dErr==nil{
		t.Errorf("key %s should be deleted", tingxin.Name)
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
