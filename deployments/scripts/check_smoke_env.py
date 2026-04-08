#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
冒烟测试环境检查工具
检查 minium 和 playwright 环境，确保测试可以正常运行
"""

import os
import sys
import subprocess
import platform
import json
import time
from pathlib import Path
from typing import Dict, List, Tuple, Optional
from dataclasses import dataclass, field
from datetime import datetime
import logging

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout),
        logging.FileHandler(f'/tmp/smoke-env-check-{time.strftime("%Y%m%d_%H%M%S")}.log')
    ]
)
logger = logging.getLogger(__name__)


@dataclass
class CheckResult:
    """检查结果"""
    name: str
    category: str
    passed: bool
    message: str
    details: Optional[str] = None
    fix_suggestion: Optional[str] = None


@dataclass
class SmokeEnvChecker:
    """冒烟环境检查器"""
    project_root: Path
    auto_fix: bool = False
    verbose: bool = False
    results: List[CheckResult] = field(default_factory=list)
    
    def __post_init__(self):
        self.passed_count = 0
        self.failed_count = 0
        self.warn_count = 0
    
    def run_command(self, cmd: List[str], check: bool = False, capture: bool = True) -> subprocess.CompletedProcess:
        """执行命令"""
        try:
            if capture:
                result = subprocess.run(
                    cmd,
                    capture_output=True,
                    text=True,
                    timeout=30
                )
            else:
                result = subprocess.run(cmd, timeout=30)
            
            if check and result.returncode != 0:
                raise subprocess.CalledProcessError(
                    result.returncode, cmd, result.stdout, result.stderr
                )
            
            return result
        except subprocess.TimeoutExpired:
            logger.error(f"命令超时: {' '.join(cmd)}")
            raise
        except FileNotFoundError:
            logger.error(f"命令不存在: {cmd[0]}")
            raise
    
    def add_result(self, result: CheckResult):
        """添加检查结果"""
        self.results.append(result)
        
        if result.passed:
            self.passed_count += 1
            logger.info(f"✅ {result.message}")
        else:
            self.failed_count += 1
            logger.error(f"❌ {result.message}")
            if result.fix_suggestion:
                logger.info(f"   修复建议: {result.fix_suggestion}")
    
    def add_warning(self, result: CheckResult):
        """添加警告结果"""
        self.results.append(result)
        self.warn_count += 1
        logger.warning(f"⚠️  {result.message}")
        if result.fix_suggestion:
            logger.info(f"   建议: {result.fix_suggestion}")
    
    # ==================== Python 环境检查 ====================
    def check_python(self) -> List[CheckResult]:
        """检查 Python 环境"""
        logger.info("\n" + "=" * 50)
        logger.info("Python 环境检查")
        logger.info("=" * 50)
        
        results = []
        
        # 检查 Python 3
        try:
            result = self.run_command(["python3", "--version"])
            version = result.stdout.strip()
            
            self.add_result(CheckResult(
                name="python3",
                category="Python",
                passed=True,
                message=f"Python 已安装: {version}"
            ))
            
            # 检查版本 >= 3.8
            major, minor = map(int, version.split()[1].split('.')[:2])
            if major >= 3 and minor >= 8:
                self.add_result(CheckResult(
                    name="python_version",
                    category="Python",
                    passed=True,
                    message=f"Python 版本符合要求 (>= 3.8)"
                ))
            else:
                self.add_result(CheckResult(
                    name="python_version",
                    category="Python",
                    passed=False,
                    message=f"Python 版本过低，需要 >= 3.8，当前: {version}",
                    fix_suggestion="升级 Python 到 3.8 或更高版本"
                ))
        except Exception as e:
            self.add_result(CheckResult(
                name="python3",
                category="Python",
                passed=False,
                message=f"Python 3 未安装: {e}",
                fix_suggestion="安装 Python 3.8 或更高版本"
            ))
        
        # 检查 pip
        try:
            result = self.run_command(["python3", "-m", "pip", "--version"])
            pip_version = result.stdout.split()[1]
            
            self.add_result(CheckResult(
                name="pip",
                category="Python",
                passed=True,
                message=f"pip 已安装: {pip_version}"
            ))
        except Exception as e:
            self.add_result(CheckResult(
                name="pip",
                category="Python",
                passed=False,
                message=f"pip 未安装: {e}",
                fix_suggestion="python3 -m ensurepip --upgrade"
            ))
        
        return results
    
    # ==================== Minium 环境检查 ====================
    def check_minium(self) -> List[CheckResult]:
        """检查 Minium 环境"""
        logger.info("\n" + "=" * 50)
        logger.info("Minium 环境检查")
        logger.info("=" * 50)
        
        # 检查 minium 包
        try:
            import minium
            version = getattr(minium, "__version__", "unknown")
            
            self.add_result(CheckResult(
                name="minium_package",
                category="Minium",
                passed=True,
                message=f"minium 包已安装: {version}"
            ))
        except ImportError:
            self.add_result(CheckResult(
                name="minium_package",
                category="Minium",
                passed=False,
                message="minium 包未安装",
                fix_suggestion="pip3 install minium"
            ))
            
            if self.auto_fix:
                logger.info("正在安装 minium...")
                self.run_command(
                    ["python3", "-m", "pip", "install", "minium", "-i",
                     "https://pypi.tuna.tsinghua.edu.cn/simple"],
                    check=True
                )
            
            return []
        
        # 检查 minium 依赖
        minium_deps = ["requests", "pytest", "allure_pytest"]
        for dep in minium_deps:
            try:
                __import__(dep.replace("-", "_"))
                self.add_result(CheckResult(
                    name=f"minium_dep_{dep}",
                    category="Minium",
                    passed=True,
                    message=f"{dep} 已安装"
                ))
            except ImportError:
                self.add_warning(CheckResult(
                    name=f"minium_dep_{dep}",
                    category="Minium",
                    passed=False,
                    message=f"{dep} 未安装",
                    fix_suggestion=f"pip3 install {dep}"
                ))
        
        # 测试 Minium 导入
        try:
            from minium import WXMinium
            self.add_result(CheckResult(
                name="minium_import",
                category="Minium",
                passed=True,
                message="Minium 导入成功"
            ))
        except Exception as e:
            self.add_result(CheckResult(
                name="minium_import",
                category="Minium",
                passed=False,
                message=f"Minium 导入失败: {e}",
                fix_suggestion="检查 minium 安装是否完整"
            ))
        
        return []
    
    # ==================== Playwright 环境检查 ====================
    def check_playwright(self) -> List[CheckResult]:
        """检查 Playwright 环境"""
        logger.info("\n" + "=" * 50)
        logger.info("Playwright 环境检查")
        logger.info("=" * 50)
        
        # 检查 playwright 包
        try:
            import playwright
            version = getattr(playwright, "__version__", "unknown")
            
            self.add_result(CheckResult(
                name="playwright_package",
                category="Playwright",
                passed=True,
                message=f"playwright 包已安装: {version}"
            ))
        except ImportError:
            self.add_result(CheckResult(
                name="playwright_package",
                category="Playwright",
                passed=False,
                message="playwright 包未安装",
                fix_suggestion="pip3 install playwright"
            ))
            
            if self.auto_fix:
                logger.info("正在安装 playwright...")
                self.run_command(
                    ["python3", "-m", "pip", "install", "playwright", "-i",
                     "https://pypi.tuna.tsinghua.edu.cn/simple"],
                    check=True
                )
            
            return []
        
        # 检查浏览器
        browsers = ["chromium", "firefox", "webkit"]
        for browser in browsers:
            try:
                from playwright.sync_api import sync_playwright
                with sync_playwright() as p:
                    browser_obj = getattr(p, browser).launch(headless=True)
                    browser_obj.close()
                
                self.add_result(CheckResult(
                    name=f"playwright_{browser}",
                    category="Playwright",
                    passed=True,
                    message=f"{browser} 浏览器已安装"
                ))
            except Exception as e:
                self.add_warning(CheckResult(
                    name=f"playwright_{browser}",
                    category="Playwright",
                    passed=False,
                    message=f"{browser} 浏览器未安装或不可用: {e}",
                    fix_suggestion=f"playwright install {browser}"
                ))
        
        # 检查系统 Chrome
        chrome_paths = [
            "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
            "/usr/bin/google-chrome",
            "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"
        ]
        
        chrome_found = False
        for chrome_path in chrome_paths:
            if Path(chrome_path).exists():
                try:
                    result = self.run_command([chrome_path, "--version"])
                    version = result.stdout.strip()
                    
                    self.add_result(CheckResult(
                        name="system_chrome",
                        category="Playwright",
                        passed=True,
                        message=f"系统 Chrome 已安装: {version}"
                    ))
                    chrome_found = True
                    break
                except Exception:
                    pass
        
        if not chrome_found:
            self.add_warning(CheckResult(
                name="system_chrome",
                category="Playwright",
                passed=False,
                message="系统 Chrome 未安装（可选）",
                fix_suggestion="安装 Google Chrome 浏览器（可选）"
            ))
        
        return []
    
    # ==================== 微信开发者工具检查 ====================
    def check_wechat_devtools(self) -> List[CheckResult]:
        """检查微信开发者工具"""
        logger.info("\n" + "=" * 50)
        logger.info("微信开发者工具检查")
        logger.info("=" * 50)
        
        # macOS 默认路径
        devtools_path = Path("/Applications/wechatwebdevtools.app")
        cli_path = devtools_path / "Contents/MacOS/cli"
        
        if devtools_path.exists():
            self.add_result(CheckResult(
                name="wechat_devtools",
                category="DevTools",
                passed=True,
                message="微信开发者工具已安装"
            ))
            
            # 检查 CLI 工具
            if cli_path.exists():
                self.add_result(CheckResult(
                    name="wechat_cli",
                    category="DevTools",
                    passed=True,
                    message=f"CLI 工具可用: {cli_path}"
                ))
                
                # 检查服务端口
                try:
                    result = self.run_command([str(cli_path), "islogin"])
                    self.add_result(CheckResult(
                        name="devtools_port",
                        category="DevTools",
                        passed=True,
                        message="开发者工具服务端口已开启"
                    ))
                except Exception:
                    self.add_warning(CheckResult(
                        name="devtools_port",
                        category="DevTools",
                        passed=False,
                        message="开发者工具服务端口可能未开启",
                        fix_suggestion="在开发者工具中: 设置 → 安全 → 开启服务端口"
                    ))
            else:
                self.add_warning(CheckResult(
                    name="wechat_cli",
                    category="DevTools",
                    passed=False,
                    message="CLI 工具不存在"
                ))
            
            # 检查是否正在运行
            try:
                result = self.run_command(["pgrep", "-f", "wechatwebdevtools"])
                if result.returncode == 0:
                    self.add_result(CheckResult(
                        name="devtools_running",
                        category="DevTools",
                        passed=True,
                        message="开发者工具正在运行"
                    ))
                else:
                    self.add_warning(CheckResult(
                        name="devtools_running",
                        category="DevTools",
                        passed=False,
                        message="开发者工具未运行",
                        fix_suggestion="启动微信开发者工具"
                    ))
            except Exception:
                pass
        else:
            self.add_result(CheckResult(
                name="wechat_devtools",
                category="DevTools",
                passed=False,
                message="微信开发者工具未安装",
                fix_suggestion="从 https://developers.weixin.qq.com/miniprogram/dev/devtools/download.html 下载"
            ))
        
        # 检查编译产物
        dist_path = self.project_root / "src/frontend/dist"
        if dist_path.exists() and (dist_path / "app.json").exists():
            self.add_result(CheckResult(
                name="miniprogram_dist",
                category="DevTools",
                passed=True,
                message=f"小程序编译产物存在: {dist_path}"
            ))
        else:
            self.add_warning(CheckResult(
                name="miniprogram_dist",
                category="DevTools",
                passed=False,
                message="小程序编译产物不存在",
                fix_suggestion="cd src/frontend && npm run build:weapp"
            ))
        
        return []
    
    # ==================== 服务检查 ====================
    def check_services(self) -> List[CheckResult]:
        """检查服务端口"""
        logger.info("\n" + "=" * 50)
        logger.info("服务端口检查")
        logger.info("=" * 50)
        
        services = [
            ("后端服务", 8080, "/api/system/health"),
            ("Knowledge 服务", 8100, "/api/v1/health"),
            ("H5 管理端", 5173, "/"),
            ("H5 教师端", 5174, "/"),
            ("H5 学生端", 5175, "/")
        ]
        
        import urllib.request
        import urllib.error
        
        for name, port, endpoint in services:
            url = f"http://localhost:{port}{endpoint}"
            try:
                req = urllib.request.Request(url, method='GET')
                urllib.request.urlopen(req, timeout=2)
                
                self.add_result(CheckResult(
                    name=f"service_{port}",
                    category="Services",
                    passed=True,
                    message=f"{name} 正常 ({url})"
                ))
            except urllib.error.URLError:
                self.add_warning(CheckResult(
                    name=f"service_{port}",
                    category="Services",
                    passed=False,
                    message=f"{name} 未启动或不可访问",
                    fix_suggestion=f"启动 {name}"
                ))
            except Exception as e:
                self.add_warning(CheckResult(
                    name=f"service_{port}",
                    category="Services",
                    passed=False,
                    message=f"{name} 检查失败: {e}"
                ))
        
        return []
    
    # ==================== Node.js 环境检查 ====================
    def check_nodejs(self) -> List[CheckResult]:
        """检查 Node.js 环境"""
        logger.info("\n" + "=" * 50)
        logger.info("Node.js 环境检查")
        logger.info("=" * 50)
        
        # 检查 Node.js
        try:
            result = self.run_command(["node", "--version"])
            version = result.stdout.strip()
            
            self.add_result(CheckResult(
                name="nodejs",
                category="Node.js",
                passed=True,
                message=f"Node.js 已安装: {version}"
            ))
        except Exception as e:
            self.add_result(CheckResult(
                name="nodejs",
                category="Node.js",
                passed=False,
                message=f"Node.js 未安装: {e}",
                fix_suggestion="安装 Node.js >= 16"
            ))
            return []
        
        # 检查 npm
        try:
            result = self.run_command(["npm", "--version"])
            version = result.stdout.strip()
            
            self.add_result(CheckResult(
                name="npm",
                category="Node.js",
                passed=True,
                message=f"npm 已安装: {version}"
            ))
        except Exception as e:
            self.add_result(CheckResult(
                name="npm",
                category="Node.js",
                passed=False,
                message=f"npm 未安装: {e}",
                fix_suggestion="安装 npm"
            ))
        
        # 检查前端依赖
        frontend_dir = self.project_root / "src/frontend"
        node_modules = frontend_dir / "node_modules"
        
        if node_modules.exists():
            self.add_result(CheckResult(
                name="frontend_deps",
                category="Node.js",
                passed=True,
                message="前端依赖已安装"
            ))
            
            # 检查 miniprogram-automator
            automator = node_modules / "miniprogram-automator"
            if automator.exists():
                self.add_result(CheckResult(
                    name="miniprogram_automator",
                    category="Node.js",
                    passed=True,
                    message="miniprogram-automator 已安装"
                ))
            else:
                self.add_warning(CheckResult(
                    name="miniprogram_automator",
                    category="Node.js",
                    passed=False,
                    message="miniprogram-automator 未安装",
                    fix_suggestion="cd src/frontend && npm install miniprogram-automator"
                ))
        else:
            self.add_warning(CheckResult(
                name="frontend_deps",
                category="Node.js",
                passed=False,
                message="前端依赖未安装",
                fix_suggestion="cd src/frontend && npm install"
            ))
        
        return []
    
    # ==================== 测试文件检查 ====================
    def check_test_files(self) -> List[CheckResult]:
        """检查测试文件"""
        logger.info("\n" + "=" * 50)
        logger.info("测试文件检查")
        logger.info("=" * 50)
        
        # Minium 测试脚本
        minium_tests = [
            self.project_root / "tests/e2e/smoke_v12_minium.py",
            self.project_root / "src/frontend/e2e/smoke_v12_minium.py"
        ]
        
        for test_file in minium_tests:
            if test_file.exists():
                self.add_result(CheckResult(
                    name=f"test_{test_file.name}",
                    category="TestFiles",
                    passed=True,
                    message=f"测试脚本存在: {test_file.name}"
                ))
            else:
                self.add_warning(CheckResult(
                    name=f"test_{test_file.name}",
                    category="TestFiles",
                    passed=False,
                    message=f"测试脚本不存在: {test_file}",
                    fix_suggestion=f"创建测试脚本: {test_file}"
                ))
        
        # Playwright 测试脚本
        playwright_tests = [
            self.project_root / "tests/e2e/smoke_playwright.py",
            self.project_root / "src/frontend/e2e/smoke_h5_e2e.py"
        ]
        
        for test_file in playwright_tests:
            if test_file.exists():
                self.add_result(CheckResult(
                    name=f"test_{test_file.name}",
                    category="TestFiles",
                    passed=True,
                    message=f"测试脚本存在: {test_file.name}"
                ))
            else:
                self.add_warning(CheckResult(
                    name=f"test_{test_file.name}",
                    category="TestFiles",
                    passed=False,
                    message=f"测试脚本不存在: {test_file}",
                    fix_suggestion=f"创建测试脚本: {test_file}"
                ))
        
        # 测试用例文档
        test_cases_doc = self.project_root / "docs/smoke-test-cases.md"
        if test_cases_doc.exists():
            self.add_result(CheckResult(
                name="test_cases_doc",
                category="TestFiles",
                passed=True,
                message="测试用例文档存在"
            ))
        else:
            self.add_warning(CheckResult(
                name="test_cases_doc",
                category="TestFiles",
                passed=False,
                message="测试用例文档不存在",
                fix_suggestion="创建测试用例文档: docs/smoke-test-cases.md"
            ))
        
        return []
    
    # ==================== 环境变量检查 ====================
    def check_env_vars(self) -> List[CheckResult]:
        """检查环境变量"""
        logger.info("\n" + "=" * 50)
        logger.info("环境变量检查")
        logger.info("=" * 50)
        
        env_file = self.project_root / ".env"
        
        if env_file.exists():
            self.add_result(CheckResult(
                name="env_file",
                category="EnvVars",
                passed=True,
                message=".env 文件存在"
            ))
            
            # 读取并检查环境变量
            env_vars = {}
            with open(env_file) as f:
                for line in f:
                    line = line.strip()
                    if line and not line.startswith('#') and '=' in line:
                        # 去除 export 前缀
                        if line.startswith('export '):
                            line = line[7:]
                        key, value = line.split('=', 1)
                        # 去除引号
                        value = value.strip().strip('"').strip("'")
                        env_vars[key.strip()] = value
            
            # 检查关键变量
            if env_vars.get("JWT_SECRET") and env_vars["JWT_SECRET"] not in ["change-me-to-a-random-string", "your-secret-key"]:
                self.add_result(CheckResult(
                    name="jwt_secret",
                    category="EnvVars",
                    passed=True,
                    message="JWT_SECRET 已配置"
                ))
            else:
                self.add_warning(CheckResult(
                    name="jwt_secret",
                    category="EnvVars",
                    passed=False,
                    message="JWT_SECRET 未正确配置",
                    fix_suggestion="在 .env 文件中设置 JWT_SECRET"
                ))
            
            if env_vars.get("OPENAI_API_KEY") and env_vars["OPENAI_API_KEY"] != "your-api-key":
                self.add_result(CheckResult(
                    name="openai_api_key",
                    category="EnvVars",
                    passed=True,
                    message="OPENAI_API_KEY 已配置"
                ))
            else:
                self.add_warning(CheckResult(
                    name="openai_api_key",
                    category="EnvVars",
                    passed=False,
                    message="OPENAI_API_KEY 未正确配置",
                    fix_suggestion="在 .env 文件中设置 OPENAI_API_KEY"
                ))
        else:
            self.add_warning(CheckResult(
                name="env_file",
                category="EnvVars",
                passed=False,
                message=".env 文件不存在",
                fix_suggestion="创建 .env 文件并配置必要的环境变量"
            ))
        
        return []
    
    # ==================== 生成报告 ====================
    def generate_report(self) -> Dict:
        """生成检查报告"""
        report = {
            "timestamp": datetime.now().isoformat(),
            "project_root": str(self.project_root),
            "summary": {
                "total": len(self.results),
                "passed": self.passed_count,
                "failed": self.failed_count,
                "warnings": self.warn_count
            },
            "results": [
                {
                    "name": r.name,
                    "category": r.category,
                    "passed": r.passed,
                    "message": r.message,
                    "details": r.details,
                    "fix_suggestion": r.fix_suggestion
                }
                for r in self.results
            ]
        }
        
        return report
    
    def print_summary(self):
        """打印检查摘要"""
        logger.info("\n" + "=" * 50)
        logger.info("环境检查报告")
        logger.info("=" * 50)
        logger.info(f"\n检查时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        logger.info(f"\n检查结果统计:")
        logger.info(f"  ✅ 通过: {self.passed_count}")
        logger.info(f"  ⚠️  警告: {self.warn_count}")
        logger.info(f"  ❌ 失败: {self.failed_count}")
        
        logger.info(f"\n详细结果:")
        for result in self.results:
            status = "✅" if result.passed else "❌"
            logger.info(f"  {status} [{result.category}] {result.message}")
        
        logger.info("")
        
        if self.failed_count == 0:
            if self.warn_count == 0:
                logger.info("=" * 50)
                logger.info("✅ 环境检查完全通过！")
                logger.info("可以开始执行冒烟测试")
                logger.info("=" * 50)
            else:
                logger.info("=" * 50)
                logger.info(f"⚠️  环境检查通过，但有 {self.warn_count} 个警告")
                logger.info("建议处理警告后再执行测试")
                logger.info("=" * 50)
            
            logger.info("\n执行冒烟测试命令:")
            logger.info("  # Minium 测试（小程序）")
            logger.info("  python3 tests/e2e/smoke_v12_minium.py")
            logger.info("")
            logger.info("  # Playwright 测试（H5）")
            logger.info("  python3 tests/e2e/smoke_playwright.py")
            
            return True
        else:
            logger.info("=" * 50)
            logger.info(f"❌ 环境检查失败，发现 {self.failed_count} 个错误")
            logger.info("请修复错误后再执行测试")
            logger.info("=" * 50)
            
            logger.info("\n修复建议:")
            logger.info(f"  1. 使用 --fix 参数自动修复")
            logger.info("  2. 手动安装缺失的依赖")
            logger.info("  3. 启动必要的服务")
            
            return False
    
    # ==================== 执行所有检查 ====================
    def run_all_checks(self) -> bool:
        """执行所有检查"""
        logger.info("=" * 50)
        logger.info("冒烟测试环境检查")
        logger.info("=" * 50)
        
        self.check_python()
        self.check_minium()
        self.check_playwright()
        self.check_wechat_devtools()
        self.check_nodejs()
        self.check_services()
        self.check_test_files()
        self.check_env_vars()
        
        return self.print_summary()


def main():
    """主函数"""
    import argparse
    
    parser = argparse.ArgumentParser(description="冒烟测试环境检查工具")
    parser.add_argument("--fix", action="store_true", help="自动修复可修复的问题")
    parser.add_argument("--verbose", "-v", action="store_true", help="显示详细输出")
    parser.add_argument("--project-root", type=Path, help="项目根目录")
    parser.add_argument("--json", action="store_true", help="输出 JSON 格式报告")
    
    args = parser.parse_args()
    
    # 确定项目根目录
    if args.project_root:
        project_root = args.project_root
    else:
        # 自动查找项目根目录
        current = Path.cwd()
        while current != current.parent:
            if (current / "docker-compose.yml").exists():
                project_root = current
                break
            current = current.parent
        else:
            logger.error("无法找到项目根目录")
            sys.exit(1)
    
    # 创建检查器
    checker = SmokeEnvChecker(
        project_root=project_root,
        auto_fix=args.fix,
        verbose=args.verbose
    )
    
    # 执行检查
    success = checker.run_all_checks()
    
    # 输出 JSON 报告
    if args.json:
        report = checker.generate_report()
        print(json.dumps(report, indent=2, ensure_ascii=False))
    
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
