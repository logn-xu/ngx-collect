# NGX-Collect

NGX-Collect 是一个用于从多台服务器收集 Nginx 配置文件的工具。它使用 SSH 连接到远程服务器，并通过 SFTP 协议下载 Nginx 配置文件到本地指定目录。

## 功能特点

- 支持从多台服务器收集 Nginx 配置
- 支持 SSH 密钥和密码认证方式
- 支持递归下载目录
- 自动组织下载的文件，按服务器分类存储
- 简单易用的 YAML 配置文件

## 安装

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/logn/ngx-collect.git
cd ngx-collect

# 构建项目
make build
```

构建完成后，会在项目根目录生成 `ngx-collect` 可执行文件。

## 配置

NGX-Collect 使用 YAML 格式的配置文件。默认配置文件位于 `config/config.yaml`。

### 配置文件示例

```yaml
machines:
  - alias: "webserver_01"  # 给机器起个别名，方便识别
    host: "10.182.6.19"
    port: 22             # SSH 端口，默认22
    user: "nginx"
    auth_method: "key"   # 'key' 或 'password'
    key_path: "~/.ssh/id_rsa" # 如果 auth_method 是 'key'
    # password: "your_password" # 如果 auth_method 是 'password'
    remote_paths: # 需要从这台机器拉取的 Nginx 配置文件路径列表
      - "/app/nginx/conf/"
      - "/app/nginx/sbin/"
    local_destination: "./data/" # 这台机器的配置下载到本地的路径
```

### 配置项说明

- `machines`: 服务器列表
  - `alias`: 服务器别名，用于在本地文件组织中标识服务器
  - `host`: 服务器 IP 地址或域名
  - `port`: SSH 端口，默认 22
  - `user`: SSH 用户名
  - `auth_method`: 认证方式，支持 `key` 或 `password`
  - `key_path`: SSH 私钥路径，当 `auth_method` 为 `key` 时使用
  - `password`: SSH 密码，当 `auth_method` 为 `password` 时使用
  - `remote_paths`: 需要收集的远程路径列表
  - `local_destination`: 本地保存路径

## 使用方法

```bash
# 使用默认配置文件
./ngx-collect

# 使用指定配置文件
./ngx-collect --config /path/to/your/config.yaml
```

## 文件组织

收集的文件将按以下结构保存：

```
{local_destination}/{alias}/{host}/{remote_path}
```

例如，如果配置为：

```yaml
alias: "webserver_01"
host: "10.182.6.19"
local_destination: "./data/"
remote_paths:
  - "/app/nginx/conf/"
```

则文件将保存在：

```
./data/webserver_01/10.182.6.19/app/nginx/conf/
```

## 许可证

[MIT License](LICENSE)