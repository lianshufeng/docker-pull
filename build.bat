@echo off
:: 创建编译目录
mkdir build >nul 2>&1
:: 进入项目
cd /d core
go build -o ../build