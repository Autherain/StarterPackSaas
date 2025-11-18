FROM docker.io/library/golang:alpine AS base

RUN addgroup -g 10001 -S agent \
  && adduser -D -G agent -s /sbin/nologin -S -u 10001 agent

RUN apk add --no-cache ca-certificates tzdata

FROM docker.io/library/golang:alpine AS builder

WORKDIR /build/

COPY ./go.* .

ENV CGO_ENABLED=0 GOARCH=amd64 GOOS=linux
COPY . .
RUN go build -o /build/bin/server ./cmd/api/

FROM scratch

COPY --from=base /etc/passwd /etc/passwd
USER agent

COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=base /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /home/agent/bin/
COPY --chown=agent --from=builder /build/bin/server .

CMD ["/home/agent/bin/server"]

