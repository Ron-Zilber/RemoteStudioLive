#!/bin/bash
server_ip=$(ifconfig | grep 'inet [0-9]\+\.[0-9]\+\.[0-9]\+\.[0-9]\+' | awk '{print $2}' | head -n 1)
connType="tcp"
go run server.go $connType "$server_ip" 7777 song