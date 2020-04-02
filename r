#!/bin/bash

function psql() {
    docker exec -it homework200301_users_db_1 psql -U homework -d users_db
}

function du() {
    docker-compose up -d --build
}

function dd() {
    docker-compose down --remove-orphans
}

function del() {
    docker volume rm \
        homework200301_users_data \
        homework200301_users_logs
}

function logs() {
    docker logs homework200301_users_rest_1
}

"$@"

