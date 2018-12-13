FROM golang:stretch

WORKDIR /go/src/github.com/PolarGeospatialCenter/local-storage-operator
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

COPY Gopkg.lock Gopkg.toml ./
RUN dep ensure -vendor-only
COPY ./ .
RUN go build -o /bin/local-storage-operator ./cmd/manager

FROM debian:stretch-slim
COPY --from=0 /bin/local-storage-operator /bin/local-storage-operator
ENTRYPOINT /bin/local-storage-operator
