#!/bin/bash
# Check if IP address argument is provided
if [ $# -eq 0 ]; then
    echo "Usage: $0 <ip_address>"
    exit 1
fi

# Assign IP address argument to a variable
ip_address="$1"

# Run the Go program with specified IP address and 'song' argument
go run ClientUtils.go client.go tcp "$ip_address" 8080 song | mpg123 -
