FROM golang:1.18 as builder

RUN apt-get update && apt-get install --no-install-recommends -y \
    build-essential \
    ca-certificates \
    git

ENV SRC github.com/segmentio/smsk-bulrush-test

COPY . /go/src/${SRC}
WORKDIR /go/src/${SRC}

RUN make install

FROM 528451384384.dkr.ecr.us-west-2.amazonaws.com/segment-alpine

COPY --from=builder /go/bin/worker /bin/worker
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# force the go resolver instead of the cgo resolver.  otherwise, address resolution breaks
# for the .local domain
ENV GODEBUG netdns=go
