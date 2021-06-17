# build stage
FROM golang:alpine AS build-env
RUN apk --no-cache add build-base git bzr mercurial gcc
ADD . /upload_service
RUN cd /upload_service && go build -o goapp

# final stage
FROM alpine
WORKDIR /upload_service
COPY --from=build-env /upload_service/goapp /upload_service/
ENTRYPOINT ./goapp
