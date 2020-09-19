FROM golang:1.14

WORKDIR /go/src/app
COPY . .
ENV GOPROXY=https://goproxy.cn,direct \
    HIT_CONFIG_PATH=cmd/hit/hit.toml
RUN go get -d -v ./... \
    && go install -v ./...
CMD ["bash","shell/run.sh"]