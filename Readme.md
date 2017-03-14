# go cache
* go-cache is concurrent, no-blocked, functional Cache by go language
 
* go-cache use the fetcher get data automatically, user don't need take care the data when the data expired or invalid

## Quick start
cache you data as follow:

    testKey := "k8s"
    testValue := "kuberneters"

    Cache().Set(testKey, testValue)

When you need use it, you can get the data like this:

    // because you know you token is string data, so you can use GetString method to get your token data, if you use Get method
    // you need do the type assertion
    
    if value, ok := go_cache.Cache().GetString(testKey); ok{
    	fmt.Print(value)
    }else {
    	fmt.Print("can not find the key")
    }
    
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
    Cache().SetKeyExpires(tingxin.Name, time.Second*2)
    time.Sleep(time.Second*3)
    _,dErr:=Cache().Get(tingxin.Name)
    
    if dErr==nil{
    	t.Errorf("key %s should be deleted", tingxin.Name)
    }
    	
In most time, your data come from network or db, you can set a fetcher to the cache, when you need use the data or the data expired, the go_cache will help you 
to get the data automatically:
    
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
    fmt.Println(v)

You can use follow method to quickly get the target type data:

     token, _ := go_cache.Cache().Get("token")
     googleValue, _ := go_cache.Cache().GetString("google")
     intValue, _ := go_cache.Cache().GetInt("intValue")
     floatValue, _ := go_cache.Cache().GetFloat64("floatValue")
     boolValue, _ := go_cache.Cache.GetBool("boolValue")
     
Notice: you always need do the type assertion when use Get method.
If you cache the "token" as string data, but you get it use GetInt or other method, it will return (nil, false)

