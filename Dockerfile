# Project Build Stage
FROM golang:1.13 as builder
WORKDIR /app
COPY . /app
RUN cd /app \
    && CGO_ENABLED=0 GOOS=linux make build \
    && cp -a bin/ghr /go/bin/ghr

# Project Image Stage
FROM scratch
COPY --from=builder /app/bin/ghr /go/bin/
ENTRYPOINT [ "/go/bin/ghr" ]