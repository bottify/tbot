#!/bin/bash
#git pull origin master --force
#git submodule update
source /opt/rh/devtoolset-7/enable
mkdir bin 2>/dev/null
go build -o bin/ .
killall tbot
nohup bin/tbot >>bin/tbot.log 2>&1 &
