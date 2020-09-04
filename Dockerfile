FROM golang
RUN mkdir -p /app/cert
ADD . /app
COPY /cert/cert.pem /app/cert/cert.pem
COPY /cert/privkey.pem /app/cert/privkey.pem

WORKDIR /app

RUN go build -o main .
CMD ["/app/main"]

EXPOSE 443