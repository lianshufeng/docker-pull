@echo off

:: 进入项目
cd ../
:: 创建编译目录
mkdir build >nul 2>&1
go build -o ./build