#!/usr/bin/env bash

if [ "$1" = "init" ]; then

    if [ -f "/tmp/minitwit.db" ]; then
        echo "Database already exists."
        exit 1
    fi
    echo "Putting a database to /tmp/minitwit.db..."
    sqlite3 /tmp/minitwit.db < schema.sql
elif [ "$1" = "start" ]; then
    echo "Starting minitwit..."
    nohup go run . > /tmp/out.log 2>&1 &
elif [ "$1" = "stop" ]; then
    echo "Stopping minitwit..."
    pkill -f minitwit
elif [ "$1" = "inspectdb" ]; then
    go run flag_tool.go -i | less
elif [ "$1" = "flag" ]; then
    go run flag_tool.go "$@"
else
  echo "I do not know this command..."
fi
