FROM golang:1.22-alpine as builder
LABEL authors="Javad Rajabzadeh"

RUN mkdir /app
WORKDIR /app
COPY . /app
RUN go build -o build/meilibridge cmd/meilibridge/main.go

FROM alpine
RUN apk --no-cache add ca-certificates tzdata
RUN mkdir /app
COPY --from=builder /app/build/meilibridge /usr/local/bin
RUN chmod +x /usr/local/bin
CMD ["meilibridge", "-c", "/etc/meilibridge/config.yaml"]