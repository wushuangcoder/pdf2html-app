FROM registry.cn-beijing.aliyuncs.com/usy/pdf2htmlex:0.18.8.rc2-master-20200820-ubuntu-20.04-x86_64

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 安装wget
RUN apt-get update && apt-get install -y wget && rm -rf /var/lib/apt/lists/*

# 安装Go 1.24
RUN wget -q https://go.dev/dl/go1.24.0.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz && \
    rm go1.24.0.linux-amd64.tar.gz

# 设置Go环境变量
ENV PATH="/usr/local/go/bin:$PATH"

# 下载依赖，使用国内代理加速
RUN GOPROXY=https://goproxy.cn,direct go mod download

# 复制源代码
COPY . .

# 构建应用
RUN go build -o pdf2html-app .

# 暴露端口
EXPOSE 8080

# 覆盖默认的ENTRYPOINT
ENTRYPOINT []

# 运行应用
CMD ["./pdf2html-app"]