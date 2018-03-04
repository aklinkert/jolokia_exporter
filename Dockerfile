FROM instrumentisto/glide:0.13.0 as builder
WORKDIR /go/src/github.com/scalify/jolokia_exporter/

COPY glide.yaml glide.lock ./
RUN glide install --strip-vendor

COPY . ./
RUN CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -o bin/jolokia_exporter .


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/scalify/jolokia_exporter/bin/jolokia_exporter .
RUN chmod +x jolokia_exporter
ENTRYPOINT ["./jolokia_exporter"]
