FROM golang:alpine

WORKDIR /workdir

RUN apk add --no-cache --virtual .build-deps \
    git \
        ; \
        go get golang.org/x/tools/cmd/goimports && \
        rm -rf /go/src/golang.org \
        apk del .build-deps;

WORKDIR /drone/src
# ENTRYPOINT ["goimports"]
