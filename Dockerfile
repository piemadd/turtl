FROM golang
RUN mkdir -p /app/cert
ADD . /app
COPY /etc/letsencrypt/live/api.turtl.cloud/cert.pem /app/cert/cert.pem
COPY /etc/letsencrypt/live/api.turtl.cloud/privkey.pem /app/cert/privkey.pem

WORKDIR /app

RUN go build -o main .
CMD ["/app/main"]

EXPOSE 80