# 营养配餐多目标优化算法Bug测试套件

## 项目概述

本项目提供营养配餐多目标优化算法测试框架，包含5类典型科学计算Bug的复现机制。

## 核心特性

### 5类典型科学计算Bug

1. **浮点数精度丢失** - 营养素计算偏差
2. **数值溢出** - 返回NaN/Inf结果
3. **约束越界** - 食材重量出现负数或超大值
4. **收敛失败** - 求解器无法收敛
5. **结果不稳定** - 多次运行结果不一致

### 技术栈

- **后端框架**: Go + Gin
- **API文档**: Swagger UI
- **优化算法**: 
  - 加权求和多目标优化
  - MOEA/D多目标进化算法
- **数据存储**: JSON文件 / MySQL数据库

## 项目结构

```
nutrient-optimizer-benchmark/
├── main.go                  # 主程序入口
├── models.go               # 数据模型
├── buggy_optimizer.go      # 含Bug的优化器
├── moead_optimizer.go      # MOEA/D优化器
├── moead_optimizer_test.go # 单元测试
├── ingredients.json       # 15种食材数据
├── ingredients_db_export.json # 200种食材数据
├── test_cases.json        # 测试用例
├── docs/                  # Swagger文档
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
# 仅原优化器
go run main.go models.go buggy_optimizer.go

# 完整服务（含MOEA/D）
go run main.go models.go buggy_optimizer.go moead_optimizer.go

# Swagger文档: http://localhost:8080/swagger/index.html
```

#### ✅ 完整版本（含MOEA/D优化器，推荐）
### 3. 使用Swagger UI

打开浏览器访问: `http://localhost:8080/swagger/index.html`

## API端点

#### 1. 健康检查
- **路径**: `GET /api/health`

#### 2. 正常优化
- **路径**: `POST /api/optimize`

#### 3. 含Bug优化
- **路径**: `POST /api/optimize-with-bugs`
- **参数**: 
  - `bug_type`: Bug类型（precision_loss, numerical_overflow等）

#### 4. MOEA/D多目标优化
- **路径**: `POST /api/optimize-moead`
- **参数**:
  - `population_size`: 种群大小（可选，默认: 50）
  - `max_iterations`: 最大迭代次数（可选，默认: 100）

#### 5. 修复后的优化（Bug修复版本）
- **路径**: `POST /api/optimize-with-bugs-fixed`
- **描述**: 解决结果不稳定和约束越界两类核心Bug的修复版本
- **修复内容**:
  - 结果稳定性：使用固定随机种子(42)，确保相同输入多次调用结果一致
  - 约束合规：食材用量限制在0-500g区间，消除负数和超大值

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

### 正常优化
```http
POST /api/optimize
Content-Type: application/json

{
  "ingredients": [...],
  "nutrition_goals": [...],
  "constraints": [...],
  "weights": [...],
  "max_iterations": 1000,
  "tolerance": 1e-6
}
```

### 含Bug优化
```http
POST /api/optimize-with-bugs?bug_type=precision_loss
Content-Type: application/json
```

支持的bug_type：
- `precision_loss`
- `numerical_overflow`
- `constraint_violation`
- `convergence_failure`
- `result_instability`

### MOEA/D多目标优化
```http
POST /api/optimize-moead
Content-Type: application/json

{
  "population_size": 50,
  "max_iterations": 100,
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
    // 更多食材...
  ],
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

## Bug现象说明

### 1. 数值精度问题
- **现象**: 营养素计算出现偏差

### 2. 数值溢出问题
- **现象**: 返回结果中出现NaN/Inf值

### 3. 约束越界问题
- **现象**: 食材重量出现负数或超过500g的超大值

### 4. 收敛失败问题
- **现象**: 求解器无法收敛，返回空方案或极端不合理用量

### 5. 结果不稳定问题
- **现象**: 同一参数多次运行结果不一致

## Bug修复方案

### 结果稳定性修复

**问题原因**: 使用`time.Now().UnixNano()`作为随机种子，每次运行种子不同导致结果不一致。

**修复方案**: 使用固定随机种子(42)，确保相同输入多次调用返回完全一致的结果。

```go
const FixedRandomSeed int64 = 42

func (o *FixedOptimizer) Optimize(req OptimizationRequest) (*OptimizationResult, error) {
    rand.Seed(FixedRandomSeed)  // 固定随机种子
    // ...
}
```

### 约束越界修复

**问题原因**: 优化过程中未对食材用量进行边界约束，导致出现负数或超大值。

**修复方案**: 使用`clampAmount`函数强制限制食材用量在0-500g区间。

```go
const (
    MinIngredientAmount float64 = 0     // 最小用量
    MaxIngredientAmount float64 = 500   // 最大用量
)

