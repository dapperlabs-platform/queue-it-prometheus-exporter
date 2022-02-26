FROM golang:1.17-bullseye as build
WORKDIR /home/app
ADD . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o /queue-it-prometheus-exporter

FROM gcr.io/distroless/base-debian11
COPY --from=build /queue-it-prometheus-exporter /
CMD [ "/queue-it-prometheus-exporter" ]
