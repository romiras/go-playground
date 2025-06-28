#!/bin/sh

ruby main.rb 100000 &
pid=$!
pidstat -r -p $pid 1 > rss_usage_100k.log