func (o *FixedOptimizer) clampAmount(amount float64) float64 {
    return math.Max(MinIngredientAmount, math.Min(MaxIngredientAmount, amount))
}
```

### 食材重量约束规则

| 约束类型 | 说明 | 默认值 |
|---------|------|--------|
| 最小用量 | 单个食材最小用量 | 0g |
| 最大用量 | 单个食材最大用量 | 500g |
| 总重量 | 所有食材总重量 | 300g |

### 营养素目标约束校验逻辑

```go
// 营养素目标校验
for _, goal := range req.NutritionGoals {
    if goal.Min > 0 && actual < goal.Min {
        // 添加警告：低于最小目标
    }
    if goal.Max > 0 && actual > goal.Max {
        // 添加警告：超过最大目标
    }
}
```

## A/B测试验证

### 测试用例1: 结果稳定性测试

```bash
# 多次调用含Bug版本（结果不一致）
curl -X POST 'http://localhost:8080/api/optimize-with-bugs?bug_type=result_instability' \
  -H 'Content-Type: application/json' \
  -d '{"count": 2, "ingredients": [...]}'

# 多次调用修复版本（结果一致）
curl -X POST 'http://localhost:8080/api/optimize-with-bugs-fixed' \
  -H 'Content-Type: application/json' \
  -d '{"count": 2, "ingredients": [...]}'
```

### 测试用例2: 约束边界测试

```bash
# 含Bug版本（可能出现负数/超大值）
curl -X POST 'http://localhost:8080/api/optimize-with-bugs?bug_type=constraint_violation' \
  -H 'Content-Type: application/json' \
  -d '{"ingredients": [...]}'

# 修复版本（用量在0-500g区间）
curl -X POST 'http://localhost:8080/api/optimize-with-bugs-fixed' \
  -H 'Content-Type: application/json' \
  -d '{"ingredients": [...]}'
```

### 验证标准

| 指标 | 含Bug版本 | 修复版本 |
|-----|----------|---------|
| 结果稳定性 | 多次调用结果不一致 | 多次调用结果完全一致 |
| 食材用量 | 可能出现负数或超大值 | 严格限制在0-500g区间 |
| 警告信息 | 包含约束越界警告 | 警告字段清空或仅含营养提示 |
| 收敛状态 | 可能不稳定 | converged=true稳定 |

## 数据存储

项目使用JSON文件存储数据，无需数据库配置：

### ingredients.json
包含15种常用食材的营养成分数据：
- 鸡胸肉、西兰花、米饭、鸡蛋
- 黄瓜、西红柿、生菜、牛肉
- 三文鱼、豆腐、胡萝卜、苹果
- 燕麦、牛奶、菠菜

每种食材包含：能量、蛋白质、脂肪、碳水化合物、钙、铁、锌、维生素C、价格等字段。

### ✅ ingredients_db_export.json
从数据库导出的200种真实食材数据（新增）：
- 从recipe_system数据库导出
- 包含小麦、五谷香、各种谷物、面条、馒头、面包、饼干等200种食材
- **MOEA/D优化器默认使用此文件**
- 无数据库环境下的完整食材库

### test_cases.json
包含3个场景：
- 常规场景
- 边界场景
- 极端数值场景

## MOEA/D算法使用示例

### 示例1: 启动服务
```bash
go run main.go models.go buggy_optimizer.go moead_optimizer.go
```

### 示例2: MOEA/D API请求
```javascript
// POST /api/optimize-moead
{
  "population_size": 30,
  "max_iterations": 100,
  "nutrition_goals": [
    {"nutrient": "energy", "target": 600, "min": 500, "max": 700, "weight": 0.3},
    {"nutrient": "protein", "target": 30, "min": 20, "max": 40, "weight": 0.4}
  ],
  "constraints": [
    {"type": "total_weight", "value": 400}
  ],
  "weights": [
    {"type": "nutrition", "value": 0.6},
    {"type": "cost", "value": 0.3},
    {"type": "variety", "value": 0.1}
  ]
}
```

### 示例3: 获取食材列表
```http
GET /api/ingredients?source=json&limit=10
```

## 扩展使用

### 运行单元测试
```bash
# 运行所有单元测试
go test -v ./...

# 仅运行MOEA/D优化器测试
go test -v -run MOEAD
```

### 启动服务

```bash
# 启动完整服务
go run main.go models.go buggy_optimizer.go moead_optimizer.go fixed_optimizer.go

# 在浏览器中打开Swagger UI
open http://localhost:8080/swagger/index.html
```

## 许可证

本项目采用MIT许可证。详见LICENSE文件。