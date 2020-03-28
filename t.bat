@echo off
goto :%1
goto :eof

:get
@echo on
curl -i http://localhost:8080/v1/user
curl -i http://localhost:8080/v1/user/john@smith.name
curl -i http://localhost:8080/v1/user/john@smith.nam
curl -i http://localhost:8080/v1/user/jane@example.com
@echo off
goto :eof

:post
curl -i -XPOST -HContent-Type:application/json -d@testdata/jane.json http://localhost:8080/v1/user
goto :eof

:del
curl -i -XDELETE http://localhost:8080/v1/user/jane@example.com
goto :eof

