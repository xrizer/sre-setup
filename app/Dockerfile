FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /booking-api .

FROM gcr.io/distroless/static-debian12
COPY --from=build /booking-api /booking-api
EXPOSE 8080
USER nonroot
ENTRYPOINT ["/booking-api"]
