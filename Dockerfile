FROM alpine:latest

RUN apk add --update-cache tzdata
COPY be-quote /be-quote

ENTRYPOINT ["/be-quote"]
