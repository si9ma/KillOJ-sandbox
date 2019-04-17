FROM golang:1.12.4

RUN apt-get -y update
RUN apt-get install -y libseccomp-dev seccomp

RUN go get -v -d github.com/si9ma/KillOJ-sandbox
RUN go build -o /usr/local/bin/kbox -v github.com/si9ma/KillOJ-sandbox

ENTRYPOINT ["kbox"]
CMD ["help"]