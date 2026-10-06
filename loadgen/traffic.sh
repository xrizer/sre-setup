#!/bin/sh
# Steady background traffic: ~10 requests per second to /book.
while true; do
  for i in 1 2 3 4 5; do
    curl -s -o /dev/null http://booking-api:8080/book &
  done
  wait
  sleep 0.5
done
