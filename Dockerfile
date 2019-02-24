FROM golang:1.11.5

#RUN apt-get update
#RUN apt-get install vim -y
RUN go get "github.com/go-sql-driver/mysql"
RUN go get "github.com/gin-gonic/gin"
RUN go get "github.com/jinzhu/gorm"


WORKDIR /go/src/app
COPY main.go main.go
RUN go build ./main.go
CMD ["./main"]