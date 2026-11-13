FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build
ARG TARGETOS TARGETARCH

WORKDIR /app
COPY . .
ENV CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH
RUN go build -mod=vendor -o /budget-ingest

FROM alpine

COPY --from=build /budget-ingest /budget-ingest

CMD /budget-ingest
