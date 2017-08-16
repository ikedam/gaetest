# https://cloud.google.com/appengine/docs/standard/go/download?hl=ja
FROM python:2.7-alpine

ENV GOAPP_VERSION 1.9.57
ENV GOAPP_URL https://storage.googleapis.com/appengine-sdks/featured/go_appengine_sdk_linux_amd64-${GOAPP_VERSION}.zip

ADD test.go .

RUN echo "Installing build tools..."
RUN apk add --update build-base \
	&& rm -rf /var/cache/apk/*

RUN echo "Installing git..."
RUN apk add --update git \
	&& rm -rf /var/cache/apk/*

RUN echo "Installing curl..."
RUN apk add --update curl \
	&& rm -rf /var/cache/apk/*

RUN echo "Installing go..."
RUN apk add --update go \
	&& rm -rf /var/cache/apk/*

RUN echo "Installing go_appengine ${GOAPP_VERSION}..."
RUN curl -o go_appengine.zip ${GOAPP_URL} \
	&& unzip -qd /usr/local go_appengine.zip \
	&& rm -f go_appengine.zip

RUN mkdir /go

ENV GOPATH /go
ENV PATH ${PATH}:/usr/local/go_appengine:${GOPATH}/bin

RUN echo "Installing glide..."
RUN go get github.com/Masterminds/glide

CMD ["go", "run", "test.go"]
