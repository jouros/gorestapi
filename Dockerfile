FROM golang:latest
RUN mkdir /restapi
ADD . /restapi
WORKDIR /restapi
## Add this go mod download command to pull in any dependencies
RUN go mod download
## Our project will now successfully build with the necessary go libraries included.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o httpd/main httpd/main.go 
CMD ["/restapi/httpd/main"]
EXPOSE 8080/tcp
