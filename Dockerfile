# Use an official Golang runtime as a parent image
FROM golang:latest

# Set the working directory to /go/src/app
WORKDIR /go/src/app
RUN apt-get update && apt-get install -y ffmpeg
RUN go install github.com/iawia002/lux@latest

# Copy the current directory contents into the container at /go/src/app
COPY . /go/src/app

# Download any needed dependencies specified in go.mod
RUN go mod download

#
#RUN go test ./... -v
# Build the application

RUN go build -o main .

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./main"]
