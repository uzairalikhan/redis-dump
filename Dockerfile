FROM golang:1.8.3-alpine

RUN apk add -uU git
RUN go get github.com/go-redis/redis github.com/Sirupsen/logrus 
COPY . /go/src/github.com/uzairalikhan/redis-dump
WORKDIR /go/src/github.com/uzairalikhan/redis-dump

CMD [ "go", "run", "redis-dump.go" ]