FROM golang:1.17-bullseye as build
WORKDIR /home/app
ADD . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o ./queue-it-prometheus-exporter

FROM gcr.io/distroless/base-debian11
COPY --from=build /home/app/queue-it-prometheus-exporter /queue-it-prometheus-exporter
ENTRYPOINT [ "/queue-it-prometheus-exporter" ]
