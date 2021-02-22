FROM golang:latest
RUN mkdir /restapi
RUN mkdir /restapi/platform
RUN mkdir /restapi/platform/data
RUN mkdir /restapi/platform/initialize
RUN mkdir /restapi/platform/initialize/sql
ADD go.mod /restapi
ADD go.sum /restapi
ADD main.go /restapi
ADD ./platform/data/data.go /restapi/platform/data
ADD ./platform/data/open_data.go /restapi/platform/data
ADD ./platform/initialize/initialize_db.go /restapi/platform/initialize
ADD ./platform/initialize/sql/1_create_table.down.sql /restapi/platform/initialize/sql
ADD ./platform/initialize/sql/1_create_table.up.sql /restapi/platform/initialize/sql
WORKDIR /restapi
#
RUN go get github.com/golang-migrate/migrate/v4/database/postgres
RUN go get github.com/golang-migrate/migrate/v4/source/file
## Add this go mod download command to pull in any dependencies
RUN go mod download
## Our project will now successfully build with the necessary go libraries included.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go 
CMD ["/restapi/main"]
EXPOSE 3000/tcp
