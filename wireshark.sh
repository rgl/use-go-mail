#!/bin/bash
set -euo pipefail

script_directory_path="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
script_path="$script_directory_path/$(basename "${BASH_SOURCE[0]}")"

container_name="${1:-quotes}"; shift || true
container_id="$(docker compose ps --no-trunc --format '{{.ID}}' "$container_name")"
container_br="br-$(docker inspect --format '{{range .NetworkSettings.Networks}}{{.NetworkID}}{{end}}' "$container_id" | cut -c1-12)"
container_mac="$(docker inspect --format '{{range .NetworkSettings.Networks}}{{.MacAddress}}{{end}}' "$container_id")"

# call wireshark.
exec wireshark \
    -o "gui.window_title:$container_name" \
    -k \
    -d 'tcp.port==1025,smtp' \
    -d 'tcp.port==8025,http' \
    -i "$container_br" \
    -f "ether host $container_mac"
