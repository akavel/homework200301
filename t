#!/bin/bash

function get() {
    (set -x ; curl -i http://localhost:8080/v1/user ) ; echo
    (set -x ; curl -i http://localhost:8080/v1/user/john@smith.name ) ; echo
    (set -x ; curl -i http://localhost:8080/v1/user/john@smith.nam ) ; echo
    (set -x ; curl -i http://localhost:8080/v1/user/jane@example.com ) ; echo
}

function post() {
    curl -i -XPOST -HContent-Type:application/json -d@testdata/jane.json http://localhost:8080/v1/user ; echo
}

function put() {
    curl -i -XPUT -HContent-Type:application/json -d@"${1:-testdata/jane2.json}" http://localhost:8080/v1/user/jane@example.com ; echo
}

function del() {
    curl -i -XDELETE http://localhost:8080/v1/user/jane@example.com ; echo
}

"$@"

