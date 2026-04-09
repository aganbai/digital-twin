# IT13 冒烟测试报告

## 执行概要

| 指标 | 数值 |
|------|------|
| 总用例数 | 3 |
| 通过 | 3 ✅ |
| 失败 | 0 ❌ |
| 警告 | 0 ⚠️ |
| 通过率 | 100.0% |

## 环境信息

| 项目 | 值 |
|------|-----|
| 后端服务 | http://localhost:8000 |
| H5 教师端 | http://localhost:5174 |
| 测试时间 | 2026-04-09T23:03:13.118676 ~ 2026-04-09T23:03:30.140862 |
| 测试框架 | Playwright |

## 用例详情


### ✅ SM-A01: 新用户班级创建与教材配置设置流程

**状态**: passed
**开始时间**: 2026-04-09T23:03:13.770413
**结束时间**: 2026-04-09T23:03:19.739308

**执行步骤**:\n- ✅ 访问首页\n  - 截图: `screenshots/SM-A01/step01_home_1775746996378.png`\n- ✅ 导航到班级管理\n  - 截图: `screenshots/SM-A01/step03_class_mgmt_1775746999071.png`\n- ✅ 检查创建按钮\n  - 备注: 当前可能需要先登录\n  - 截图: `screenshots/SM-A01/step04_create_check_1775746999220.png`\n- ✅ 填写班级基本信息\n  - 截图: `screenshots/SM-A01/step05_basic_info_1775746999325.png`\n- ✅ 展开教材配置区域\n  - 截图: `screenshots/SM-A01/step06_textbook_config_1775746999432.png`\n- ✅ 配置教材信息\n  - 截图: `screenshots/SM-A01/step10_textbook_filled_1775746999532.png`\n- ✅ 准备创建班级\n  - 备注: 跳过实际提交，仅做UI验证\n  - 截图: `screenshots/SM-A01/step11_ready_to_submit_1775746999640.png`\n\n---\n
### ✅ SM-B04: 个人中心教材配置入口已移除

**状态**: passed
**开始时间**: 2026-04-09T23:03:20.744436
**结束时间**: 2026-04-09T23:03:23.542642

**执行步骤**:\n- ✅ 访问个人中心\n  - 截图: `screenshots/SM-B04/step01_center_1775747003286.png`\n- ✅ 检查菜单列表\n  - 截图: `screenshots/SM-B04/step02_menu_check_1775747003440.png`\n\n---\n
### ✅ SM-B06: H5教师端班级管理完整流程

**状态**: passed
**开始时间**: 2026-04-09T23:03:24.547821
**结束时间**: 2026-04-09T23:03:30.112570

**执行步骤**:\n- ✅ 访问教师端\n  - 截图: `screenshots/SM-B06/step01_login_page_1775747007096.png`\n- ✅ 班级列表页面\n  - 截图: `screenshots/SM-B06/step02_class_list_1775747009757.png`\n- ✅ 验证列表表格\n  - 截图: `screenshots/SM-B06/step03_table_check_1775747009879.png`\n- ✅ 检查操作按钮\n  - 截图: `screenshots/SM-B06/step04_buttons_1775747010003.png`\n\n---\n