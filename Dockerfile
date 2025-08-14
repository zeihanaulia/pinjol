FROM golang:1.24.5-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${GIT_SHA:-dev} -X main.buildTime=$(date -u +%FT%TZ)" -o /out/app .

FROM gcr.io/distroless/base-debian12
ENV PORT=8080 APP_ENV=prod
EXPOSE 8080
COPY --from=build /out/app /app
USER 65532:65532
ENTRYPOINT ["/app"]
