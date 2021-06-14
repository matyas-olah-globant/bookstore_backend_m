FROM golang:alpine
WORKDIR /app
ADD go.mod go.sum main.go books.json setup.sql /app/
RUN go mod tidy

EXPOSE 1151

#RUN go build -o backend .
#CMD ["/app/backend"]
