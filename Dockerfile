FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk update && apk add git
# RUN git clone --depth 1 https://github.com/ponyo877/folks-ui.git /src
COPY . .
RUN go mod download
RUN go build -o /matchmaking main.go

FROM alpine:latest
WORKDIR /
COPY --from=build /matchmaking /matchmaking
ENTRYPOINT ["/matchmaking"]