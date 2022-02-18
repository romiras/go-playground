FROM golang:1.15

ADD . /app
WORKDIR /app

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN go build -ldflags="-w -s" -o ttt


FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
#FROM debian:stable-slim
#RUN apt-get update && apt-get install -y ca-certificates tzdata
WORKDIR /app
COPY --from=0 /app/ttt .

CMD ["./ttt", "/var/log/app.log"]
