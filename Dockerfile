FROM golang:1.23-bookworm

# 安装必要的库
RUN apt-get update && apt-get install -y \
    build-essential \
    ca-certificates \
    libstdc++6 \
    libc6 \
    libpthread-stubs0-dev \
    tzdata \
    tini \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /root/wjd/go-demo1

# 安装特定版本的 Go
RUN wget https://golang.google.cn/dl/go1.23.1.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.23.1.linux-amd64.tar.gz && \
    rm go1.23.1.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"

RUN go version

# 复制所有代码到容器中
COPY . .

# 使用 tini 启动应用程序
#ENTRYPOINT ["/usr/bin/tini", "--"]

CMD ["tail", "-f", "/dev/null"]
#CMD ["bash", "-c", "go mod tidy && go run /root/wjd/go-demo1/eth-scan-erigon/daemon/main.go"]
#CMD ["bash", "-c", "go mod tidy && go build -o /root/wjd/myapp /root/wjd/go-demo1/eth-scan-erigon/scan/main.go && /root/wjd/myapp"]

