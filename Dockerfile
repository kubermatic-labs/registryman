FROM golang:1.16.6-alpine3.14@sha256:a8df40ad1380687038af912378f91cf26aeabb05046875df0bfedd38a79b5499
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
RUN apk add --no-cache ca-certificates && update-ca-certificates

ADD . /go/src/github.com/kubermatic-labs/registryman
RUN cd /go/src/github.com/kubermatic-labs/registryman && \
    go build -a -ldflags '-extldflags "static" -w -s' -tags "exclude_graphdriver_devicemapper exclude_graphdriver_btrfs containers_image_openpgp" github.com/kubermatic-labs/registryman

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /go/src/github.com/kubermatic-labs/registryman/registryman /registryman
ENTRYPOINT ["/registryman"]
