# 概述

Welcome to the paas-operator wiki!

## quick start

1. 在 build/apiserver/docker/Dockerfile 中替换 ${BASE_IMAGE} 为你想要的基础镜像，比如： centos:7
2. 给 build/apiserver/docker/startup.sh 中 ${OPERATOR_IP} 一个恰当的值，如果你使用 Nodeport 方式暴露服务，那么这里对应你的宿主机 ip
3.

## 资源分类

考虑到很多场景下数据库和中间件的差异性，可能多数企业会分别对待这两类软件，包括维护人员、部署模式等等。所以一开始我们就将 Database 和 Middleware 分开处理。资源分为：

- Database
- Middleware

## 资源创建

### 数据库创建

#### request

| method | url                            | desc           |
| ------ | ------------------------------ | -------------- |
| POST   | /apis/v1alpha1/database/create | 创建数据库实例 |

#### body

```json
{
  "name": "mysql-5.7-192.168.19.100",
  "host": [
    {
      "ip": "192.168.19.100",
      "// auth": "if only one username & password set, use it anywhere; if two, use the first to run the agent and another to ssh connect.",
      "auth": [
        {
          "username": "root",
          "password": "root123"
        },
        {
          "username": "sshuser",
          "password": "ssh123"
        }
      ]
    }
  ],
  "app": {
    "repo_url": "http://192.168.19.200:123/ftp/software/mysql/5.7/",
    "install": "install.sh",
    "start": "start.sh",
    "stop": "stop.sh",
    "restart": "restart.sh",
	"check": "check.sh",
    "uninstall": "uninstall.sh",
    "package": "mysql-5.7.tar.gz",
    "metadata": {
      "APP_USER": "mysql",
      "APP_PASSWD": "MYSQL123",
    },
    "// status": "if set status at create time, program will use it for ease of testing.",
    "status": {
      "expect": "not-installed",
      "realtime": "not-installed"
    }
  }
}
```

#### response

statuscode: 201

body:

```json
{
  "name": "mysql-5.7-192.168.19.100",
  "// type":"generate by po",
  "type": "database",
  "host": [
    {
      "ip": "192.168.19.100",
      "// auth": "if only one username & password set, use it anywhere; if two, use the first to run the agent and another to ssh connect.",
      "auth": [
        {
          "username": "root",
          "password": "root123"
        },
        {
          "username": "sshuser",
          "password": "ssh123"
        }
      ]
    }
  ],
  "app": {
    "repo_url": "http://192.168.19.200:123/ftp/software/mysql/5.7/",
    "install": "install.sh",
    "start": "start.sh",
    "stop": "stop.sh",
    "restart": "restart.sh",
	"check": "check.sh",
    "uninstall": "uninstall.sh",
    "package": "mysql-5.7.tar.gz",
    "metadata": {
      "APP_USER": "mysql",
      "APP_PASSWD": "MYSQL123",
      "// CreateAt": "it's set by po.",
      "CreateAt": "2019-05-29 10:33:55"
    },
    "// status": "if set status at create time, program will use it for ease of testing.",
    "status": {
      "expect": "not-installed",
      "realtime": "not-installed"
    }
  }
}
```

### 中间件创建

#### request

| method | url                              | desc           |
| ------ | -------------------------------- | -------------- |
| POST   | /apis/v1alpha1/middleware/create | 创建中间件实例 |

#### body

```json
{
  "name": "nginx-1.16-192.168.19.100",
  "host": [
    {
      "ip": "192.168.19.100",
      "// auth": "if only one username & password set, use it anywhere; if two, use the first to run the agent and another to ssh connect.",
      "auth": [
        {
          "username": "root",
          "password": "root123"
        },
        {
          "username": "sshuser",
          "password": "ssh123"
        }
      ]
    }
  ],
  "app": {
    "repo_url": "http://192.168.19.200:123/ftp/software/nginx/1.16/",
    "install": "install.sh",
    "start": "start.sh",
    "stop": "stop.sh",
    "restart": "restart.sh",
	"check": "check.sh",
    "uninstall": "uninstall.sh",
    "package": "nginx-1.16.tar.gz",
    "metadata": {
      "APP_USER": "nginx",
      "APP_PASSWD": "NGINX123",
    },
    "// status": "if set status at create time, program will use it for ease of testing.",
    "status": {
      "expect": "not-installed",
      "realtime": "not-installed"
    }
  }
}
```

#### response

statuscode: 201

body:

