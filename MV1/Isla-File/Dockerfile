FROM golang:1.23.0

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o /app/bin/IslaFile .

EXPOSE 50052
CMD ["/app/bin/IslaFile"]