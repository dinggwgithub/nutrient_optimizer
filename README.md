# 营养配餐多目标优化算法Bug测试套件

## 项目概述

本项目提供营养配餐多目标优化算法测试框架，包含5类典型科学计算Bug的复现机制和修复方案。

## 核心特性

### 5类典型科学计算Bug

1. **浮点数精度丢失** - 营养素计算偏差
2. **数值溢出** - 返回NaN/Inf结果
3. **约束越界** - 食材重量出现负数或超大值
4. **收敛失败** - 求解器无法收敛
5. **结果不稳定** - 多次运行结果不一致

### 修复特性

1. **数值计算修复** - 使用float64确保精度，避免精度丢失
2. **约束验证** - 食材重量限制在0-500g范围内
3. **收敛保证** - 设置5%收敛阈值，确保结果稳定
4. **结果验证** - 自动检测NaN/Inf/负数/超大值等异常

### 技术栈

- **后端框架**: Go + Gin
- **API文档**: Swagger UI
- **优化算法**: 
  - 加权求和多目标优化（修复版本）
  - MOEA/D多目标进化算法
- **数据存储**: JSON文件 / MySQL数据库

## 项目结构

```
nutrient-optimizer-benchmark/
├── main.go                     # 主程序入口
├── models.go                   # 数据模型
├── buggy_optimizer.go          # 含Bug的优化器（测试用）
├── fixed_optimizer.go          # 修复后的优化器
├── moead_optimizer.go          # MOEA/D优化器
├── moead_optimizer_test.go     # MOEA/D单元测试
├── fixed_optimizer_test.go     # 修复版本单元测试
├── ingredients.json            # 15种食材数据
├── ingredients_db_export.json  # 200种食材数据
├── test_cases.json             # 测试用例
├── docs/                       # Swagger文档
│   └── docs.go
├── go.mod
└── README.md
```

## 快速开始

### 1. 环境准备

```bash
go mod tidy
```

### 2. 启动服务

```bash
# 完整服务（含MOEA/D和修复版本）
go run .

# 或编译后运行
go build -o optimizer.exe .
optimizer.exe

# Swagger文档: http://localhost:8080/swagger/index.html
```

### 3. 运行单元测试

```bash
# 运行所有测试
go test -v .

# 运行特定测试
go test -v -run TestFixedOptimizer
go test -v -run TestMOEADOptimizer
```

## API端点

#### 1. 健康检查
- **路径**: `GET /api/health`

#### 2. 修复后优化（推荐）
- **路径**: `POST /api/optimize-fixed`
- **描述**: 使用修复后的优化器，无Bug

#### 3. 含Bug优化（测试用）
- **路径**: `POST /api/optimize-with-bugs`
- **参数**: 
  - `bug_type`: Bug类型（precision_loss, numerical_overflow等）

#### 4. A/B测试
- **路径**: `POST /api/ab-test`
- **描述**: 对比Bug版本和修复版本的结果

#### 5. MOEA/D多目标优化
- **路径**: `POST /api/optimize-moead`
- **参数**:
  - `population_size`: 种群大小（可选，默认: 50）
  - `max_iterations`: 最大迭代次数（可选，默认: 100）

#### 6. 获取食材列表
- **路径**: `GET /api/ingredients`
- **参数**:
  - `source`: 数据源（可选，db或json，默认: db）
  - `limit`: 返回数量限制（可选，默认: 20）

## API接口详情

### 健康检查
```http
GET /api/health
```

### 修复后优化（推荐）
```http
POST /api/optimize-fixed
Content-Type: application/json

{
  "ingredients": [
    {"id": 2, "name": "小麦", "energy": 338, "protein": 11.9, "fat": 1.3, "carbs": 75.2, "calcium": 34, "iron": 5.1, "zinc": 2.33, "vitamin_c": 0, "price": 0},
    {"id": 3, "name": "五谷香", "energy": 378, "protein": 9.9, "fat": 2.6, "carbs": 78.9, "calcium": 2, "iron": 0.5, "zinc": 0.23, "vitamin_c": 0, "price": 0}
  ],
  "constraints": [
    {"type": "total_weight", "value": 300}
  ],
  "max_iterations": 100,
  "tolerance": 0.05
}
```

### 含Bug优化
```http
POST /api/optimize-with-bugs?bug_type=precision_loss
Content-Type: application/json
```

支持的bug_type：
- `precision_loss` - 浮点数精度丢失
- `numerical_overflow` - 数值溢出
- `constraint_violation` - 约束越界
- `convergence_failure` - 收敛失败
- `result_instability` - 结果不稳定

### A/B测试
```http
POST /api/ab-test
Content-Type: application/json

{
  "bug_type": "precision_loss",
  "request": {
    "ingredients": [...],
    "constraints": [...]
  }
}
```

响应示例：
```json
{
  "bug_type": "precision_loss",
  "buggy_version": {
    "success": true,
    "result": {...},
    "warnings": [...]
  },
  "fixed_version": {
    "success": true,
    "result": {...},
    "warnings": []
  },
  "analysis": {
    "energy_diff": 6085.84,
    "energy_diff_pct": 566.65,
    "calcium_diff": 9999946.00,
    "calcium_diff_pct": 185184185.19,
    "issues_detected": [
      "Bug版本能量值异常",
      "Bug版本钙含量异常",
      "修复版本数值计算正常"
    ]
  }
}
```

### MOEA/D多目标优化
```http
POST /api/optimize-moead
Content-Type: application/json

{
  "population_size": 50,
  "max_iterations": 100,
  "ingredients": [...],
  "nutrition_goals": [...],
  "constraints": [...],
  "weights": [...]
}
```

