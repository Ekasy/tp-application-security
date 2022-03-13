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
< HTTP/1.1 200 Connection established
< 
* Proxy replied 200 to CONNECT request
* CONNECT phase completed!
* ALPN, offering h2
* ALPN, offering http/1.1
* successfully set certificate verify locations:
*   CAfile: /etc/ssl/certs/ca-certificates.crt
  CApath: /etc/ssl/certs
* TLSv1.3 (OUT), TLS handshake, Client hello (1):
* CONNECT phase completed!
* CONNECT phase completed!
* TLSv1.3 (IN), TLS handshake, Server hello (2):
* TLSv1.3 (IN), TLS handshake, Encrypted Extensions (8):
* TLSv1.3 (IN), TLS handshake, Certificate (11):
* TLSv1.3 (IN), TLS handshake, CERT verify (15):
* TLSv1.3 (IN), TLS handshake, Finished (20):
* TLSv1.3 (OUT), TLS change cipher, Change cipher spec (1):
* TLSv1.3 (OUT), TLS handshake, Finished (20):
* SSL connection using TLSv1.3 / TLS_AES_128_GCM_SHA256
* ALPN, server did not agree to a protocol
* Server certificate:
*  subject: CN=mail.ru
*  start date: Mar 13 09:24:41 2022 GMT
*  expire date: Mar 10 09:24:41 2032 GMT
*  common name: mail.ru (matched)
*  issuer: CN=yngwie proxy CA
*  SSL certificate verify ok.
> GET / HTTP/1.1
> Host: mail.ru
> User-Agent: curl/7.68.0
> Accept: */*
> 
* TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Transfer-Encoding: chunked
< Cache-Control: no-cache,no-store,must-revalidate
< Connection: keep-alive
< Content-Security-Policy: ...; report-uri https://cspreport.mail.ru/splash?mode=csp&v=08.02.22;
< Content-Type: text/html; charset=utf-8
< Date: Sun, 13 Mar 2022 09:24:41 GMT
< Expires: Sat, 13 Mar 2021 09:24:41 GMT
< Pragma: no-cache
< Server: nginx/1.14.1
< Set-Cookie: act=5c0ebbbd6b874e1d91b3494f33384597; path=/; domain=.mail.ru; Secure; HttpOnly; SameSite=None
< Set-Cookie: mrcu=4619622DB85953E3B2E19B7CFC6D; expires=Wed, 10 Mar 2032 09:24:41 GMT; path=/; domain=.mail.ru; Secure; HttpOnly; SameSite=None
< Strict-Transport-Security: max-age=16070400
< X-Content-Type-Options: nosniff
< X-Etime: 0.027
< X-Frame-Options: SAMEORIGIN
< X-Host: lf208.m.smailru.net
< X-Mru-Request-Id: 0a7b2bc3
< X-Xss-Protection: 1; mode=block; report=https://cspreport.mail.ru/xxssprotection
< 
<!DOCTYPE html> ...
```

## Организация хранения запрос-ответов в базе MongoDB
Генерируем уникальный reqid на запрос, и храним рядом данные запроса и ответа
```
{
	"_id" : ObjectId("622dc6755f881fdf493c60d8"),
	"reqid" : "yLwzMIOuPC!pJTXu0imTyU&ruMtiyNAsP!*x=jYOkJ7mmq16V6zHrEetivcuZ8KK",
	"request" : {
		"timestamp" : 1647167093,
		"method" : "GET",
		"path" : "/",
		"cgi_params" : {
			
		},
		"headers" : {
			"Proxy-Connection" : "Keep-Alive",
			"User-Agent" : "curl/7.68.0",
			"Accept" : "*/*"
		},
		"cookies" : {
			
		},
		"body" : ""
	},
	"response" : {
		"timestamp" : 1647167093,
		"status_code" : 301,
		"message" : "301 Moved Permanently",
		"headers" : {
			"Location" : "https://mail.ru/",
			"Server" : "nginx/1.14.1",
			"Date" : "Sun, 13 Mar 2022 10:24:53 GMT",
			"Content-Type" : "text/html",
			"Content-Length" : "185",
			"Connection" : "keep-alive"
		},
		"body" : ""
	}
}
```