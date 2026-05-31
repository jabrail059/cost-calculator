FROM golang:1.25.6-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/app

FROM alpine:3.22

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=build /out/app ./app
COPY web ./web
COPY docs ./docs

ENV SERVER_ADDR=:8080 \
    FRONTEND_API_ADDR=:3000 \
    WEB_DIR=./web

EXPOSE 8080 3000

USER app

CMD ["./app"]
