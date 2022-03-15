FROM golang:1.17.2-buster

WORKDIR /app

COPY src /app
RUN chmod +x /app/run-tests.sh

ENTRYPOINT ["./run-tests.sh"]
