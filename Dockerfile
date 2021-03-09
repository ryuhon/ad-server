FROM golang:1.11.2-alpine as builder
RUN mkdir /go/src/github.com
RUN mkdir /go/src/github.com/ryuhon/ad-server
WORKDIR /go/src/github.com/ryuhon/ad-server
COPY . /go/src/github.com/ryuhon/ad-server
RUN apk update && apk upgrade && apk add --no-cache bash git openssh
RUN go get github.com/labstack/echo
RUN go get github.com/go-sql-driver/mysql
RUN go get github.com/dgrijalva/jwt-go
RUN go build


FROM golang:1.11.2-alpine
RUN mkdir /app
WORKDIR /app
COPY --from=builder /go/src/github.com/ryuhon/ad-server /app
RUN apk add tzdata
RUN cp /usr/share/zoneinfo/Asia/Seoul /etc/localtime
RUN echo "Asia/Seoul" > /etc/timezone
RUN apk del tzdata


ENTRYPOINT ["./ad-server"]