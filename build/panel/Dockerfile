FROM golang:1.17 AS dependencies
WORKDIR /go/src/panel
COPY go.mod .
COPY go.sum .
RUN go mod download

FROM dependencies AS build
COPY . /go/src/panel
WORKDIR /go/src/panel
RUN make build-panel

FROM debian:buster-slim
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/src/panel/bin/* /app/
COPY --from=build /go/src/panel/scripts/wait-for-it.sh /app/
RUN chmod +x /app/*
EXPOSE 8080/tcp
CMD /app/panel serve
