FROM alpine

RUN apk --update add ca-certificates

#Fix for missing glibc
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

ADD metricPodScaler /

ENTRYPOINT ["/metricPodScaler"]
