FROM golang:1.14

WORKDIR /go/src/app
COPY . .
ENV GOPROXY=https://goproxy.cn,direct \
    HIT_API=1 \
    HIT_ADDR=http://localhost:80 \
    HIT_API_PORT=9999 \
    HIT_GROUPS=default,tmp \
    HIT_PEER_ADDRS=""
RUN go get -d -v ./... \
    && go install -v ./...

CMD ["bash","shell/run.sh"]