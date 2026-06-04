FROM golang:1.26.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build: static binary, no C dependencies, linux target
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# staage 2: Minimal runtime image
FROM scratch 

COPY --from=builder /app/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]