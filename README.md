# Дз по курсу Безопасность интернет приложений

## Запуск прокси сервера

```
make run
```

## Пример
1. HTTP-прокси
```
$ curl -v -x http://127.0.0.1:8080 http://mail.ru
*   Trying 127.0.0.1:8080...
* TCP_NODELAY set
* Connected to 127.0.0.1 (127.0.0.1) port 8080 (#0)
> GET http://mail.ru/ HTTP/1.1
> Host: mail.ru
> User-Agent: curl/7.68.0
> Accept: */*
> Proxy-Connection: Keep-Alive
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 301 Moved Permanently
< Connection: keep-alive
< Content-Length: 185
< Content-Type: text/html
< Date: Mon, 21 Feb 2022 21:58:54 GMT
< Location: https://mail.ru/
< Server: nginx/1.14.1
< 
<html>
<head><title>301 Moved Permanently</title></head>
<body bgcolor="white">
<center><h1>301 Moved Permanently</h1></center>
<hr><center>nginx/1.14.1</center>
</body>
</html>
```
2. HTTPS-прокси
```
$ curl -v -x http://127.0.0.1:8080 https://mail.ru
*   Trying 127.0.0.1:8080...
* TCP_NODELAY set
* Connected to 127.0.0.1 (127.0.0.1) port 8080 (#0)
* allocate connect buffer!
* Establish HTTP proxy tunnel to mail.ru:443
> CONNECT mail.ru:443 HTTP/1.1
> Host: mail.ru:443
> User-Agent: curl/7.68.0
> Proxy-Connection: Keep-Alive
> 
```
