FROM golang:1.22-alpine as builder
LABEL authors="Javad Rajabzadeh"

RUN mkdir /app
WORKDIR /app

COPY . /app
RUN make build

#FROM ubuntu:22.04

#RUN apk --no-cache add ca-certificates tzdata
#RUN mkdir /app
#COPY --from=builder /app/main /app
RUN chmod +x /app/main
CMD ["./main", "-c", "/etc/artogenia/config.yaml", "run"]