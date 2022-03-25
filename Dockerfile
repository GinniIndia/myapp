FROM golang:1.16
RUN mkdir /build
ADD go.mod go.sum app.go /build/
WORKDIR /build
RUN go build