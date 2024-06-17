from golang:1-alpine as builder

workdir /app

env GOCACHE=/root/.cache/go-build

copy . .

run \
  --mount=type=cache,target="/root/.cache/go-build" \
  go build -o main ./main.go

from alpine:3

copy --from=builder /app/main .

ENV BINDHOLE_RPZ_FILE=bindhole.zone

volume /etc/bindhole/output
volume /etc/bindhole/config

entrypoint ./main \
  -config /etc/bindhole/config/blacklists.toml \
  -output /etc/bindhole/output/$BINDHOLE_RPZ_FILE
