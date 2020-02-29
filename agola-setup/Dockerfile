#######
####### Build the backend
#######

# base build image
FROM golang:1.13-buster AS builder

WORKDIR /app

# use go modules
ENV GO111MODULE=on

# only copy go.mod and go.sum
COPY . .

RUN go build

#######
####### Build the final image
#######
FROM debian:buster

WORKDIR /

COPY --from=builder /app/agola-setup /bin/

ENTRYPOINT ["/bin/agola-setup"]

