# https://cloud.google.com/appengine/docs/standard/go/download?hl=ja

FROM python:2.7.15-alpine3.8

# Packages for go (go_appengine, glide)
RUN apk add --no-cache build-base git curl go glide chromium

# Packages for e2e
RUN apk add --no-cache yarn chromium chromium-chromedriver

## external packages
# nodejs & yarn
ARG NODEJS_VERSION="8.11.4"
ARG NODEJS_URL="https://nodejs.org/dist/v${NODEJS_VERSION}/node-v${NODEJS_VERSION}-linux-x64.tar.xz"

RUN curl -o node.tar.xz ${NODEJS_URL} \
	&& mkdir /usr/local/nodejs \
	&& tar xJ -C /usr/local/nodejs -f node.tar.xz --strip 1 \
	&& rm -f node.tar.xz

ARG GOAPP_VERSION="1.9.68"
ARG GOAPP_URL="https://storage.googleapis.com/appengine-sdks/featured/go_appengine_sdk_linux_amd64-${GOAPP_VERSION}.zip"

RUN curl -o go_appengine.zip ${GOAPP_URL} \
	&& unzip -qd /usr/local go_appengine.zip \
	&& rm -f go_appengine.zip

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
