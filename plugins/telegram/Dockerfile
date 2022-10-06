FROM golang:latest

# COPY . /tmp/build/tbcpusher
WORKDIR /tmp/build
RUN git clone https://github.com/turbitcat/tbcpusher.git
WORKDIR /tmp/build/tbcpusher/plugins/telegram
RUN go build
RUN cp telegram /root
RUN rm -rf /tmp/build
WORKDIR /root

EXPOSE 8000

ENTRYPOINT ["/root/telegram"]