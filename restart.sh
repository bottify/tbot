#!/bin/bash
killall tbot
nohup bin/tbot >>bin/tbot.log 2>&1 &
