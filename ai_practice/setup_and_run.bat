@echo off
chcp 65001 >nul
echo ========================================
echo   论文关键词分析工具 - 安装与运行
echo ========================================
echo.

REM 检查Python是否安装
py --version >nul 2>&1
if errorlevel 1 (
    echo [错误] 未找到Python，请先安装Python 3.x
    echo 安装包: python-3.13.14-amd64.exe
    pause
    exit /b 1
)

echo [1/4] Python版本:
py --version
echo.

REM 创建目录
echo [2/4] 创建项目目录...
if not exist data mkdir data
if not exist output mkdir output
echo 目录创建完成
echo.

REM 复制数据文件
echo [3/4] 检查数据文件...
if not exist data\论文信息.xlsx (
    if exist ..\论文信息.xlsx (
        copy ..\论文信息.xlsx data\ >nul
        echo 已复制数据文件到 data\论文信息.xlsx
    ) else if exist 论文信息.xlsx (
        copy 论文信息.xlsx data\ >nul
        echo 已复制数据文件到 data\论文信息.xlsx
    ) else (
        echo [警告] 未找到论文信息.xlsx，请手动复制到 data\ 目录
    )
) else (
    echo 数据文件已存在
)
echo.

REM 安装依赖
echo [4/4] 安装Python依赖包...
echo 这可能需要几分钟时间...
py -m pip install -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple
if errorlevel 1 (
    echo [错误] 依赖安装失败，请检查网络连接
    pause
    exit /b 1
)
echo 依赖安装完成
echo.

echo ========================================
echo   开始运行分析程序...
echo ========================================
echo.
py analysis_report.py

echo.
echo ========================================
echo   分析完成！
echo   结果保存在 output\ 目录
echo ========================================
pause
