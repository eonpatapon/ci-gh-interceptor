FROM golang:1.13 AS builder
ADD . /build
WORKDIR /build
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /main .

# final stage
FROM alpine:latest
COPY --from=builder /main ./
RUN chmod +x ./main
ENTRYPOINT ["./main"]
EXPOSE 8080
