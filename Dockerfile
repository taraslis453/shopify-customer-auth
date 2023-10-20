# build stage
FROM golang:1.19 AS build

ARG POSTGRESQL_HOST
ARG POSTGRESQL_USER
ARG POSTGRESQL_PASSWORD
ARG POSTGRESQL_DATABASE

# set env variables
ENV GOOS=linux \
  GOARCH=amd64 \
  CGO_ENABLED=0 

ENV POSTGRESQL_HOST $POSTGRESQL_HOST
ENV POSTGRESQL_USER $POSTGRESQL_USER
ENV POSTGRESQL_PASSWORD $POSTGRESQL_PASSWORD
ENV POSTGRESQL_DATABASE $POSTGRESQL_DATABASE

# copy and download mods
WORKDIR /srv/app/pkg
COPY go.mod .
COPY go.sum .
RUN go mod download

# copy and build code
COPY . .
RUN go build -o /srv/app/app cmd/main.go

# run stage
FROM golang:1.19-alpine as run
ENV GIN_MODE release
EXPOSE 8080

# copy binary
WORKDIR /srv
RUN mkdir -p /srv
COPY --from=build /srv/app/app /srv/app

# run binary
CMD ["/srv/app"]
