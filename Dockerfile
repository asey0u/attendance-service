FROM golang:1.25.7 AS build

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -mod=vendor -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=build /out/server /app/server
COPY migrations /app/migrations

EXPOSE 8080
CMD ["/app/server"]
