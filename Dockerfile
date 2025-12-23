FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/api

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/server /app/server
EXPOSE 8085
ENTRYPOINT ["/app/server"]