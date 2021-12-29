FROM golang:1.17 AS dependencies
WORKDIR /go/src/grader
COPY go.mod .
COPY go.sum .
RUN go mod download

FROM dependencies AS build
COPY . /go/src/grader
WORKDIR /go/src/grader
RUN make build-grader

FROM debian:buster-slim
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/src/grader/bin/* /app/
COPY --from=build /go/src/grader/scripts/wait-for-it.sh /app/
RUN chmod +x /app/*
EXPOSE 8080/tcp
CMD /app/grader serve
