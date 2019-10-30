# build stage
FROM golang:alpine AS build-env
RUN apk --no-cache add build-base git bzr mercurial gcc
ADD . /rebateton
RUN cd /rebateton && go build -o goapp

# final stage
FROM alpine
WORKDIR /rebateton
COPY --from=build-env /rebateton/goapp /rebateton/
ENTRYPOINT ./goapp
