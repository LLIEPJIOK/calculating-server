FROM golang:1.22

WORKDIR /

COPY ./ ./

RUN go mod download
RUN go build -o calculating-server ./cmd/main.go

EXPOSE 8080

CMD ["./calculating-server"]