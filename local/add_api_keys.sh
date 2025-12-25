#!/bin/bash

# Function to generate a UUID
generate_uuid() {
    uuidgen | tr '[:upper:]' '[:lower:]'
}

KEY1=$(generate_uuid)
KEY2=$(generate_uuid)

echo "Generated API Keys:"
echo "1: $KEY1"
echo "2: $KEY2"

# Insert into Redis using podman exec
# Use 'api_keys' as the set name as defined in auth.go
echo "Inserting keys into Redis set 'api_keys'..."

if podman exec redis-dev redis-cli SADD api_keys "$KEY1" "$KEY2" > /dev/null 2>&1; then
    echo "Successfully inserted keys into Redis."
else
    echo "Error: Failed to insert keys into Redis. Is the 'redis-dev' container running?"
    exit 1
fi

echo "Current keys in 'api_keys' set:"
podman exec redis-dev redis-cli SMEMBERS api_keys
