#!/bin/bash

if [ "$1" = "deploy" ]; then
    go run genkey.go deploy
elif [ "$1" = "destroy" ]; then
    go run genkey.go destroy
else
    echo "Usage: $0 [deploy|destroy]"
    exit 1
fi

