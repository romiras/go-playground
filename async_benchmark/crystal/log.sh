#!/bin/sh

crystal main.cr 50000 &
pid=$!
pidstat -r -p $pid 1 > rss_usage_50k.log
