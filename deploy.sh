#!/bin/bash
git pull origin master --force
mkdir bin 2>/dev/null
go build -o bin/ .
killall tbot
nohup bin/tbot >bin/tbot.log 2>&1 &
