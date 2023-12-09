#!/bin/bash

# Define Go app endpoint
API_URL="http://localhost:8080/tags"

# Create a temporary file for storing the batch of tags
TMP_FILE=$(mktemp)

# Generate 1000 random tags
echo "[" >> "$TMP_FILE"
for i in {1..1000}; do
  tag_name="tag-$RANDOM"
  product_id=$((RANDOM % 1000))
  echo "{\"tag\":\"$tag_name\",\"product_id\":$product_id}" >> "$TMP_FILE"
  if [ "$i" -lt 1000 ]; then
    echo "," >> "$TMP_FILE"
  fi
done
echo "]" >> "$TMP_FILE"

# Send the batch of tags to the Go app endpoint
curl -X POST -H "Content-Type: application/json" -d "@$TMP_FILE" "$API_URL"

# Clean up the temporary file
rm "$TMP_FILE"
