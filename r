#!/bin/bash

function psql() {
    PGPASSWORD=DazBMyGQdKqKG command psql -h localhost -U homework
}

"$@"

