#!/bin/bash

cmd=$1

if [ $cmd == "up" ]; then
    sea-orm-cli migrate -u $DATABASE_URI up
elif [ $cmd == "down" ]; then
    sea-orm-cli migrate -u $DATABASE_URI down
else
    echo "No such command '$cmd'"
fi
