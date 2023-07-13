####### Agola binary
FROM sorintlab/agola:v0.8.0 AS agola

####### Build environment
FROM golang:1.20-buster AS builder
WORKDIR /app
ENV GO111MODULE=on
COPY . .
RUN go build

####### Build the final image
FROM debian:buster
WORKDIR /
COPY --from=agola /bin/agola /bin/
COPY --from=builder /app/local-setup /bin/
ENTRYPOINT ["/bin/local-setup"]
