# Use an official Golang runtime as a parent image
FROM golang:latest

# Create and set the working directory
#RUN mkdir /go/src/app
WORKDIR /go/src/app

# Install ffmpeg and setup Python virtual environment
RUN apt-get update && apt-get install -y ffmpeg python3 python3-pip python3-venv

# Setup virtual environment
RUN python3 -m venv /opt/venv
# Activate virtual environment and install yt-dlp
RUN . /opt/venv/bin/activate && pip install yt-dlp

ENV PATH="/opt/venv/bin:$PATH"

# Copy the current directory contents into the container
COPY . .

# Install CompileDaemon for live reloading in development
RUN go install github.com/githubnemo/CompileDaemon@latest

# Download any needed dependencies specified in go.mod
RUN go mod download

# Expose the application port
EXPOSE 6060

# Set the entry point to use CompileDaemon for live reloading
ENTRYPOINT CompileDaemon --build="go build -buildvcs=false -o main ./" --command=./main
