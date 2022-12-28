FROM golang:1.19.4-bullseye AS builder
COPY . /app
WORKDIR /app
RUN go build -o tz main.go

FROM alpine:3.17.0
WORKDIR /app
COPY --from=builder  /app/tz /app/tz

EXPOSE 8080
CMD ["./tz"]