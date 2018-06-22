FROM golang

WORKDIR /go/src/app
COPY . .

RUN go get -v github.com/influxdata/influxdb/client/v2
RUN go build

CMD ["./app"]
