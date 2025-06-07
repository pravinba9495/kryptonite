FROM docker.io/golang:1.24-alpine3.21 AS build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM docker.io/alpine:3.21
WORKDIR /app
COPY --from=build-stage /app/main .
CMD ["./main"]

