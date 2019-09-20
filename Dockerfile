FROM golang:alpine as builder
WORKDIR /go/src/github.com/scalify/jolokia_exporter/
COPY . ./
RUN apk add --no-cache git \
 && go get \
 && CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -o bin/jolokia_exporter .

FROM alpine:latest
COPY --from=builder /go/src/github.com/scalify/jolokia_exporter/bin/jolokia_exporter .
RUN apk --no-cache add ca-certificates \
 && chmod +x jolokia_exporter
ENTRYPOINT ["./jolokia_exporter"]