```json
{
  "name": "nginx-1.16-192.168.19.100",
  "// type":"generate by po",
  "type": "middleware",
  "host": [
    {
      "ip": "192.168.19.100",
      "// auth": "if only one username & password set, use it anywhere; if two, use the first to run the agent and another to ssh connect.",
      "auth": [
        {
          "username": "root",
          "password": "root123"
        },
        {
          "username": "sshuser",
          "password": "ssh123"
        }
      ]
    }
  ],
  "app": {
    "repo_url": "http://192.168.19.200:123/ftp/software/nginx/1.16/",
    "install": "install.sh",
    "start": "start.sh",
    "stop": "stop.sh",
    "restart": "restart.sh",
	"check": "check.sh",
    "uninstall": "uninstall.sh",
    "package": "nginx-1.16.tar.gz",
    "metadata": {
      "APP_USER": "nginx",
      "APP_PASSWD": "NGINX123",
      "// CreateAt": "it's set by program.",
      "CreateAt": "2019-05-29 10:33:55"
    },
    "// status": "if set status at create time, program will use it for ease of testing.",
    "status": {
      "expect": "not-installed",
      "realtime": "not-installed"
    }
  }
}
```

## 资源状态查询

### 数据库状态

#### request

| method | url                                     | desc           |
| ------ | --------------------------------------- | -------------- |
| GET    | /apis/v1alpha1/database/{d_name}/status | 数据库实例状态 |

#### response

```json
{
	"name": "mysql-5.6-single-192.168.19.100-xxx",
	"status": {
		"expect": "not-installed", # 期望的状态
		"realtime": "not-installed" # 实时状态，由agent回写
	}
}
```

### 中间件状态

#### request

| method | url                                       | desc           |
| ------ | ----------------------------------------- | -------------- |
| GET    | /apis/v1alpha1/middleware/{a_name}/status | 数据库实例状态 |

#### response

```json
{
	"name": "nginx-1.16-single-192.168.19.100-xxx",
	"status": {
		"expect": "not-installed", # 期望的状态
		"realtime": "not-installed" # 实时状态，由agent回写
	}
}
```

## 资源状态修改

### 数据库状态修改

#### request

| method | url                                          | desc                                                         |
| ------ | -------------------------------------------- | ------------------------------------------------------------ |
| put    | apis/v1alpha1/database{a_name}/running       | 期望应用运行起来（对应安装和启动）                           |
| put    | apis/v1alpha1/database{a_name}/stopped       | 期望应用停止运行                                             |
| put    | apis/v1alpha1/database{a_name}/not-installed | 期望应用处于未安装状态（比如卸载操作）                       |
| put    | apis/v1alpha1/database{a_name}/restart       | 重启操作，期望状态不变，实际状态 -> running -> stopping -> stopped -> starting -> running |

#### response

statuscode: 202

#### body

```json
{
	"name": "mysql-5.6-single-192.168.19.100",
	"status": {
		"expect": "starting", # 期望的状态
		"realtime": "not-installed" # 实时状态，由agent回写
	}
}
```

### 中间件状态修改

#### request

| method | url                                            | desc                                                         |
| ------ | ---------------------------------------------- | ------------------------------------------------------------ |
| put    | apis/v1alpha1/middleware{a_name}/running       | 期望应用运行起来（对应安装和启动）                           |
| put    | apis/v1alpha1/middleware{a_name}/stopped       | 期望应用停止运行                                             |
| put    | apis/v1alpha1/middleware{a_name}/not-installed | 期望应用处于未安装状态（比如卸载操作）                       |
| put    | apis/v1alpha1/middleware{a_name}/restart       | 重启操作，期望状态不变，实际状态 -> running -> stopping -> stopped -> starting -> running |

#### response

statuscode: 202

#### body

```json
{
	"name": "nginx-1.16-single-192.168.19.100",
	"status": {
		"expect": "starting", # 期望的状态
		"realtime": "not-installed" # 实时状态，由agent回写
	}
}
```

## 应用状态集

### 终态

| 稳定状态值                | 期望状态可能值？ | 实际状态可能值？ |
| ------------------------- | ---------------- | ---------------- |
| not-installed             | true             | true             |
| running                   | true             | true             |
| stopped                   | true             | true             |
| failed                    |                  | true             |
| unknow                    |                  | true             |
| restart(特例，是一个动作) | true             |                  |

### 中间态

| 中间状态值 | 含义   |
| ---------- | ------ |
| starting   | 启动中 |
| installing | 安装中 |
| stopping   | 停止中 |
| restarting | 重启中 |

## 资源删除

### 数据库删除

#### request

| method | url                              | desc           |
| ------ | -------------------------------- | -------------- |
| DELETE | /apis/v1alpha1/database/{a_name} | 删除数据库实例 |

#### response

200 ok

- 删除成功：返回200，附带被删除的实例名；
- 不存在：也返回200，附带提示信息；
- 有错误：返回相应错误码和错误信息。

