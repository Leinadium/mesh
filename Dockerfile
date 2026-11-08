# syntax=docker/dockerfile:1

# FROM golang:1.26-alpine
# WORKDIR /
# COPY go.mod go.sum /
# RUN go mod download
# COPY main.go /
# RUN go build /main.go
# CMD ["/main"]

# bulding backend
FROM golang:1.26-alpine AS backend
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN go build -o server ./cmd/main.go

# final image
FROM gcr.io/distroless/static-debian13
WORKDIR /app
COPY --from=backend /app/server .

CMD ["./server"]
