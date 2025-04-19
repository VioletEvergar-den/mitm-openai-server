#!/bin/bash

# 脚本错误时退出
set -e

# 显示帮助信息
show_help() {
  echo "用法: $0 [选项]"
  echo "选项:"
  echo "  -h, --help                   显示帮助信息"
  echo "  -d, --dev                    开发模式构建"
  echo "  -p, --prod                   生产模式构建"
  echo "  -c, --clean                  清理构建文件"
  echo ""
}

# 清理函数
clean() {
  echo "清理构建文件..."
  rm -rf ./react-ui/dist
  rm -f ./mitm-openai-server
  echo "清理完成！"
}

# 构建前端应用
build_frontend() {
  echo "构建前端应用..."
  cd ./react-ui
  npm install
  npm run build
  # 确保dist目录存在
  if [ ! -d "dist" ]; then
    echo "错误: 前端构建失败，没有生成dist目录"
    exit 1
  fi
  cd ..
  echo "前端构建完成！"
}

# 构建后端应用
build_backend() {
  local mode=$1
  echo "构建后端应用 (模式: $mode)..."
  if [ "$mode" = "dev" ]; then
    GO_ENV=development go build -o mitm-openai-server
  else
    go build -o mitm-openai-server
  fi
  echo "后端构建完成！"
}

# 主函数
main() {
  local mode="prod"

  # 处理命令行参数
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -h|--help)
        show_help
        exit 0
        ;;
      -d|--dev)
        mode="dev"
        shift
        ;;
      -p|--prod)
        mode="prod"
        shift
        ;;
      -c|--clean)
        clean
        exit 0
        ;;
      *)
        echo "未知选项: $1"
        show_help
        exit 1
        ;;
    esac
  done

  # 执行构建流程
  if [ "$mode" = "dev" ]; then
    echo "执行开发模式构建..."
    build_backend "dev"
    echo ""
    echo "开发模式构建完成！"
    echo "使用以下命令启动后端服务器："
    echo "  ./mitm-openai-server server"
    echo ""
    echo "然后在另一个终端启动前端开发服务器："
    echo "  cd react-ui && npm run dev"
  else
    echo "执行生产模式构建..."
    build_frontend
    build_backend "prod"
    echo ""
    echo "生产模式构建完成！"
    echo "使用以下命令启动服务器："
    echo "  ./mitm-openai-server server"
  fi
}

# 执行主函数
main "$@" 