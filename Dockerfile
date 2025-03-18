# build stage
FROM golang:1.24-alpine AS build

# set working directory
WORKDIR /app

# copy source code
COPY . .

# install dependencies
RUN go mod download

# build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/harajuku ./cmd/main.go

# final stage
FROM alpine:latest AS final
LABEL maintainer="RamMaths"

# set working directory
WORKDIR /app

# copy binary
COPY --from=build /app/bin/harajuku ./

EXPOSE 8080

ENTRYPOINT [ "./harajuku" ]
