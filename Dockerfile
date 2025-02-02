###############################################################################
# BUILD STAGE

FROM cgr.dev/chainguard/go:latest-dev AS builder
RUN mkdir /build
COPY app /build/
WORKDIR /build
RUN apk update \
    && apk upgrade \
    && apk add --no-cache git \
    && go mod tidy \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' . \
    && chmod 755 /build/*
    
###############################################################################
# PACKAGE STAGE

FROM cgr.dev/chainguard/go:latest-dev
EXPOSE 8080
COPY --from=builder /build/* /app/
ENTRYPOINT ["/app/envoy-api"]