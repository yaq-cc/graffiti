FROM    golang:1.17-buster as builder
WORKDIR /app
COPY    . ./
RUN     go build

FROM    debian:buster-slim
RUN     set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
            ca-certificates && \
            rm -rf /var/lib/apt/lists/*
COPY    --from=builder /app/graffiti /app/graffiti

CMD     /app/graffiti