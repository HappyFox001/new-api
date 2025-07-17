# Docker 部署指南 (端口 8000)

本指南说明如何使用 Docker 在端口 8000 上运行 New API 服务。

## 快速启动

### 1. 使用 Docker Compose（推荐）

```bash
# 启动所有服务（New API + MySQL + Redis）
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f new-api
```

服务将在以下端口运行：
- **New API**: http://localhost:8000
- **MySQL**: 内部端口 3306
- **Redis**: 内部端口 6379

### 2. 单独运行 New API Docker 容器

如果你已有数据库和 Redis：

```bash
docker run -d \
  --name new-api \
  -p 8000:8000 \
  -v ./data:/data \
  -v ./logs:/app/logs \
  -e PORT=8000 \
  -e SQL_DSN="your_database_connection_string" \
  -e REDIS_CONN_STRING="your_redis_connection_string" \
  calciumion/new-api:latest \
  --port 8000 --log-dir /app/logs
```

## 配置说明

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `PORT` | 应用端口 | 8000 |
| `SQL_DSN` | 数据库连接字符串 | - |
| `REDIS_CONN_STRING` | Redis 连接字符串 | - |
| `TZ` | 时区 | Asia/Shanghai |
| `ERROR_LOG_ENABLED` | 是否启用错误日志 | true |
| `SESSION_SECRET` | 会话密钥（多机部署必需） | - |

### 端口配置

- **容器内端口**: 8000
- **主机映射端口**: 8000
- **访问地址**: http://localhost:8000

## 数据持久化

### 数据目录

```bash
# 应用数据
./data:/data

# 日志文件
./logs:/app/logs

# MySQL 数据（如果使用 docker-compose）
mysql_data:/var/lib/mysql
```

## 常用命令

### 启动和停止

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 重启服务
docker-compose restart new-api

# 停止并删除所有数据
docker-compose down -v
```

### 日志查看

```bash
# 查看实时日志
docker-compose logs -f new-api

# 查看最近100行日志
docker-compose logs --tail=100 new-api

# 查看所有服务日志
docker-compose logs -f
```

### 容器管理

```bash
# 进入容器
docker exec -it new-api sh

# 查看容器状态
docker ps

# 查看容器资源使用
docker stats new-api
```

## 健康检查

服务包含自动健康检查：

```bash
# 手动检查服务状态
curl http://localhost:8000/api/status

# 检查健康状态
docker inspect --format='{{.State.Health.Status}}' new-api
```

## 更新服务

```bash
# 拉取最新镜像
docker-compose pull

# 重新启动服务
docker-compose up -d --force-recreate
```

## API 端点测试

启动后，你可以测试以下端点：

```bash
# 检查服务状态
curl http://localhost:8000/api/status

# 创建 API Token（需要先设置管理员账户）
curl -X POST http://localhost:8000/api/auto-token/create \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your_username",
    "password": "your_password",
    "token_name": "test_token",
    "remain_quota": 100000
  }'
```

## 故障排除

### 常见问题

1. **端口被占用**
   ```bash
   # 检查端口使用情况
   lsof -i :8000
   
   # 修改 docker-compose.yml 中的端口映射
   ports:
     - "8001:8000"  # 使用其他端口
   ```

2. **数据库连接失败**
   ```bash
   # 检查 MySQL 容器状态
   docker-compose logs mysql
   
   # 重启数据库服务
   docker-compose restart mysql
   ```

3. **权限问题**
   ```bash
   # 修复数据目录权限
   sudo chown -R $USER:$USER ./data ./logs
   ```

### 调试模式

```bash
# 以调试模式运行
docker-compose -f docker-compose.yml up --build
```

## 生产环境建议

1. **安全配置**
   - 设置强密码
   - 配置防火墙
   - 使用 HTTPS
   - 定期备份数据

2. **性能优化**
   - 配置适当的资源限制
   - 使用外部数据库和 Redis
   - 启用日志轮转

3. **监控**
   - 配置健康检查
   - 设置日志监控
   - 监控资源使用情况

## 备份和恢复

### 数据备份

```bash
# 备份应用数据
tar -czf backup-$(date +%Y%m%d).tar.gz ./data

# 备份数据库
docker exec mysql mysqldump -u root -p123456 new-api > backup-db-$(date +%Y%m%d).sql
```

### 数据恢复

```bash
# 恢复应用数据
tar -xzf backup-20240101.tar.gz

# 恢复数据库
docker exec -i mysql mysql -u root -p123456 new-api < backup-db-20240101.sql
``` 