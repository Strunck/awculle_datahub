#!/bin/bash

echo "pod-uv4pod.service muss laufen"
if systemctl --user is-active --quiet pod-uv4pod.service; then
    echo "pod-uv4pod.service, schön..."
    systemctl --user start pod-uv4pod.service
else
    echo "pod-uv4pod.service ist nicht aktiv\nDenn starte ich den pod mal..."
    systemctl --user start pod-uv4pod.service
fi

echo "middleware Container püfen..."
if systemctl --user is-active --quiet container-middleware.service; then
    echo "container-middleware.service läuft, beende den container...."
    systemctl --user stop container-middleware.service
fi

VERSION=$1-dev
echo "dev Container für middleware $VERSION starten...."

podman run -it --rm \
    --name middelware-dev \
    --pod=uv4pod \
    --workdir="/go/src/github.com/strunck/uv4-middleware" \
    -e "VERSION=$VERSION" \
    -e "air_wd=/go/src/github.com/strunck/uv4-middleware/" \
    -e "KO_DATA_PATH=/go/src/github.com/strunck/uv4-middleware/kodata" \
    --env-file="kodata/.env_AAB" \
    -v "/opt/uv4/shared:/shared:z" \
    -v $(pwd):/go/src/github.com/strunck/uv4-middleware:Z \
    localhost/swair:v1.63.4 -c air-in-pod.toml
    
    # cosmtrek/air -c air-in-pod.toml

#--build.cmd "go build main.go" --build.delay "100"