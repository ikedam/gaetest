# https://cloud.google.com/appengine/docs/standard/go/download?hl=ja

# alpine は glibc がないため、 nodejs の chrome-webdriver が動作しない
FROM ubuntu:latest

## Packages provided by the distribution
# build-essential is especially heavy.
RUN apt-get -y update \
	&& apt-get install --no-install-recommends -y \
	build-essential

# Packages for go (go_appengine, glide)
RUN apt-get -y update \
	&& apt-get install --no-install-recommends -y \
	build-essential \
	python2.7 \
	git \
	golang \
	&& update-alternatives --install /usr/bin/python python /usr/bin/python2.7 1

# Packages for e2e
RUN apt-get -y update \
	&& apt-get install --no-install-recommends -y \
	chromium-browser \
	libgconf-2-4

# utilities, add-apt-repository, support https://... by apt-get
RUN apt-get -y update \
	&& apt-get install --no-install-recommends -y \
	curl \
	unzip \
	software-properties-common \
	apt-transport-https

## external packages
# nodejs & yarn
ARG NODEJS_VERSION="6.11.2"
ARG NODEJS_URL="https://nodejs.org/dist/v${NODEJS_VERSION}/node-v${NODEJS_VERSION}-linux-x64.tar.xz"

RUN curl -o node.tar.xz ${NODEJS_URL} \
	&& mkdir /usr/local/nodejs \
	&& tar xJ -C /usr/local/nodejs -f node.tar.xz --strip 1 \
	&& rm -f node.tar.xz

RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -
RUN echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list
RUN apt-get -y update \
	&& apt-get install --no-install-recommends -y \
	yarn

# appengine
RUN add-apt-repository ppa:masterminds/glide \
	&& apt-get -y update \
	&& apt-get install --no-install-recommends -y \
	glide

ARG GOAPP_VERSION="1.9.58"
ARG GOAPP_URL="https://storage.googleapis.com/appengine-sdks/featured/go_appengine_sdk_linux_amd64-${GOAPP_VERSION}.zip"

RUN curl -o go_appengine.zip ${GOAPP_URL} \
	&& unzip -qd /usr/local go_appengine.zip \
	&& rm -f go_appengine.zip

RUN rm /var/lib/apt/lists/*_*

RUN mkdir /go
ENV GOPATH /go
ENV PATH ${PATH}:/usr/local/go_appengine:/usr/local/nodejs/bin:${GOPATH}/bin
ENV CHROME_BIN /usr/bin/chromium-browser

# glide, yarn で使うキャッシュディレクトリ
# docker volume create cache
# docker run -v cache:/cache ...
# とすればライブラリのセットアップ処理を高速化できます。
RUN mkdir /cache

ENV GLIDE_HOME /cache/glide
ENV YARN_CACHE_FOLDER /cache/yarn

# ログ出力
# docker run -v `pwd`:/log ...
# とすればログをカレントディレクトリに保存できます。
RUN mkdir /log

ADD . /go/src/github.com/ikedam/gaetest

ENTRYPOINT ["go", "run", "/go/src/github.com/ikedam/gaetest/tool/test.go", "--docker", "--log=/log"]
#CMD ["go", "run", "/go/src/github.com/ikedam/gaetest/tool/test.go", "--docker", "--log=/log"]
