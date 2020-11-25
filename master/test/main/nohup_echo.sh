#!/bin/sh
nohup ./http_echo config.json > /dev/null 2> /dev/null & echo $!