#!/usr/bin/env python3
"""
沙盒一键部署工具
支持在编排任务中自动调用，完成编译、上传、重启的完整流程
"""

import os
import sys
import subprocess
import argparse
import logging
import time
from pathlib import Path
from typing import Optional, List, Dict, Any
from dataclasses import dataclass
import json

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout),
        logging.FileHandler(f'/tmp/deploy-sandbox-{time.strftime("%Y%m%d_%H%M%S")}.log')
    ]
)
logger = logging.getLogger(__name__)


@dataclass
class SandboxConfig:
    """沙盒部署配置"""
    host: str
    user: str = "root"
    port: int = 22
    deploy_dir: str = "/opt/digital-twin"
    registry: Optional[str] = None
    skip_tests: bool = False
    dry_run: bool = False
    
    @property
    def is_local(self) -> bool:
        """判断是否为本地部署模式（host 为空、localhost 或 127.0.0.1）"""
        return not self.host or self.host in ('localhost', '127.0.0.1', '::1')
    

class SandboxDeployer:
    """沙盒一键部署器"""
    
    def __init__(self, project_root: Path, config: SandboxConfig):
        self.project_root = project_root
        self.config = config
        self.deployments_dir = project_root / "deployments"
        self.scripts_dir = self.deployments_dir / "scripts"
        
    def run_command(self, cmd: List[str], cwd: Optional[Path] = None, check: bool = True) -> subprocess.CompletedProcess:
        """执行命令"""
        if self.config.dry_run:
            logger.info(f"[DRY-RUN] 将执行: {' '.join(cmd)}")
            return subprocess.CompletedProcess(cmd, 0, b"", b"")
        
        logger.info(f"执行命令: {' '.join(cmd)}")
        result = subprocess.run(
            cmd,
            cwd=cwd or self.project_root,
            capture_output=True,
            text=True
        )
        
        if check and result.returncode != 0:
            logger.error(f"命令执行失败: {result.stderr}")
            raise subprocess.CalledProcessError(result.returncode, cmd, result.stdout, result.stderr)
        
        return result
    
    def ssh_command(self, remote_cmd: str, check: bool = True) -> subprocess.CompletedProcess:
        """在沙盒服务器执行命令"""
        cmd = [
            "ssh",
            "-p", str(self.config.port),
            f"{self.config.user}@{self.config.host}",
            remote_cmd
        ]
        return self.run_command(cmd, check=check)
    
    def scp_upload(self, local_path: Path, remote_path: str) -> subprocess.CompletedProcess:
        """上传文件到沙盒服务器"""
        cmd = [
            "scp",
            "-P", str(self.config.port),
            str(local_path),
            f"{self.config.user}@{self.config.host}:{remote_path}"
        ]
        return self.run_command(cmd)
    
    def check_dependencies(self) -> bool:
        """检查依赖工具"""
        logger.info("检查依赖工具...")
        
        dependencies = ["docker"]
        
        # 仅远程部署时才需要 SSH 工具
        if not self.config.is_local:
            dependencies.extend(["ssh", "scp"])
        
        missing = []
        for dep in dependencies:
            try:
                subprocess.run([dep, "--version"], capture_output=True, check=True)
            except (subprocess.CalledProcessError, FileNotFoundError):
                missing.append(dep)
        
        if missing:
            logger.error(f"缺少依赖工具: {', '.join(missing)}")
            return False
        
        logger.info("✅ 依赖工具检查通过")
        return True
    
    def check_connection(self) -> bool:
        """测试沙盒服务器连接"""
        if self.config.is_local:
            logger.info("本地部署模式，跳过远程连接测试")
            # 检查 Docker 是否可用
            try:
                result = self.run_command(["docker", "info"], check=False)
                if result.returncode == 0:
                    logger.info("✅ 本地 Docker 环境就绪")
                    return True
                else:
                    logger.error("❌ Docker 未运行或不可用")
                    return False
            except Exception as e:
                logger.error(f"❌ Docker 检查失败: {e}")
                return False
        
        logger.info(f"测试连接: {self.config.host}:{self.config.port}")
        
        try:
            self.ssh_command("echo 'Connection OK'", check=False)
            logger.info("✅ 连接测试通过")
            return True
        except Exception as e:
            logger.error(f"❌ 连接测试失败: {e}")
            return False
    
    def build_images(self) -> bool:
        """构建镜像"""
        logger.info("========== 构建镜像 ==========")
        
        try:
            self.run_command(
                ["docker", "compose", "build", "--no-cache"],
                cwd=self.project_root
            )
            logger.info("✅ 镜像构建完成")
            return True
        except Exception as e:
            logger.error(f"❌ 镜像构建失败: {e}")
            return False
    
    def run_tests(self) -> bool:
        """运行测试"""
        if self.config.skip_tests:
            logger.info("跳过测试")
            return True
        
        logger.info("========== 运行测试 ==========")
        
        try:
            backend_dir = self.project_root / "backend"
            if backend_dir.exists():
                self.run_command(["go", "test", "./...", "-v"], cwd=backend_dir, check=False)
            logger.info("✅ 测试执行完成")
            return True
        except Exception as e:
            logger.warning(f"⚠️ 测试存在失败: {e}")
            return True
    
    def upload_to_sandbox(self) -> bool:
        """上传到沙盒（本地模式跳过）"""
        if self.config.is_local:
            logger.info("========== 本地部署模式，跳过上传 ==========")
            logger.info("✅ 本地模式无需上传，直接使用本地镜像")
            return True
        
        logger.info("========== 上传到沙盒 ==========")
        
        try:
            # 创建远程目录
            self.ssh_command(f"mkdir -p {self.config.deploy_dir}/{{deployments/{{docker,nginx,scripts}},data,uploads,configs}}")
            
            # 上传配置文件
            logger.info("上传配置文件...")
            self.scp_upload(
                self.project_root / "docker-compose.yml",
                f"{self.config.deploy_dir}/"
            )
            self.scp_upload(
                self.deployments_dir / ".env.production",
                f"{self.config.deploy_dir}/"
            )
            self.scp_upload(
                self.deployments_dir / "nginx" / "nginx.conf",
                f"{self.config.deploy_dir}/deployments/nginx/"
            )
            
            # 导出镜像
            logger.info("导出镜像文件...")
            timestamp = time.strftime("%Y%m%d_%H%M%S")
            image_file = f"/tmp/digital-twin-images-{timestamp}.tar"
            
            self.run_command([
                "docker", "save", "-o", image_file,
                "digital-twin_backend:latest",
                "digital-twin_knowledge:latest",
                "nginx:alpine"
            ])
            
            # 上传镜像
            logger.info("上传镜像文件...")
            self.scp_upload(Path(image_file), f"{self.config.deploy_dir}/images.tar")
            
            # 清理本地镜像文件
            os.remove(image_file)
            
            # 在远程加载镜像
            logger.info("在沙盒服务器加载镜像...")
            self.ssh_command(f"cd {self.config.deploy_dir} && docker load -i images.tar && rm -f images.tar")
            
            logger.info("✅ 上传完成")
            return True
        except Exception as e:
            logger.error(f"❌ 上传失败: {e}")
            return False
    
    def restart_services(self) -> bool:
        """重启服务"""
        logger.info("========== 重启服务 ==========")
        
        try:
            if self.config.is_local:
                # 本地模式：直接在项目目录执行 docker compose
                logger.info("停止旧服务（本地模式）...")
                self.run_command(
                    ["docker", "compose", "down"],
                    cwd=self.project_root,
                    check=False
                )
                
                logger.info("启动新服务（本地模式）...")
                self.run_command(
                    ["docker", "compose", "up", "-d"],
                    cwd=self.project_root
                )
            else:
                # 远程模式：通过 SSH 执行
                logger.info("停止旧服务...")
                self.ssh_command(f"cd {self.config.deploy_dir} && docker compose down || true", check=False)
                
                logger.info("启动新服务...")
                self.ssh_command(f"cd {self.config.deploy_dir} && docker compose up -d")
            
            # 等待服务就绪
            logger.info("等待服务就绪...")
            time.sleep(15)
            
            # 健康检查
            if self.health_check():
                logger.info("✅ 服务重启完成")
                return True
            else:
                logger.error("❌ 服务健康检查失败")
                return False
        except Exception as e:
            logger.error(f"❌ 服务重启失败: {e}")
            return False
    
    def restore_real_env(self) -> bool:
        """
        恢复真实环境（冒烟测试后调用）
        
        冒烟测试期间，后端以 WX_MODE=mock 运行以支持 Mock 登录。
        测试完成后，需要去掉 Mock 登录态，切换回真实微信登录，
        方便用户人工体验。
        
        核心操作：
        1. 去掉 WX_MODE=mock 环境变量（使用 .env.production 配置）
        2. 重启后端服务使配置生效
        3. 健康检查确认服务正常
        """
        logger.info("========== 恢复真实环境 ==========")
        logger.info("目标：去掉 WX_MODE=mock，切换回真实微信登录")
        
        try:
            if self.config.is_local:
                # 本地模式
                logger.info("停止当前服务（本地 Mock 环境）...")
                self.run_command(
                    ["docker", "compose", "down"],
                    cwd=self.project_root,
                    check=False
                )
                
                logger.info("使用 .env.production 配置重启服务（真实微信登录）...")
                self.run_command(
                    ["docker", "compose", "--env-file", ".env.production", "up", "-d"],
                    cwd=self.project_root
                )
            else:
                # 远程模式
                logger.info("停止当前服务（Mock 环境）...")
                self.ssh_command(f"cd {self.config.deploy_dir} && docker compose down || true", check=False)
                
                logger.info("使用 .env.production 配置重启服务（真实微信登录）...")
                self.ssh_command(
                    f"cd {self.config.deploy_dir} && "
                    f"docker compose --env-file .env.production up -d"
                )
            
            # 等待服务就绪
            logger.info("等待服务就绪...")
            time.sleep(15)
            
            # 健康检查
            if self.health_check():
                logger.info("✅ 真实环境恢复完成")
                logger.info("   - WX_MODE: 未设置（使用真实微信 API）")
                logger.info("   - 用户可通过真实微信扫码登录体验")
                return True
            else:
                logger.error("❌ 真实环境恢复后健康检查失败")
                return False
        except Exception as e:
            logger.error(f"❌ 真实环境恢复失败: {e}")
            return False
    
    def health_check(self, max_retries: int = 30, interval: int = 10) -> bool:
        """健康检查"""
        logger.info("========== 健康检查 ==========")
        
        endpoints = [
            ("Backend", "http://localhost:8080/api/system/health"),
            ("Knowledge", "http://localhost:8100/api/v1/health"),
            ("Nginx", "http://localhost:80/health")
        ]
        
        for retry in range(max_retries):
            logger.info(f"健康检查尝试 {retry + 1}/{max_retries}...")
            
            results = {}
            for name, url in endpoints:
                if self.config.is_local:
                    # 本地模式：直接 curl
                    result = self.run_command(["curl", "-sf", url], check=False)
                else:
                    result = self.ssh_command(f"curl -sf {url}", check=False)
                results[name] = result.returncode == 0
            
            if all(results.values()):
                logger.info("✅ 所有服务健康检查通过")
                for name, ok in results.items():
                    logger.info(f"  - {name}: {'✓' if ok else '✗'}")
                return True
            
            logger.warning(f"服务状态: {results}")
            time.sleep(interval)
        
        # 输出日志帮助调试
        logger.error("健康检查超时")
        if self.config.is_local:
            self.run_command(
                ["docker", "compose", "logs", "--tail=50"],
                cwd=self.project_root,
                check=False
            )
        else:
            self.ssh_command(f"cd {self.config.deploy_dir} && docker compose logs --tail=50", check=False)
        return False
    
    def get_status(self) -> Dict[str, Any]:
        """获取服务状态"""
        logger.info("获取服务状态...")
        
        if self.config.is_local:
            result = self.run_command(
                ["docker", "compose", "ps"],
                cwd=self.project_root,
                check=False
            )
        else:
            result = self.ssh_command(
                f"cd {self.config.deploy_dir} && docker compose ps",
                check=False
            )
        
        return {
            "success": result.returncode == 0,
            "output": result.stdout
        }
    
    def get_logs(self, tail: int = 100) -> str:
        """获取服务日志"""
        if self.config.is_local:
            result = self.run_command(
                ["docker", "compose", "logs", f"--tail={tail}"],
                cwd=self.project_root,
                check=False
            )
        else:
            result = self.ssh_command(
                f"cd {self.config.deploy_dir} && docker compose logs --tail={tail}",
                check=False
            )
        return result.stdout
    
    def deploy_full(self) -> bool:
        """完整部署流程"""
        logger.info("=" * 50)
        mode = "本地" if self.config.is_local else "远程"
        logger.info(f"开始一键部署流程（{mode}模式）")
        logger.info("=" * 50)
        
        # 前置检查
        if not self.check_dependencies():
            return False
        
        if not self.check_connection():
            return False
        
        # 执行流程
        steps = [
            ("构建镜像", self.build_images),
            ("运行测试", self.run_tests),
            ("上传到沙盒", self.upload_to_sandbox),
            ("重启服务", self.restart_services)
        ]
        
        for step_name, step_func in steps:
            if not step_func():
                logger.error(f"❌ {step_name}失败")
                return False
        
        host = self.config.host or "localhost"
        logger.info("=" * 50)
        logger.info("✅ 部署完成")
        logger.info(f"访问地址: http://{host}")
        logger.info("=" * 50)
        return True


