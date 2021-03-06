FROM golang:1.15-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk add --no-cache tzdata
ENV TZ Europe/Minsk
WORKDIR /root
COPY --from=build /app/main .
CMD ["./main"]