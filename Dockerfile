FROM golang:1.26 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/app

FROM gcr.io/distroless/static-debian12:nonroot

ENV PORT=5000
ENV DATA_DIR=/data
WORKDIR /app
COPY --from=build /out/app /app/app
EXPOSE 5000
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
