FROM golang:1.25.7

WORKDIR /app

COPY . .

RUN go build -mod=vendor -o main cmd/server/main.go

EXPOSE 8080

CMD ["./main"]