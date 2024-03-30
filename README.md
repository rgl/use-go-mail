# About

This uses the [wneessen/go-mail library](https://github.com/wneessen/go-mail) to send a text mail to a [SMTP(S)](https://en.wikipedia.org/wiki/Simple_Mail_Transfer_Protocol) mail server.

# Usage (Ubuntu 22.04)

```bash
# create the environment defined in docker-compose.yml
# and leave it running in the background.
docker compose up --detach --build --wait

# show running containers.
docker compose ps

# show logs.
docker compose logs

# open a container network interface in wireshark.
./wireshark.sh mailpit &

# send email.
http \
  --verbose \
  POST \
  http://localhost:8000

# open mailpit (email).
xdg-open http://localhost:8025

# destroy the environment.
docker compose down --remove-orphans --volumes --timeout=0
```
