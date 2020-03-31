#!/bin/bash

function psql() {
    docker exec -it homework200301_postgres_1 psql -U homework -d users_db
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

