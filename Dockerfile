FROM golang:alpine AS builder
WORKDIR /go/src/app
COPY ./main.go /go/src/app
RUN CGO_ENABLED=0 GOOS=linux go build -tags timetzdata -o timezones main.go
EXPOSE 8080

FROM scratch
COPY --from=builder /go/src/app/timezones /
CMD ["./timezones"]
