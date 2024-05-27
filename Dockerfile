FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder
LABEL maintainer="taodev <taodev@qq.com>"
COPY . /go/src/github.com/taodev/gotun
WORKDIR /go/src/github.com/taodev/gotun
ARG TARGETOS TARGETARCH
ARG GOPROXY="https://goproxy.cn,direct"
ENV GOPROXY ${GOPROXY}
ENV CGO_ENABLED=0
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
RUN set -ex \
    && apk add git \
    && go build \
        -o /go/bin/gotun \
        ./cmd/gotun
FROM --platform=$TARGETPLATFORM alpine AS dist
LABEL maintainer="taodev <taodev@qq.com>"
COPY --from=builder /go/bin/gotun /usr/local/bin/gotun
ENTRYPOINT ["gotun"]