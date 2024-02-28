FROM golang:1.22 as builder

WORKDIR /usr/src/app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o metrics cmd/metrics.go

FROM alpine as certificates

RUN apk add --no-cache ca-certificates

FROM scratch

COPY --from=builder /usr/src/app/metrics /metrics
COPY --from=certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 2112

ENTRYPOINT ["/metrics"]