[HTTP](https://en.wikipedia.org/wiki/Hypertext_Transfer_Protocol) is the
protocol that powers the web. In this challenge, you'll build a HTTP/1.1 server
that is capable of serving multiple clients.

Along the way you'll learn about TCP servers,
[HTTP request syntax](https://www.w3.org/Protocols/rfc2616/rfc2616-sec5.html),
and more.

To init the Server:
```
go run app/server.go
```

Test command examples:
```
curl -v -X GET http://localhost:4221/

curl -v -X GET http://localhost:4221/big/massive-huge

curl -v -X GET http://localhost:4221/echo/huge/guy

curl -v -X GET http://localhost:4221/user-agent -H "User-Agent: Big/large-guy"

curl -v -X GET http://localhost:4221/files/a_very_large_guy

curl -v -X GET http://localhost:4221/files/non-existent_smallGuy

curl -v -X POST http://localhost:4221/files/big_large_huge_massive_guy -d 'what a massive dude'
```
