#!/bin/bash
for i in {1..6}; do
    curl http://localhost:8080/
    echo
done

for i in {1..11}; do
    curl -H "API_KEY: test_token" http://localhost:8080/
    echo
done
