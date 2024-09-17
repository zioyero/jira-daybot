FROM golang:1.22.0

RUN mkdir /app

ADD . /app

WORKDIR /app
RUN make build

CMD ["/app/bin/cmd", "-output=slack"]
