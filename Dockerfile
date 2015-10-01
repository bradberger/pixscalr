# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
<<<<<<< HEAD
ADD . /go/src/github.com/golang/example/outyet
=======
ADD . /go/src/github.com/bradberger/pixscalr
>>>>>>> 906b1e6c7012bd8558d5bd32f4df827cf01cff2a

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go install github.com/bradberger/pixscalr

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/pixscalr

# Document that the service listens on port 3000.
EXPOSE 3000
