#!/bin/bash

#  Teste de IP do rate limiter
echo "Testando IP rate limiting..."
for i in {1..6}; do
    echo "Request $i:"
    curl -i http://localhost:8080/
    echo -e "\n"
    sleep 0.1
done

echo "Aguarde o teste reiniciar..."
sleep 2

#  Teste de token do rate limiter
echo "Testando token rate limiting..."
for i in {1..11}; do
    echo "Request $i:"
    curl -i -H "API_KEY: abc123" http://localhost:8080/
    echo -e "\n"
    sleep 0.1
done
