FROM golang:1.8.3-alpine

COPY . /go/src/github.com/uzairalikhan/redis-dump

WORKDIR /go/src/github.com/uzairalikhan/redis-dump

RUN apk add -uU git
RUN go get github.com/go-redis/redis github.com/Sirupsen/logrus 

CMD [ "go", "run", "redis-dump.go" ]