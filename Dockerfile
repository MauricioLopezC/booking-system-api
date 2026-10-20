FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/reservation-api ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/seed ./cmd/seed

FROM alpine:3.22
# tzdata: needed for BUSINESS_TIMEZONE (time.LoadLocation)
RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -H -u 10001 app
COPY --from=build /out/ /usr/local/bin/
USER app
EXPOSE 3000
CMD ["reservation-api"]
