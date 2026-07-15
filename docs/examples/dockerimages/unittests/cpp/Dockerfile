FROM alpine:3.8

RUN apk --no-cache add cmake clang clang-dev make gcc g++ libc-dev linux-headers

RUN mkdir /src
COPY run.sh run.sh
RUN chmod -R 705 run.sh

ENTRYPOINT ["/run.sh"]