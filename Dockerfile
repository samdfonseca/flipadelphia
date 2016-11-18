FROM golang:1.7

COPY . /go/src/app
RUN mkdir -p /go/src/github.com/samdfonseca/ && \
    mv /go/src/app/ /go/src/github.com/samdfonseca/flipadelphia && \
    cd /go/src/github.com/samdfonseca/flipadelphia && \
    go get -u github.com/tools/godep && \
    mkdir -p $HOME/.flipadelphia && \
    cp config/config.example.json $HOME/.flipadelphia/config.json && \
    go get -u -v && \
    make build && \
    make install

ENV FLIPADELPHIA_ENV=noauth
EXPOSE 3006
ENTRYPOINT ["flipadelphia", "&"]
