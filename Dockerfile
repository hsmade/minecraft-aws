FROM golang:1.19-alpine as build
ARG app
RUN apk add --no-cache ca-certificates
RUN adduser -S -u 1000 user
COPY . /app/
WORKDIR /app
RUN go mod vendor
RUN CGO_ENABLED=0 go build -o $app app/$app/*.go

FROM scratch
ARG app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/shadow /etc/
COPY --from=build /app/$app /app
USER 1000
ENTRYPOINT ["/app"]