FROM golang:1.24.1-bullseye AS builder
LABEL authors="scotta01"

WORKDIR /go/src/
COPY ./ /go/src/
WORKDIR /go/src/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/app /go/src/cmd/tls-scrape/
RUN echo "nobody:x:65534:65534:nobody:/:" > /go/src/passwd
RUN chown nobody /go/bin/app
RUN chmod +x /go/bin/app

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/passwd /etc/passwd
COPY --from=builder /go/bin/app /scotta01/tls-scrape
USER nobody

ENTRYPOINT ["/scotta01/tls-scrape"]
