FROM golang:1.10 AS build
WORKDIR /go/src/github.com/hpidcock/zfsi
ADD cmd cmd
ADD pkg pkg
ADD vendor vendor
WORKDIR /output
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /output/zfsi-config-service \
    github.com/hpidcock/zfsi/cmd/zfsi-config-service

FROM alpine:3.7
RUN apk --no-cache add ca-certificates
COPY --from=build /output/zfsi-config-service /etc/local/bin/zfsi-config-service
ENTRYPOINT ["/etc/local/bin/zfsi-config-service"]