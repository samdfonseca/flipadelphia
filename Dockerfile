FROM golang:1.6.0

COPY . /go/src/app
RUN mkdir -p /go/src/github.com/samdfonseca/ && \
    mkdir -p /go/src/app && \
    mv /go/src/app/ /go/src/github.com/samdfonseca/flipadelphia && \
    cd /go/src/github.com/samdfonseca/flipadelphia && \
    go build github.com/samdfonseca/flipadelphia && \
    mkdir -p $HOME/.flipadelphia && \
    cp config/config.example.json $HOME/.flipadelphia/config.json && \
    make install

ENV FLIPADELPHIA_ENV=development
EXPOSE 3006
ENTRYPOINT ["flipadelphia", "serve"]

