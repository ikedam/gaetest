# https://cloud.google.com/appengine/docs/standard/go/download?hl=ja

# alpine ÇÕ glibc Ç™Ç»Ç¢ÇΩÇﬂÅA nodejs ÇÃ chrome-webdriver Ç™ìÆçÏÇµÇ»Ç¢
FROM ubuntu:latest

ENV GOAPP_VERSION 1.9.57
ENV GOAPP_URL https://storage.googleapis.com/appengine-sdks/featured/go_appengine_sdk_linux_amd64-${GOAPP_VERSION}.zip
ENV NODEJS_VERSION 6.11.2
ENV NODEJS_URL https://nodejs.org/dist/v${NODEJS_VERSION}/node-v${NODEJS_VERSION}-linux-x64.tar.xz

RUN echo "Installing packages..."

# for add-apt-repository, https://...
RUN apt-get -y update \
	&& apt-get install --no-install-recommends -y \
	software-properties-common \
	apt-transport-https

RUN echo "Preparing go-appengine..."
RUN add-apt-repository ppa:masterminds/glide \
	&& apt-get -y update \
	&& apt-get install --no-install-recommends -y \
	build-essential \
	python2.7 \
	git \
	curl \
	golang \
	glide \
	unzip \
	&& update-alternatives --install /usr/bin/python python /usr/bin/python2.7 1

RUN curl -o go_appengine.zip ${GOAPP_URL} \
	&& unzip -qd /usr/local go_appengine.zip \
	&& rm -f go_appengine.zip


RUN echo "Preparing angular..."
RUN curl -o node.tar.xz ${NODEJS_URL} \
	&& mkdir /usr/local/nodejs \
	&& tar xJ -C /usr/local/nodejs -f node.tar.xz --strip 1 \
	&& rm -f node.tar.xz

RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -
RUN echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list
RUN apt-get -y update \
	&& apt-get install --no-install-recommends -y \
	yarn

RUN echo "Preparing e2e..."

RUN apt-get -y update \
	&& apt-get install --no-install-recommends -y \
	chromium-browser \
	libgconf-2-4

RUN rm /var/lib/apt/lists/*_*

RUN mkdir /go
ENV GOPATH /go
ENV PATH ${PATH}:/usr/local/go_appengine:/usr/local/nodejs/bin:${GOPATH}/bin

ADD test.go .

CMD ["go", "run", "test.go", "--docker"]