### 中间件删除

#### request

| method | url                                | desc           |
| ------ | ---------------------------------- | -------------- |
| DELETE | /apis/v1alpha1/middleware/{a_name} | 删除中间件实例 |

#### response

200 ok

- 删除成功：返回200，附带被删除的实例名；
- 不存在：也返回200，附带提示信息；
- 有错误：返回相应错误码和错误信息。

## 状态检测

agent 调用 check 脚本时附带 metadata 信息；脚本执行完返回值格式定义如下：

```json
{
	"code": "0",  # 注意0是字符串，0表示正常运行，其他值都是异常
	"msg": "some message"
}
```

## 资源实时状态修改（仅通过agent调用）

### 数据库状态修改

#### request

| method | url                                    | desc         |
| ------ | -------------------------------------- | ------------ |
| PUT    | /apis/v1alpha1/database/{a_name}/check | health check |

### body

```json
{
	"code": "0",  # 注意0是字符串，0表示正常运行，其他值都是异常
	"msg": "some message"
}
```

#### response

202

### 中间件状态修改

#### request

| method | url                                      | desc         |
| ------ | ---------------------------------------- | ------------ |
| PUT    | /apis/v1alpha1/middleware/{a_name}/check | health check |

### body

```json
{
	"code": "0",  # 注意0是字符串，0表示正常运行，其他值都是异常
	"msg": "some message"
}
```

#### response

202

## 发生状态变化的应用列表查询

### 发生状态变化的数据库

#### request

| method | url                                         | desc               |
| ------ | ------------------------------------------- | ------------------ |
| GET    | /apis/v1alpha1/database/status/changed/{id} | 状态发生变化的应用 |

id 使用当前日期，比如 6月28日 的数据就是 0628

#### response

200 ok

```sh
["mysql-5.6-single-192.168.19.100","mysql-5.6-single-192.168.19.100-1"]
```

### 发生状态变化的中间件

#### request

| method | url                                           | desc               |
| ------ | --------------------------------------------- | ------------------ |
| GET    | /apis/v1alpha1/middleware/status/changed/{id} | 状态发生变化的应用 |

id 使用当前日期，比如 6月28日 的数据就是 0628

#### response

200 ok

```sh
["nginx-1.16-single-192.168.19.100","nginx-1.16-single-192.168.19.100-1"]
```

## agent api

### request

| method | url       | desc                                                         |
| ------ | --------- | ------------------------------------------------------------ |
| POST   | /{action} | action: install / start / stop / restart / uninstall / check |

### body

#### 数据库

```json
{
  "name": "mysql-5.6-single-192.168.19.100-xxx",
  "type": "database",
  "// operator_ip": "where operator in",
  "operator_ip": "192.168.19.13", 
  "operator_port": "3335",
  "repo_url": "http://192.168.19.200:123/ftp/software/mysql/5.7/",
  "install": "install.sh",
  "start": "start.sh",
  "stop": "stop.sh",
  "restart": "restart.sh",
  "uninstall": "uninstall.sh",
  "check": "check.sh",
  "package": "mysql-5.7.tar.gz",
  "metadata": {
    "// repo_url and package": "REPO_URL & PACKAGE are copy of repo_url & package, because they may needed by scripts",
    "REPO_URL": "http://192.168.19.200:123/ftp/software/mysql/5.7/",
    "PACKAGE": "mysql-5.7.tar.gz",
    "// env below": "they are set by who written scripts, because only that person know what env needed by scripts.",
    "APP_USER": "mysql",
    "APP_PASSWD": "MYSQL123"
  }
}
```

#### 中间件

```json
{
  "name": "nginx-1.16-single-192.168.19.100-xxx",
  "type": "middleware",
  "// operator_ip": "where operator in",
  "operator_ip": "192.168.19.13", 
  "operator_port": "3335",
  "repo_url": "http://192.168.19.200:123/ftp/software/nginx/1.16/",
  "install": "install.sh",
  "start": "start.sh",
  "stop": "stop.sh",
  "restart": "restart.sh",
  "uninstall": "uninstall.sh",
  "check": "check.sh",
  "package": "nginx-1.16.tar.gz",
  "metadata": {
    "// repo_url and package": "REPO_URL & PACKAGE are copy of repo_url & package, because they may needed by scripts",
    "REPO_URL": "http://192.168.19.200:123/ftp/software/nginx/1.16/",
    "PACKAGE": "mysql-5.7.tar.gz",
    "// env below": "they are set by who written scripts, because only that person know what env needed by scripts.",
    "APP_USER": "nginx",
    "APP_PASSWD": "NGINX123"
  }
}
```