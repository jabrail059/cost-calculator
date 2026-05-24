FROM golang:1.25.6-alpine
WORKDIR "/app"
COPY . .
RUN go mod download
RUN go build -o server ./cmd/app
CMD ["./server"]
