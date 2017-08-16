# https://cloud.google.com/appengine/docs/standard/go/download?hl=ja
FROM alpine:latest

ENV GOAPP_VERSION 1.9.57
ENV GOAPP_URL https://storage.googleapis.com/appengine-sdks/featured/go_appengine_sdk_linux_amd64-${GOAPP_VERSION}.zip

ADD test.go .

RUN echo "Installing packages..."
RUN apk add --update build-base python2 git curl go glide \
	&& rm -rf /var/cache/apk/*

RUN echo "Installing go_appengine ${GOAPP_VERSION}..."
RUN curl -o go_appengine.zip ${GOAPP_URL} \
	&& unzip -qd /usr/local go_appengine.zip \
	&& rm -f go_appengine.zip

RUN mkdir /go

ENV GOPATH /go
ENV PATH ${PATH}:/usr/local/go_appengine:${GOPATH}/bin

CMD ["go", "run", "test.go"]
