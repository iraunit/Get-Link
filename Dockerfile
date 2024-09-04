FROM golang:1.21.5

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go mod vendor
RUN go mod download
RUN go mod tidy

RUN go install github.com/google/wire/cmd/wire@latest

WORKDIR /app/cmd

RUN wire

RUN go build -o server .

EXPOSE 8080

CMD ["./server"]