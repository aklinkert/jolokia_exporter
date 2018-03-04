FROM golang:1.9.1 as builder
WORKDIR /go/src/github.com/scalify/jolokia_exporter/fixtures/

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-s' -installsuffix cgo -o test_server test_server.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/scalify/jolokia_exporter/fixtures/test_server .
COPY metrics.json .
RUN chmod +x test_server
ENTRYPOINT ["./test_server"]
