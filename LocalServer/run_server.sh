#!/bin/bash
server_ip=$(ifconfig | grep 'inet [0-9]\+\.[0-9]\+\.[0-9]\+\.[0-9]\+' | awk '{print $2}' | head -n 1)
go run server.go "$server_ip" 8080 song