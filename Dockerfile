FROM alpine:3.11

RUN apk update && apk add ca-certificates

COPY ./webhook . 
ENTRYPOINT ["./webhook"]