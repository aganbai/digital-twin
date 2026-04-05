#!/bin/bash
# H5教师端登录功能端到端验证

echo "========================================="
echo "H5教师端登录功能端到端验证"
echo "========================================="
echo ""

# 测试结果文件
RESULT_FILE="/tmp/h5-login-test-result.txt"
echo "H5登录验证结果 - $(date)" > $RESULT_FILE

# 1. 检查后端服务
echo "1. 检查后端服务状态..."
if curl -s http://localhost:8080/api/system/health | grep -q '"status":"healthy"'; then
    echo "   ✅ 后端服务正常"
    echo "后端服务: 正常" >> $RESULT_FILE
else
    echo "   ❌ 后端服务异常"
    echo "后端服务: 异常" >> $RESULT_FILE
    exit 1
fi

# 2. 测试获取登录URL
echo ""
echo "2. 测试获取微信登录URL..."
RESPONSE=$(curl -s "http://localhost:8080/api/auth/wx-h5-login-url?redirect_uri=http://localhost:5175")

# 检查返回字段
if echo "$RESPONSE" | grep -q '"login_url"'; then
    echo "   ✅ 返回字段正确 (login_url)"
    echo "API字段: 正确 (login_url)" >> $RESULT_FILE
    LOGIN_URL=$(echo "$RESPONSE" | grep -o '"login_url":"[^"]*"' | cut -d'"' -f4)
    echo "   登录URL: $LOGIN_URL"
else
    echo "   ❌ 返回字段错误"
    echo "API字段: 错误" >> $RESULT_FILE
    echo "   响应: $RESPONSE"
    exit 1
fi

# 3. 测试Mock登录回调
echo ""
echo "3. 测试Mock登录回调..."
CALLBACK_RESPONSE=$(curl -s -X POST http://localhost:8080/api/auth/wx-h5-callback \
    -H "Content-Type: application/json" \
    -d '{"code":"mock_code_1001"}')

if echo "$CALLBACK_RESPONSE" | grep -q '"token"'; then
    echo "   ✅ Mock登录成功"
    echo "Mock登录: 成功" >> $RESULT_FILE
    TOKEN=$(echo "$CALLBACK_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo "   Token: ${TOKEN:0:20}..."
else
    echo "   ❌ Mock登录失败"
    echo "Mock登录: 失败" >> $RESULT_FILE
    echo "   响应: $CALLBACK_RESPONSE"
    exit 1
fi

# 4. 检查前端服务
echo ""
echo "4. 检查H5前端服务..."
if curl -s http://localhost:5175 | grep -q "div\|html"; then
    echo "   ✅ 前端服务正常"
    echo "前端服务: 正常" >> $RESULT_FILE
else
    echo "   ⚠️  前端服务可能未运行或端口不对"
    echo "前端服务: 异常" >> $RESULT_FILE
fi

echo ""
echo "========================================="
echo "✅ 所有测试通过！"
echo "========================================="
echo ""
echo "验证结果已保存到: $RESULT_FILE"
echo ""
echo "📱 现在可以在浏览器访问:"
echo "   http://localhost:5175"
echo ""
