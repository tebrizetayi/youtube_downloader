# Use an official Golang runtime as a parent image
FROM golang:latest

RUN mkdir /go/src/app

# Set the working directory to /go/src/app
WORKDIR /go/src/app
RUN apt-get update && apt-get install -y ffmpeg
RUN go install github.com/iawia002/lux@latest

# Copy the current directory contents into the container at /go/src/app
COPY . /go/src/app

RUN go get github.com/githubnemo/CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon

# Download any needed dependencies specified in go.mod
RUN go mod download



#RUN go test ./... -v
# Build the application

#RUN go build -o main ./


# Expose port 8080

EXPOSE 7070


# Run the application
#CMD ["./main"]

ENTRYPOINT CompileDaemon --build="go build -buildvcs=false -o main ./" --command=./main