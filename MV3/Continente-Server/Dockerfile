FROM golang:1.23.0

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o /app/bin/ContinenteServer .

CMD ["/app/bin/ContinenteServer"]