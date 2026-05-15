#!/bin/bash

# Create an internal list of image IDs matching specific repository names
image_ids=$(podman images --format "{{.ID}}" \
  --filter "reference=localhost/strunck/main.go*" \
  --filter "reference=docker.io/strunck/main.go*" \
  --filter "reference=docker.io/strunck/uv4-middleware*" \
  --filter "reference=localhost/strunck/uv4-middleware*" \
  --filter "reference=localhost/uv4-middleware*")

# Remove each image in the list
if [ -n "$image_ids" ]; then
  echo "Removing images:"
  echo "$image_ids"
  podman image rm -f $image_ids
else
  echo "No matching images found"
fi

# Prune dangling images
echo "Pruning dangling images..."
podman image prune -f
