FROM alpine:3.8

RUN apk upgrade --update --no-cache

USER nobody

ADD build/_output/bin/local-storage-operator /usr/local/bin/local-storage-operator
