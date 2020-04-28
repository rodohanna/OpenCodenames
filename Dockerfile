FROM golang:alpine

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN mkdir /client
RUN mkdir /server
RUN mkdir -p /dist/data
RUN mkdir -p /dist/static-assets

RUN apk update && apk add yarn && apk add git

WORKDIR /client
COPY ./client .
RUN yarn && yarn build
RUN cp -a build/. /dist/static-assets/


WORKDIR /server
COPY ./server/go.mod .
COPY ./server/go.sum .
RUN go mod download
COPY ./server .
RUN go build -o main .
RUN cp main /dist
RUN cp data/wordlist.txt /dist/data
RUN cp chunkynut-key.json /dist
RUN cp recaptcha-key.txt /dist

ENV HEROKU_APP_URL=https://chunky-codenames.herokuapp.com/

EXPOSE 8080

WORKDIR /dist

CMD ["./main"]