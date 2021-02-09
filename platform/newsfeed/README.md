This is my Go programming learning and training. Thanks to https://www.youtube.com/watch?v=LOn1GUsjOF4 

Post testing:  
```
curl -i -H "Content-Type: application/json" -d '{"Title":"Hello","Post":"World"}' http://10.0.1.248:8080/newsfeed
HTTP/1.1 204 No Content
date: Tue, 09 Feb 2021 13:26:29 GMT
strict-transport-security: max-age=15768000
```

Read post:
```
curl -i http://10.0.1.248:8080/newsfeed
HTTP/1.1 200 OK
content-type: application/json; charset=utf-8
date: Tue, 09 Feb 2021 13:26:52 GMT
content-length: 34
strict-transport-security: max-age=15768000

[{"title":"Hello","post":"World"}]
```