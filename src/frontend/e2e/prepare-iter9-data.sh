#!/bin/bash
# 迭代9 测试数据准备脚本

API_BASE="http://localhost:8080"

echo "=== 迭代9 测试数据准备 ==="

# 1. 教师登录并切换到教师分身
echo -e "\n1. 教师登录..."
TCH_LOGIN=$(curl -s -X POST "$API_BASE/api/auth/wx-login" \
  -H "Content-Type: application/json" \
  -d '{"code":"v9iter_tch_001"}')
TCH_TOKEN=$(echo "$TCH_LOGIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "   教师Token获取成功"

# 切换到教师分身(id=38)
SWITCH_RES=$(curl -s -X PUT "$API_BASE/api/personas/38/switch" \
  -H "Authorization: Bearer $TCH_TOKEN")
TCH_TOKEN=$(echo "$SWITCH_RES" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "   已切换到教师分身(ID=38)"

# 2. 学生登录并切换到学生分身
echo -e "\n2. 学生登录..."
STU_LOGIN=$(curl -s -X POST "$API_BASE/api/auth/wx-login" \
  -H "Content-Type: application/json" \
  -d '{"code":"v9iter_stu_001"}')
STU_TOKEN=$(echo "$STU_LOGIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "   学生Token获取成功"

# 切换到学生分身(id=48)
SWITCH_RES=$(curl -s -X PUT "$API_BASE/api/personas/48/switch" \
  -H "Authorization: Bearer $STU_TOKEN")
STU_TOKEN=$(echo "$SWITCH_RES" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "   已切换到学生分身(ID=48)"

# 3. 检查并建立师生关系
echo -e "\n3. 检查师生关系..."
RELATIONS=$(curl -s "$API_BASE/api/relations?status=approved" \
  -H "Authorization: Bearer $TCH_TOKEN")
RELATION_COUNT=$(echo "$RELATIONS" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('data',{}).get('items',[])))")

if [ "$RELATION_COUNT" -eq 0 ]; then
  echo "   建立师生关系..."
  
  # 创建分享码
  SHARE_RES=$(curl -s -X POST "$API_BASE/api/shares" \
    -H "Authorization: Bearer $TCH_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"persona_id":38}')
  SHARE_CODE=$(echo "$SHARE_RES" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('share_code',''))")
  echo "   分享码: $SHARE_CODE"
  
  if [ -n "$SHARE_CODE" ]; then
    # 学生加入
    curl -s -X POST "$API_BASE/api/shares/$SHARE_CODE/join" \
      -H "Authorization: Bearer $STU_TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"student_persona_id":48}' > /dev/null
    echo "   学生已申请加入"
    
    # 教师审批
    PENDING=$(curl -s "$API_BASE/api/relations?status=pending" \
      -H "Authorization: Bearer $TCH_TOKEN")
    RELATION_ID=$(echo "$PENDING" | python3 -c "import sys,json; items=json.load(sys.stdin).get('data',{}).get('items',[]); print(items[0]['id'] if items else '')")
    
    if [ -n "$RELATION_ID" ]; then
      curl -s -X PUT "$API_BASE/api/relations/$RELATION_ID/approve" \
        -H "Authorization: Bearer $TCH_TOKEN" > /dev/null
      echo "   ✅ 师生关系已建立"
    fi
  fi
else
  echo "   ✅ 师生关系已存在($RELATION_COUNT条)"
fi

# 4. 检查并创建班级
echo -e "\n4. 检查班级..."
CLASSES=$(curl -s "$API_BASE/api/classes" \
  -H "Authorization: Bearer $TCH_TOKEN")
CLASS_COUNT=$(echo "$CLASSES" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('data',{}).get('items',[])))")

if [ "$CLASS_COUNT" -eq 0 ]; then
  echo "   创建班级..."
  CLASS_RES=$(curl -s -X POST "$API_BASE/api/classes" \
    -H "Authorization: Bearer $TCH_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"name":"V9测试班级","description":"V9自动化测试班级"}')
  CLASS_ID=$(echo "$CLASS_RES" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))")
  
  if [ -n "$CLASS_ID" ]; then
    echo "   ✅ 班级已创建(ID=$CLASS_ID)"
    
    # 添加学生到班级
    curl -s -X POST "$API_BASE/api/classes/$CLASS_ID/members" \
      -H "Authorization: Bearer $TCH_TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"student_persona_id":48}' > /dev/null
    echo "   ✅ 学生已添加到班级"
  else
    echo "   ❌ 班级创建失败"
  fi
else
  echo "   ✅ 班级已存在($CLASS_COUNT个)"
fi

echo -e "\n=== 数据准备完成 ===\n"
