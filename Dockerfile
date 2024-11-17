FROM golang:latest

# 安装编译所需的依赖
RUN apt-get update && apt-get install -y \
    build-essential \
    libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY . .

RUN go mod tidy && \
    go build -o oucsearch

CMD ["./oucsearch"]