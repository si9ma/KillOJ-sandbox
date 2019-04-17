FROM golang:1.12.4

RUN go get -v -d github.com/si9ma/KillOJ-sandbox
RUN go install -v github.com/si9ma/KillOJ-sandbox