#!/bin/sh

# Parse the JSON configuration file and retrieve values
NSDataBase=$(jq -r '.NSDataBase' db/config.json)
NSSonar=$(jq -r '.NSSonar' db/config.json)

printf "✅ Delete namespace %s\n" "$NSSonar"
kubectl delete ns "$NSSonar"

printf "✅ Delete namespace %s\n" "$NSDataBase"
kubectl delete ns "$NSDataBase"



