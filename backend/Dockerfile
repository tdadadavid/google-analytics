FROM golang:1.22.5 AS builder
WORKDIR /app
COPY . .
RUN go build -o ganalytics

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/ganalytics .
CMD ["./ganalytics"]