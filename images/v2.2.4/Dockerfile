FROM golang:alpine as builder

WORKDIR /go/delivery
COPY . /go/delivery

RUN apk --update add git gcc curl alpine-sdk libc6-compat
RUN apt-get update && apt-get install --no-install-recommends -y \
    build-essential \
    ca-certificates \
    git

RUN export ENV GO111MODULE=on && \
    export GOPRIVATE=github.com/segment*
RUN go build \
  -o /go/run/worker \
  -a ./cmd/worker

FROM 528451384384.dkr.ecr.us-west-2.amazonaws.com/segment-alpine
COPY --from=builder /go/run/worker /
