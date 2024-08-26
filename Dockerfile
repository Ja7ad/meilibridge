FROM golang:1.23-alpine AS builder
LABEL authors="Javad Rajabzadeh"

RUN apk add make

RUN mkdir /app
WORKDIR /app
COPY . /app
RUN make build

FROM alpine
RUN apk --no-cache add ca-certificates tzdata
RUN mkdir /etc/meilibridge/
COPY --from=builder /app/build/meilibridge /usr/local/bin
RUN chmod +x /usr/local/bin
CMD ["meilibridge", "sync", "start", "-c", "/etc/meilibridge/config.yml"]