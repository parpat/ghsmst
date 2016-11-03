#FROM golang:onbuild


FROM golang:latest
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o main .
CMD ["/app/main"]

#Service listens on port 7575.
EXPOSE 7575
