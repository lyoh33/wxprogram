# 使用官方 Go 镜像作为基础镜像
FROM golang:1.23.5

# 设置工作目录
WORKDIR /app

# 将 go.mod 和 go.sum 复制到工作目录
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 将项目的所有文件复制到工作目录
COPY . .

# 构建可执行文件
RUN go build -o main .

# 暴露服务端口
EXPOSE 8080

# 运行应用程序
CMD ["./main"]