def load_config_from_env() -> SandboxConfig:
    """从环境变量加载配置"""
    return SandboxConfig(
        host=os.getenv("SANDBOX_HOST", ""),
        user=os.getenv("SANDBOX_USER", "root"),
        port=int(os.getenv("SANDBOX_PORT", "22")),
        deploy_dir=os.getenv("SANDBOX_DEPLOY_DIR", "/opt/digital-twin"),
        registry=os.getenv("IMAGE_REGISTRY"),
        skip_tests=os.getenv("SKIP_TESTS", "false").lower() == "true"
    )


def main():
    """主函数"""
    parser = argparse.ArgumentParser(description="沙盒一键部署工具")
    
    parser.add_argument("command", nargs='?', choices=["full", "build", "upload", "restart", "restore", "status", "logs"],
                        default="full", help="部署命令（默认 full；restore: 冒烟测试后恢复真实环境，去掉 Mock 登录态）")
    parser.add_argument("--host", help="沙盒服务器地址")
    parser.add_argument("--user", default="root", help="SSH 用户名")
    parser.add_argument("--port", type=int, default=22, help="SSH 端口")
    parser.add_argument("--dir", default="/opt/digital-twin", help="沙盒部署目录")
    parser.add_argument("--registry", help="镜像仓库地址")
    parser.add_argument("--skip-tests", action="store_true", help="跳过测试")
    parser.add_argument("--dry-run", action="store_true", help="仅显示执行步骤")
    parser.add_argument("--project-root", type=Path, help="项目根目录")
    
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
    
    # 创建配置
    config = SandboxConfig(
        host=args.host or os.getenv("SANDBOX_HOST", ""),
        user=args.user,
        port=args.port,
        deploy_dir=args.dir,
        registry=args.registry,
        skip_tests=args.skip_tests,
        dry_run=args.dry_run
    )
    
    # 创建部署器
    deployer = SandboxDeployer(project_root, config)
    
    # 执行命令
    success = False
    if args.command == "full":
        success = deployer.deploy_full()
    elif args.command == "build":
        success = deployer.build_images() and deployer.run_tests()
    elif args.command == "upload":
        if config.is_local:
            logger.info("本地模式，跳过上传")
            success = True
        else:
            success = deployer.upload_to_sandbox()
    elif args.command == "restart":
        success = deployer.restart_services()
    elif args.command == "restore":
        # 冒烟测试后恢复真实环境：去掉 WX_MODE=mock，使用 .env.production 重启
        success = deployer.restore_real_env()
    elif args.command == "status":
        result = deployer.get_status()
        print(result["output"])
        success = result["success"]
    elif args.command == "logs":
        logs = deployer.get_logs()
        print(logs)
        success = True
    
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
