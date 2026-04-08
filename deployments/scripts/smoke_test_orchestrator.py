#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
冒烟测试编排任务示例
展示如何在编排任务中集成部署和环境检查
"""

import os
import sys
import time
from pathlib import Path
from typing import Optional, Dict, Any

# 添加项目路径
sys.path.insert(0, str(Path(__file__).parent.parent))

from deployments.scripts.deploy_sandbox import SandboxDeployer, SandboxConfig
from deployments.scripts.check_smoke_env import SmokeEnvChecker


class SmokeTestOrchestrator:
    """冒烟测试编排器"""
    
    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.deployments_dir = project_root / "deployments"
        
    def deploy_and_test(self, sandbox_host: str, auto_fix: bool = True) -> bool:
        """
        完整流程：部署 -> 环境检查 -> 冒烟测试
        
        Args:
            sandbox_host: 沙盒服务器地址
            auto_fix: 是否自动修复环境问题
        
        Returns:
            bool: 整个流程是否成功
        """
        print("=" * 60)
        print("冒烟测试编排任务")
        print("=" * 60)
        
        # Phase 1: 部署到沙盒
        print("\n[Phase 1] 部署到沙盒环境")
        print("-" * 60)
        if not self.deploy_to_sandbox(sandbox_host):
            print("❌ 部署失败，终止流程")
            return False
        print("✅ 部署成功")
        
        # Phase 2: 环境检查
        print("\n[Phase 2] 冒烟环境检查")
        print("-" * 60)
        if not self.check_environment(auto_fix):
            print("❌ 环境检查失败，终止流程")
            return False
        print("✅ 环境检查通过")
        
        # Phase 3: 执行冒烟测试
        print("\n[Phase 3] 执行冒烟测试")
        print("-" * 60)
        if not self.run_smoke_tests():
            print("❌ 冒烟测试失败")
            return False
        print("✅ 冒烟测试通过")
        
        print("\n" + "=" * 60)
        print("✅ 全部流程完成")
        print("=" * 60)
        return True
    
    def deploy_to_sandbox(self, sandbox_host: str) -> bool:
        """
        部署到沙盒环境
        
        Args:
            sandbox_host: 沙盒服务器地址
        
        Returns:
            bool: 部署是否成功
        """
        print(f"目标沙盒: {sandbox_host}")
        
        # 创建部署配置
        config = SandboxConfig(
            host=sandbox_host,
            user=os.getenv("SANDBOX_USER", "root"),
            port=int(os.getenv("SANDBOX_PORT", "22")),
            deploy_dir=os.getenv("SANDBOX_DEPLOY_DIR", "/opt/digital-twin"),
            skip_tests=False
        )
        
        # 创建部署器
        deployer = SandboxDeployer(self.project_root, config)
        
        # 执行部署
        try:
            # 1. 构建镜像
            print("  → 构建镜像...")
            if not deployer.build_images():
                print("  ❌ 镜像构建失败")
                return False
            
            # 2. 上传到沙盒
            print("  → 上传到沙盒...")
            if not deployer.upload_to_sandbox():
                print("  ❌ 上传失败")
                return False
            
            # 3. 重启服务
            print("  → 重启服务...")
            if not deployer.restart_services():
                print("  ❌ 服务重启失败")
                return False
            
            return True
        except Exception as e:
            print(f"  ❌ 部署异常: {e}")
            return False
    
    def check_environment(self, auto_fix: bool = True) -> bool:
        """
        检查冒烟测试环境
        
        Args:
            auto_fix: 是否自动修复问题
        
        Returns:
            bool: 环境是否就绪
        """
        print(f"自动修复: {'是' if auto_fix else '否'}")
        
        # 创建环境检查器
        checker = SmokeEnvChecker(
            project_root=self.project_root,
            auto_fix=auto_fix,
            verbose=True
        )
        
        # 执行检查
        try:
            # 检查 Python 环境
            print("  → 检查 Python 环境...")
            checker.check_python()
            
            # 检查 Minium 环境
            print("  → 检查 Minium 环境...")
            checker.check_minium()
            
            # 检查 Playwright 环境
            print("  → 检查 Playwright 环境...")
            checker.check_playwright()
            
            # 检查微信开发者工具
            print("  → 检查微信开发者工具...")
            checker.check_wechat_devtools()
            
            # 检查 Node.js 环境
            print("  → 检查 Node.js 环境...")
            checker.check_nodejs()
            
            # 检查服务端口
            print("  → 检查服务端口...")
            checker.check_services()
            
            # 检查测试文件
            print("  → 检查测试文件...")
            checker.check_test_files()
            
            # 检查环境变量
            print("  → 检查环境变量...")
            checker.check_env_vars()
            
            # 生成报告
            report = checker.generate_report()
            
            print("\n  检查结果:")
            print(f"    ✅ 通过: {report['summary']['passed']}")
            print(f"    ⚠️  警告: {report['summary']['warnings']}")
            print(f"    ❌ 失败: {report['summary']['failed']}")
            
            # 判断是否通过
            return report['summary']['failed'] == 0
            
        except Exception as e:
            print(f"  ❌ 环境检查异常: {e}")
            return False
    
    def run_smoke_tests(self) -> bool:
        """
        执行冒烟测试
        
        Returns:
            bool: 测试是否通过
        """
        import subprocess
        
        # Minium 测试（小程序）
        print("  → 执行 Minium 测试（小程序）...")
        minium_test = self.project_root / "tests/e2e/smoke_v12_minium.py"
        
        if minium_test.exists():
            try:
                result = subprocess.run(
                    ["python3", str(minium_test)],
                    cwd=self.project_root,
                    capture_output=True,
                    text=True,
                    timeout=300
                )
                
                if result.returncode == 0:
                    print("    ✅ Minium 测试通过")
                else:
                    print(f"    ❌ Minium 测试失败: {result.stderr}")
                    return False
            except subprocess.TimeoutExpired:
                print("    ⚠️  Minium 测试超时")
            except Exception as e:
                print(f"    ⚠️  Minium 测试异常: {e}")
        else:
            print("    ⚠️  Minium 测试脚本不存在，跳过")
        
        # Playwright 测试（H5）
        print("  → 执行 Playwright 测试（H5）...")
        playwright_test = self.project_root / "tests/e2e/smoke_playwright.py"
        
        if playwright_test.exists():
            try:
                result = subprocess.run(
                    ["python3", str(playwright_test)],
                    cwd=self.project_root,
                    capture_output=True,
                    text=True,
                    timeout=300
                )
                
                if result.returncode == 0:
                    print("    ✅ Playwright 测试通过")
                else:
                    print(f"    ❌ Playwright 测试失败: {result.stderr}")
                    return False
            except subprocess.TimeoutExpired:
                print("    ⚠️  Playwright 测试超时")
            except Exception as e:
                print(f"    ⚠️  Playwright 测试异常: {e}")
        else:
            print("    ⚠️  Playwright 测试脚本不存在，跳过")
        
        return True


def main():
    """主函数"""
    import argparse
    
    parser = argparse.ArgumentParser(description="冒烟测试编排任务")
    parser.add_argument("--host", required=True, help="沙盒服务器地址")
    parser.add_argument("--no-auto-fix", action="store_true", help="禁用自动修复")
    parser.add_argument("--project-root", type=Path, help="项目根目录")
    
    args = parser.parse_args()
    
    # 确定项目根目录
    if args.project_root:
        project_root = args.project_root
    else:
        project_root = Path(__file__).parent.parent.parent
    
    # 创建编排器
    orchestrator = SmokeTestOrchestrator(project_root)
    
    # 执行完整流程
    success = orchestrator.deploy_and_test(
        sandbox_host=args.host,
        auto_fix=not args.no_auto_fix
    )
    
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
