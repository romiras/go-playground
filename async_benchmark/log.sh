#!/bin/sh

./async_benchmark 500000 &
pid=$!
pidstat -r -p $pid 1 > rss_usage_500k.log
