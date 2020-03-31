#!/bin/bash

function psql() {
    PGPASSWORD=DazBMyGQdKqKG command psql -h localhost -U homework
}

function du() {
    docker-compose up -d
}

function dd() {
    docker-compose down --remove-orphans
}

function del() {
    docker volume rm homework200301_users_data
}

"$@"

