FROM golang:1.13-alpine
WORKDIR /app
ARG PORT_ENV=8084
ARG HOST_SENSORS=192.168.0.49
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .
EXPOSE 8084
CMD ["./main"]