# Start with a base image that has Go pre-installed
FROM golang:1.20 as builder

# Install required tools
RUN apt-get update && apt-get install -y \
    vsftpd \
    nano \
    && rm -rf /var/lib/apt/lists/*

# Set the working directory
WORKDIR /app

# Copy Go source code into the container
COPY watcher/ /app/

# Install GoLand (Optional: if you're running locally and want an IDE, configure it outside the container)
# RUN wget -qO- https://download.jetbrains.com/go/goland-<version>.tar.gz | tar xvz

# Build Go application

RUN go mod init ftp-watcher && go mod tidy && go build -o watcher main.go

# Prepare final image
FROM debian:bullseye-slim

# Install required packages
RUN apt-get update && apt-get install -y \
    vsftpd \
    && rm -rf /var/lib/apt/lists/*

# Copy FTP configuration
COPY vsftpd.conf /etc/vsftpd.conf

# Create FTP user and folder
RUN useradd -m ftpuser && echo "ftpuser:password" | chpasswd && \
    mkdir -p /home/ftpuser/ftp && chown -R ftpuser:ftpuser /home/ftpuser/ftp

# Expose FTP ports
EXPOSE 21 21000-21010

# Copy the built Go binary
COPY --from=builder /app/watcher /usr/local/bin/watcher

# Run FTP server and watcher
CMD ["/bin/bash", "-c", "vsftpd /etc/vsftpd.conf & watcher"]
