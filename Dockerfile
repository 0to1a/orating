FROM oven/bun:1 AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/bun.lock* ./
RUN bun install --frozen-lockfile
COPY frontend ./
RUN bun run build

FROM golang:1.26-alpine AS backend
WORKDIR /app
RUN apk add --no-cache git make
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/build ./internal/platform/frontend/frontend_dist
RUN sqlc generate
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=backend /server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
