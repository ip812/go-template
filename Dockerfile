FROM 678468774710.dkr.ecr.eu-central-1.amazonaws.com/ip812/go-ssr:20250321 AS build-stage
WORKDIR /app
COPY . .
RUN ./bin/tailwindcss-extra-linux-x64 -i ./static/css/input.css -o ./static/css/output.css --minify
RUN sqlc generate
RUN templ generate
RUN go mod tidy
RUN go mod vendor
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/main .

FROM gcr.io/distroless/base-debian12 AS run-stage
COPY --from=build-stage /app/sql/migrations /sql/migrations
COPY --from=build-stage /app/bin/main /bin/main
CMD ["/bin/main"]
