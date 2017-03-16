# go cache
go-cache is concurrent, no-blocked, functional Cache by go language
 
## Features
* Support lazy getting by fetcher
* Support data expires mechanism, it will use fetcher to refresh data automatically when data expires
* Support key expires mechanism, it will delete expired key automatically
* Support high concurrent

## Quick start
### Cache data
cache you data as follow:

    testKey := "k8s"
    testValue := "kuberneters"
    Cache().Set(testKey, testValue)
    
In most time, your data come from network, db, or you need calculating it while using, so you can set a fetcher to the cache, when you need use the data or the data expired, the go_cache will help you to get the data automatically:

    //Please check the httpGet method in cache_test.go file
    testKeyOnRemote:="k8sOnRemote"
    Cache().SetWithFetcher(testKeyOnRemote, func(arguments ...Object) (content Object, err error) {
        	key, _ := arguments[0].(string)
        	if remoteContent, checker := httpGet(mockService.URL+"/"+key, ""); checker {
        		content = string(googleContent)
        		err = nil
        	} else {
        		err = fmt.Errorf("get error when call %s", key)
        	}
        	return
        }, testKeyOnRemote)
        
You fetcher function should be this type
    
    CFetcher

### Get data
You can get the data like this:

    // because you know your data is string, so you can use GetString method to get your token data, if you use Get method
    // you need do the type assertion
    value, err := go_cache.Cache().GetString(testKey)
    // if you did not set the key for cache, you will got the "key didn't found" error form err 
    
If you cache a Struct data, you need use Get method and do the type assertion:

    type person struct {
    	Name string
    	Age  int
    }
    
    tingxin := person{"tingxin", 30}
    Cache().Set(tingxin.Name, tingxin)
    
    p, err := Cache().Get(tingxin.Name)

    pTingxin, _ := p.(person)
    fmt.Println(pTingxin.Age)
    	
### Set Data Expires  
you  can set the expires for a cache item, if the item have been set fetcher, it will update the data when it is expired, or you will get
an "Data is expired" error when you get it

    Cache().SetValueExpires(tingxin.Name, time.Second*2)
    	
### Set key Expires
if key expires, the cache will delete it automatically
    	
    Cache().SetKeyExpires(tingxin.Name, time.Second*2)
    	
### more example
You can use follow method to quickly get the target type data:

     token, _ := go_cache.Cache().Get("token")
     googleValue, _ := go_cache.Cache().GetString("google")
     intValue, _ := go_cache.Cache().GetInt("intValue")
     floatValue, _ := go_cache.Cache().GetFloat64("floatValue")
     boolValue, _ := go_cache.Cache.GetBool("boolValue")
     
Notice: you always need do the type assertion when use Get method.
If you cache the "token" as string data, but you get it use GetInt or other method, it will return (nil, false)

### You can get more example from cache_test.go file

