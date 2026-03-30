#!/bin/bash
# ============================================================
# 阅芽 ReadBud — Development Environment Setup
# ============================================================

set -e

echo "=========================================="
echo "  阅芽 ReadBud — 开发环境初始化"
echo "  让写作从一个词开始生长。"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

ok()   { echo -e "  ${GREEN}✓${NC} $1"; }
fail() { echo -e "  ${RED}✗${NC} $1"; }
warn() { echo -e "  ${YELLOW}!${NC} $1"; }

# --- 1. Check prerequisites ---
echo "1. 检查环境依赖..."

if command -v go &> /dev/null; then
  ok "Go $(go version | awk '{print $3}')"
else
  fail "Go 未安装"; exit 1
fi

if command -v node &> /dev/null; then
  ok "Node.js $(node -v)"
else
  fail "Node.js 未安装"; exit 1
fi

if command -v npm &> /dev/null; then
  ok "npm $(npm -v)"
else
  fail "npm 未安装"; exit 1
fi

echo ""

# --- 2. Check PostgreSQL ---
echo "2. 检查 PostgreSQL..."

if command -v psql &> /dev/null; then
  ok "psql 已安装"
  if psql -U postgres -lqt 2>/dev/null | cut -d \| -f 1 | grep -qw readbud; then
    ok "数据库 'readbud' 已存在"
  else
    warn "数据库 'readbud' 不存在，尝试创建..."
    if createdb -U postgres readbud 2>/dev/null; then
      ok "数据库 'readbud' 创建成功"
    else
      warn "无法自动创建数据库，请手动执行: CREATE DATABASE readbud;"
    fi
  fi
else
  warn "psql 未安装，请确保 PostgreSQL 已配置"
fi

echo ""

# --- 3. Check Redis ---
echo "3. 检查 Redis..."

if command -v redis-cli &> /dev/null; then
  if redis-cli ping 2>/dev/null | grep -q PONG; then
    ok "Redis 已运行"
  else
    warn "Redis 未运行，请启动 Redis 服务"
  fi
else
  warn "redis-cli 未安装，请确保 Redis 已配置"
fi

echo ""

# --- 4. Go backend setup ---
echo "4. 初始化 Go 后端..."

cd backend
go mod tidy && go mod download
ok "Go 依赖安装完成"
cd ..

echo ""

# --- 5. Vue frontend setup ---
echo "5. 初始化 Vue 3 前端..."

cd frontend
npm install
ok "前端依赖安装完成"
cd ..

echo ""

# --- Summary ---
echo "=========================================="
echo "  初始化完成！"
echo "=========================================="
echo ""
echo "启动方式："
echo ""
echo "  后端 API (端口 19881):"
echo "    cd backend && go run cmd/api/main.go"
echo ""
echo "  后端 Worker:"
echo "    cd backend && go run cmd/worker/main.go"
echo ""
echo "  前端 (端口 19880):"
echo "    cd frontend && npm run dev"
echo ""
