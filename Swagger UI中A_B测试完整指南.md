## 🧪 Swagger UI中A/B测试完整指南

### 📋 准备工作

**步骤1：启动完整服务**
```bash
# 启动含MOEA/D的完整服务
go run main.go models.go buggy_optimizer.go moead_optimizer.go
```

**步骤2：访问Swagger UI**
```
浏览器打开: http://localhost:8080/swagger/index.html
```

---

### 🎯 核心A/B测试场景

#### 场景1：基准版本 vs 含Bug版本

| 端点A（基准） | 端点B（含Bug） | 对比目的 |
|--------------|----------------|----------|
| `POST /api/optimize` | `POST /api/optimize-with-bugs` | 验证Bug的存在，观察异常现象 |

#### 场景2：含Bug版本 vs MOEA/D版本

| 端点A（含Bug） | 端点B（MOEA/D） | 对比目的 |
|----------------|-----------------|----------|
| `POST /api/optimize-with-bugs` | `POST /api/optimize-moead` | 验证MOEA/D算法的改进效果 |

#### 场景3：数据源切换测试

| 参数A（数据库） | 参数B（JSON文件） | 对比目的 |
|-----------------|-------------------|----------|
| `GET /api/ingredients?source=db` | `GET /api/ingredients?source=json` | 验证无数据库降级功能 |

---

### 📝 详细操作步骤

#### 🔧 测试案例1：验证"约束越界"Bug

**第一步：调用含Bug版本，观察异常**
```
1. 在Swagger中找到: POST /api/optimize-with-bugs
2. 点击: Try it out
3. 设置 bug_type = constraint_violation
4. 输入请求体（使用test_cases.json中常规场景）
5. 点击: Execute
```

🔍 **预期Bug现象（异常）：**
```json
{
  "ingredients": [
    {"name": "鸡胸肉", "amount": -50.5},   // ❌ 负数！
    {"name": "米饭", "amount": 650.0},     // ❌ 超过500g！
    {"name": "西兰花", "amount": 0.0}      // ❌ 0用量
  ],
  "converged": false    // ❌ 未收敛
}
```

**第二步：调用基准版本，观察正常输出**
```
1. 在Swagger中找到: POST /api/optimize
2. 点击: Try it out
3. 输入**完全相同**的请求体
4. 点击: Execute
```

✅ **预期正常结果：**
```json
{
  "ingredients": [
    {"name": "鸡胸肉", "amount": 85.2},    // ✅ 合理正数
    {"name": "米饭", "amount": 180.5},     // ✅ 在合理范围
    {"name": "西兰花", "amount": 65.3},    // ✅ 多样化
    {"name": "鸡蛋", "amount": 69.0}
  ],
  "nutrition": {
    "energy": 598.5,    // ✅ 接近目标600kcal
    "protein": 29.8     // ✅ 接近目标30g
  },
  "converged": true     // ✅ 成功收敛
}
```

---

### 📊 5类Bug的A/B测试对比表

| Bug类型 | 含Bug版本（异常表现） | 基准/修复版本（正常表现） |
|---------|----------------------|--------------------------|
| **浮点数精度丢失** | 能量: 450 kcal（目标600）<br>偏差>25% | 能量: 598 kcal<br>偏差<1% |
| **数值溢出** | 蛋白质: NaN 或 Infinity | 蛋白质: 29.8 g |
| **约束越界** | 食材用量: -50g 或 650g | 食材用量: 0 < x < 500g |
| **收敛失败** | converged: false<br>返回空方案或极端值 | converged: true<br>返回合理配餐 |
| **结果不稳定** | 多次调用能量值波动:<br>450 → 700 → 300 | 多次调用能量值稳定:<br>598 → 597 → 599 |

---

### 💡 A/B测试高级技巧

#### 技巧1：请求体复用（关键！）

```
1. 第一次调用后，复制Swagger UI中的Request Body
2. 在另一个端点测试时，粘贴完全相同的Body
3. 确保两次测试参数完全一致！
```

#### 技巧2：多标签页对比

```
1. 打开两个浏览器标签页，都访问Swagger UI
2. 标签A: 测试含Bug版本
3. 标签B: 测试基准/修复版本
4. 左右分屏对比结果
```

#### 技巧3：MOEA/D算法特殊测试

```
测试: POST /api/optimize-moead
请求体可以简化（自动加载JSON食材）：
{
  "population_size": 30,
  "max_iterations": 100,
  "nutrition_goals": [...],   // 仅需营养目标
  "constraints": [...],       // 仅需约束条件
  "weights": [...]
}
```
✅ **自动加载200种食材数据**，无需在请求中提供！

---

### 🎯 测试成功的判断标准

在Swagger UI中观察到**以下所有变化**即为修复/改进有效：

| 检查维度 | 通过标准 |
|----------|----------|
| **格式正确性** | 返回有效的JSON，无语法错误 |
| **数值有效性** | 无NaN、无Inf、无负数 |
| **范围合理性** | 食材用量: 0 < x < 500g<br>营养值: 接近目标值±10% |
| **收敛状态** | converged: true |
| **结果稳定性** | 3次连续调用结果差异 < 5% |
| **约束满足** | 总重量接近目标值（如400g） |

---

### ❌ 常见问题排查

| 现象 | 可能原因 | 解决方法 |
|------|----------|----------|
| Swagger显示"No operations defined in spec!" | 服务未完全启动 | 等待10秒后刷新 |
| 调用后显示"ERROR" | 端口冲突或依赖缺失 | 检查8080端口，执行`go mod tidy` |
| 返回500错误 | 请求体格式错误 | 对照Schema检查JSON格式 |
| MOEA/D返回空食材 | ingredients_db_export.json不存在 | 确认文件在项目根目录 |

这样，您就可以在Swagger UI中完成专业的A/B测试了！需要我演示某个具体Bug类型的测试流程吗？