### 获取食材列表
```http
GET /api/ingredients?limit=20
GET /api/ingredients?source=json&limit=20
```

## Bug修复说明

### 1. 浮点数精度丢失修复

**问题原因**:
- 使用float32进行浮点运算，导致精度丢失
- 多次迭代累加放大精度问题
- 小数值被大数淹没

**修复方案**:
- 使用float64进行所有数值计算
- 避免不必要的累加操作
- 使用Kahan求和算法（可选）

**修复效果**:
```
Bug版本:  能量=7159.84 kcal, 钙=10000000.00 mg
修复版本: 能量=1074.00 kcal, 钙=54.00 mg
期望值:   能量=1074.00 kcal, 钙=54.00 mg
```

### 2. 数值溢出修复

**问题原因**:
- 除零操作导致Inf
- 非法操作导致NaN
- 指数运算溢出

**修复方案**:
- 添加除零检查
- 验证数值范围
- 使用math.IsNaN和math.IsInf检测异常值

### 3. 约束越界修复

**问题原因**:
- 食材重量出现负数
- 食材重量超过500g
- 约束条件未正确应用

**修复方案**:
- 添加食材重量约束（0-500g）
- 使用math.Max和math.Min限制范围
- 约束修复机制确保可行性

### 4. 收敛失败修复

**问题原因**:
- 收敛阈值设置过严
- 初始可行解构造不当
- 迭代次数不足

**修复方案**:
- 设置合理的收敛阈值（5%）
- 改进初始解生成
- 增加最大迭代次数

### 5. 结果不稳定修复

**问题原因**:
- 随机数种子未固定
- 优化过程受随机因素影响
- 目标函数计算不一致

**修复方案**:
- 固定随机数种子（可选）
- 多次运行取平均
- 结果验证机制

## 测试用例

### 常规场景
- **描述**: 正常营养配餐场景
- **食材**: 鸡胸肉、西兰花、米饭、鸡蛋
- **营养目标**: 能量600kcal，蛋白质30g

### 边界场景（低热量目标）
- **描述**: 极低热量目标
- **食材**: 黄瓜、西红柿、生菜
- **营养目标**: 能量100kcal，蛋白质5g

### 极端数值场景
- **描述**: 极端数值情况
- **食材**: 高能量、高蛋白、高微量营养素食材
- **营养目标**: 能量2000kcal，蛋白质100g

## 单元测试结果

```
=== RUN   TestFixedOptimizer_Optimize
    fixed_optimizer_test.go:72: 优化结果: 能量=1074.00 kcal, 钙=54.00 mg
    fixed_optimizer_test.go:74:   小麦: 150.00g
    fixed_optimizer_test.go:74:   五谷香: 150.00g
--- PASS: TestFixedOptimizer_Optimize (0.00s)

=== RUN   TestFixedOptimizer_CalculateNutrition
    fixed_optimizer_test.go:103: 营养计算: 能量=1074.00 kcal, 钙=54.00 mg
--- PASS: TestFixedOptimizer_CalculateNutrition (0.00s)

=== RUN   TestFixedOptimizer_NoPrecisionLoss
    fixed_optimizer_test.go:132: 多次累加测试: 能量=1000.00, 钙=10.00
--- PASS: TestFixedOptimizer_NoPrecisionLoss (0.00s)

=== RUN   TestFixedOptimizer_IngredientConstraints
    fixed_optimizer_test.go:177: 总重量: 399.99g
--- PASS: TestFixedOptimizer_IngredientConstraints (0.00s)

=== RUN   TestFixedOptimizer_NoRandomness
    fixed_optimizer_test.go:262: 多次运行能量结果: [448.58, 448.58, 448.58, 448.58, 448.58]
    fixed_optimizer_test.go:263: 均值: 448.58, 标准差: 0.00
--- PASS: TestFixedOptimizer_NoRandomness (0.00s)

=== RUN   TestMOEADOptimizer_Optimize
    moead_optimizer_test.go:280: 优化结果: 能量=533.90 kcal, 成本=0.90 元
    moead_optimizer_test.go:281: 食材数量: 3
    moead_optimizer_test.go:283:   鸡胸肉: 69.03g
    moead_optimizer_test.go:283:   西兰花: 10.69g
    moead_optimizer_test.go:283:   米饭: 320.28g
--- PASS: TestMOEADOptimizer_Optimize (0.00s)

PASS
ok      nutrient-optimizer-benchmark    0.960s
```

## 数据存储

### JSON文件格式

```json
{
  "ingredients": [
    {
      "id": 1,
      "name": "鸡胸肉",
      "energy": 165,
      "protein": 31,
      "fat": 3.6,
      "carbs": 0,
      "calcium": 11,
      "iron": 1,
      "zinc": 0.7,
      "vitamin_c": 0,
      "price": 0.8
    }
  ]
}
```

### 数据库表结构

```sql
CREATE TABLE ingredients (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    energy FLOAT,
    protein FLOAT,
    fat FLOAT,
    carbs FLOAT,
    calcium FLOAT,
    iron FLOAT,
    zinc FLOAT,
    vitamin_c FLOAT,
    price FLOAT,
    status VARCHAR(20) DEFAULT 'ACTIVE'
);
```

## 性能基准

### MOEA/D算法性能
- **种群大小**: 50
- **最大迭代次数**: 100
- **平均执行时间**: < 100ms
- **收敛率**: > 95%

### 修复版本性能
- **最大迭代次数**: 100
- **收敛阈值**: 5%
- **平均执行时间**: < 10ms
- **结果稳定性**: 100%

## 许可证

MIT License
