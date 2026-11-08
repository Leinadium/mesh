FROM golang:1.26-alpine

WORKDIR /
COPY go.mod go.sum /
RUN go mod download

COPY main.go /
RUN go build /main.go

CMD ["/main"]
