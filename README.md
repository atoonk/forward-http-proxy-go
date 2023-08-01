A simple HTTP forward proxy server in Go

start the proxy, it will listen on port 8080
```
go run main.go
```

then test using curl:
```
 curl -x http://localhost:8080 https://www.example.com
```

the proxy will log the request like this
```
$ go run main.go
2023/07/31 19:51:53 Received request CONNECT www.example.com:443 127.0.0.1:56651
2023/07/31 19:51:53 Tunneling from 127.0.0.1:56651 to www.example.com:443
```
