# gorouter

gorouter 是一个轻便的HTTP API 路由库。

### 项目创建背景
之前一直使用大名鼎鼎的 [httprouter](https://github.com/julienschmidt/httprouter)。但由于我写的RESTful API不规范，导致存在一些路由冲突。例如github上讨论的这个[问题](https://github.com/gin-gonic/gin/issues/388)

```
r.GET("/teachers/list", func (c *gin.Context){})
r.GET("/teachers/:id/profile", func (c *gin.Context){})

Error:
[GIN-debug] GET /teachers/list --> main.func·001 (3 handlers)
[GIN-debug] GET /teachers/:id/profile --> main.func·002 (3 handlers)
panic: wildcard route ':id' conflicts with existing children in path '/teachers/:id/profile'
```
当然我们可以把 ```GET /teachers/list``` 改成 ```GET /teachers```。或者把 ``` GET /teachers/:id/profile ``` 改成 ``` GET /teacher/:id/profile ```。 按照restful风格应该采用第一种，但有时候接口太多或者没有严格按照restful风格风格就会导致路由冲突。所以我就偶尔我就会采用第二种，但第二种又会导致我没办法把两个接口归纳在同一个group。gorouter就是为了解决这个问题。

### 用法

```

func main()  {

	router := gorouter.New()

	router.GET("/teachers/list", func(resp http.ResponseWriter, req *http.Request, params *gorouter.Param) {
		resp.Write([]byte("/teachers/list"))
	})

	router.GET("/teachers/:id/profile", func(resp http.ResponseWriter, req *http.Request, params *gorouter.Param) {
		resp.Write([]byte(fmt.Sprintf("%s = %s", "id", params.GetValue("id"))))
	})

	router.GET("/teachers/:id/profile/:id", func(resp http.ResponseWriter, req *http.Request, params *gorouter.Param) {
		resp.Write([]byte(fmt.Sprintf("id1 = %s; id2 = %s", params.Values[0], params.Values[1])))
	})

	http.ListenAndServe(":3001", router)

}

```

### 路由规则

gorouter 借鉴了httprouter的基数树实现方法。但当存在通配符和静态路由都匹配url时，优先匹配静态路由，如果匹配失败则返回再去匹配通配符。

```
路由：
① GET /users/:id/name   
② GET /users/id/name

请求：
/users/id/name   匹配②
/users/idd/name  匹配①

```

### Benchmark

echo的[测试用例](https://github.com/vishr/web-framework-benchmark)编写了[gorouter-example](https://github.com/ihornet/gorouter-example)，跑了下基准测试，感觉性能还不错。因为功能简单可能占些便宜。

```

goos: darwin
goarch: amd64
Benchmark_Echo_Static-8            	   30000	     42460 ns/op	    2413 B/op	     157 allocs/op
Benchmark_Echo_GitHubAPI-8         	   20000	     61322 ns/op	    2496 B/op	     203 allocs/op
Benchmark_Echo_GplusAPI-8          	  500000	      3255 ns/op	     173 B/op	      13 allocs/op
Benchmark_Echo_ParseAPI-8          	  300000	      5634 ns/op	     323 B/op	      26 allocs/op

Benchmark_Gorouter_Static-8        	   50000	     29292 ns/op	    1007 B/op	     157 allocs/op
Benchmark_Gorouter_GitHubAPI-8     	   30000	     58802 ns/op	    5666 B/op	     275 allocs/op
Benchmark_Gorouter_GplusAPI-8      	  500000	      3164 ns/op	     437 B/op	      22 allocs/op
Benchmark_Gorouter_ParseAPI-8      	  300000	      4543 ns/op	     615 B/op	      37 allocs/op

Benchmark_Gin_Static-8             	   30000	     52282 ns/op	    8693 B/op	     157 allocs/op
Benchmark_Gin_GitHubAPI-8          	   20000	     79637 ns/op	   10616 B/op	     203 allocs/op
Benchmark_Gin_GplusAPI-8           	  300000	      4409 ns/op	     681 B/op	      13 allocs/op
Benchmark_Gin_ParseAPI-8           	  200000	      8040 ns/op	    1421 B/op	      26 allocs/op

Benchmark_Beego_Static-8           	   10000	    198317 ns/op	   76586 B/op	    1099 allocs/op
Benchmark_Beego_GitHubAPI-8        	    5000	    269359 ns/op	   98868 B/op	    1422 allocs/op
Benchmark_Beego_GplusAPI-8         	  100000	     15628 ns/op	    6356 B/op	      91 allocs/op
Benchmark_Beego_ParseAPI-8         	   50000	     29614 ns/op	   12712 B/op	     182 allocs/op

Benchmark_Httprouter_Static-8      	  100000	     15696 ns/op	    1006 B/op	     157 allocs/op
Benchmark_Httprouter_GitHubAPI-8   	   50000	     38157 ns/op	   15583 B/op	     370 allocs/op
Benchmark_Httprouter_GplusAPI-8    	 1000000	      1874 ns/op	     735 B/op	      24 allocs/op
Benchmark_Httprouter_ParseAPI-8    	  500000	      2866 ns/op	     830 B/op	      42 allocs/op
PASS

```



