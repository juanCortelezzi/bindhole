from golang:1-alpine as builder

workdir /app

env GOCACHE=/root/.cache/go-build

copy . .

run \
  --mount=type=cache,target="/root/.cache/go-build" \
  go build -o main ./main.go

from alpine:3

copy --from=builder /app/main .

env RPZ_CONFIG_FILE=blacklists.toml
env RPZ_OUTPUT_FILE=latest_block.list

volume /etc/bindhole/output
volume /etc/bindhole/config

entrypoint ["./main", "-config", "/etc/bindhole/config/$RPZ_CONFIG_FILE", "-output", "/etc/bindhole/output/$RPZ_OUTPUT_FILE"]
