#!/usr/bin/env bash

# for i in {1..10}; do
  curl -X POST http://localhost:8000/send \
    -H "Content-Type: application/json" \
    -d '{
      "message": "Hello there!",
      "phone_number": "+1234567890",
      "org": "arvancloud",
      "is_express": false
    }'
  # curl -X POST http://localhost:8000/send \
  #   -H "Content-Type: application/json" \
  #   -d '{
  #     "message": "Hello there!",
  #     "phone_number": "+1234567890",
  #     "org": "foo",
  #     "is_express": false
  #   }'
# done