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

### Bug修复功能

- **数值溢出修复** (`/api/optimize-with-bugs-fixed`) - 专门修复NaN/Inf/负数/超大值问题
- **A/B测试对比** (`/api/ab-test-numerical-overflow`) - 对比原始Bug版本和修复版本的输出差异

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
├── fixed_optimizer.go      # 修复后的优化器（新增）
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

# 完整服务（含MOEA/D和Bug修复）
go run main.go models.go buggy_optimizer.go moead_optimizer.go fixed_optimizer.go

# Swagger文档: http://localhost:8080/swagger/index.html
```

#### ✅ 完整版本（含MOEA/D优化器和Bug修复，推荐）
### 3. 使用Swagger UI

打开浏览器访问: `http://localhost:8080/swagger/index.html`

## API端点

#### 1. 健康检查
- **路径**: `GET /api/health`

#### 2. 含Bug优化（原始版本）
- **路径**: `POST /api/optimize-with-bugs`
- **参数**: 
  - `bug_type`: Bug类型（precision_loss, numerical_overflow等）

#### 3. 含Bug优化（修复版本）⭐新增
- **路径**: `POST /api/optimize-with-bugs-fixed`
- **参数**: 
  - `bug_type`: Bug类型（当前支持: numerical_overflow）
- **功能**: 修复NaN/Inf/负数/超大值等数值异常

#### 4. 数值溢出A/B测试 ⭐新增
- **路径**: `POST /api/ab-test-numerical-overflow`
- **功能**: 对比原始Bug版本和修复版本的输出差异

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

### 含Bug优化（原始版本）
```http
POST /api/optimize-with-bugs?bug_type=numerical_overflow
Content-Type: application/json
```

支持的bug_type：
- `precision_loss`
- `numerical_overflow`
- `constraint_violation`
- `convergence_failure`
- `result_instability`

### 含Bug优化（修复版本）⭐
```http
POST /api/optimize-with-bugs-fixed?bug_type=numerical_overflow
Content-Type: application/json

{
  "count": 2,
  "ingredients": [
    {
      "id": 2,
      "name": "小麦",
      "energy": 338,
      "protein": 11.9,
      "fat": 1.3,
      "carbs": 75.2,
      "calcium": 34,
      "iron": 5.1,
      "zinc": 2.33,
      "vitamin_c": 0,
      "price": 0
    },
    {
      "id": 3,
      "name": "五谷香",
      "energy": 378,
      "protein": 9.9,
      "fat": 2.6,
      "carbs": 78.9,
      "calcium": 2,
      "iron": 0.5,
      "zinc": 0.23,
      "vitamin_c": 0,
      "price": 0
    }
  ]
}
```

**修复后响应示例**:
```json
{
  "bug_type": "numerical_overflow",
  "fixed": true,
  "description": "数值溢出Bug已修复：NaN/Inf/负数/超大值已被处理",
  "fixes": [
    "🛠️ Energy: 超大值1.80e+308被限制为5000.00",
    "🛠️ Protein: 负数值被重置为0.00",
    "🛠️ Fat: 超大值1.80e+308被限制为500.00",
    "✅ 营养素已根据修复后的用量重新计算",
    "✅ 所有数值验证通过：无NaN/Inf/负数/超大值"
  ],
  "result": {
    "nutrition": {
      "energy": 1074,
      "protein": 32.7,
      "fat": 5.85,
      "carbs": 231.15,
      "calcium": 54,
      "iron": 8.4,
      "zinc": 3.84,
      "vitamin_c": 0
    }
  }
}
```

### 数值溢出A/B测试 ⭐
```http
POST /api/ab-test-numerical-overflow
Content-Type: application/json
```

**响应包含**:
- `group_a_buggy`: 原始Bug版本结果
- `group_b_fixed`: 修复版本结果
- `comparison`: 详细对比数据
- `improvements`: 改进指标

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

## 数值计算逻辑说明 ⭐新增

### 数值溢出防护机制

修复后的优化器 (`fixed_optimizer.go`) 实现了以下数值保护机制：

#### 1. 数值保护配置 (NumericalGuard)
```go
type NumericalGuard struct {
    MaxValidValue       float64  // 最大有效值: 100000.0
    MinValidValue       float64  // 最小有效值: -0.001 (允许微小浮点误差)
    MaxIngredientWeight float64  // 最大食材重量: 500g
    MinIngredientWeight float64  // 最小食材重量: 0g
    Epsilon            float64  // 浮点精度阈值: 1e-10
}
```

#### 2. 异常值检测与修复逻辑

| 异常类型 | 检测条件 | 修复策略 |
|---------|---------|---------|
| NaN | `math.IsNaN(value)` | 重置为最小默认值 |
| 正无穷 | `math.IsInf(value, 1)` | 限制为最大默认值 |
| 负无穷 | `math.IsInf(value, -1)` | 重置为最小默认值 |
| 超大值 | `value > MaxValidValue` | 限制为最大默认值 |
| 负数 | `value < MinValidValue` | 重置为最小默认值 |
| 微小负数 | `-0.001 < value < 0` | 归零（浮点误差处理） |

#### 3. 营养素默认值配置
| 营养素 | 最小默认值 | 最大默认值 |
|-------|-----------|-----------|
| Energy | 0 | 5000 kcal |
| Protein | 0 | 500 g |
| Fat | 0 | 500 g |
| Carbs | 0 | 1000 g |
| Calcium | 0 | 5000 mg |
| Iron | 0 | 100 mg |
| Zinc | 0 | 50 mg |
| VitaminC | 0 | 1000 mg |

#### 4. 修复流程
1. **生成Bug结果**: 模拟原始Bug行为，生成含异常值的结果
2. **应用数值保护**: 检测并修复所有异常值
3. **重新计算营养素**: 基于修复后的用量重新计算
4. **验证修复结果**: 确保无NaN/Inf/负数/超大值残留

## 重量约束规则 ⭐新增

### 食材重量约束

```go
const (
    MaxIngredientWeight = 500.0  // 单食材最大500g
    MinIngredientWeight = 0.0    // 最小0g（不允许负数）
)
```

### 约束修复策略

| 场景 | 修复前 | 修复后 |
|-----|-------|-------|
| NaN/Inf用量 | `amount = NaN/Inf` | `amount = 150g` (默认值) |
| 负数用量 | `amount = -50g` | `amount = 0g` |
| 超大用量 | `amount = 1000g` | `amount = 500g` |

### 总重量约束
- 默认总重量目标: 300-400g
- 通过约束修复函数确保总重量在合理范围内
- 支持自定义约束: `{"type": "total_weight", "value": 400}`

## 算法基准验证结果 ⭐新增

### 单元测试结果

```bash
$ go test -v
=== RUN   TestMOEADOptimizer_CreateIndividual
--- PASS: TestMOEADOptimizer_CreateIndividual (0.00s)
=== RUN   TestMOEADOptimizer_CalculateNutrition
--- PASS: TestMOEADOptimizer_CalculateNutrition (0.00s)
=== RUN   TestMOEADOptimizer_CalculateTotalCost
--- PASS: TestMOEADOptimizer_CalculateTotalCost (0.00s)
=== RUN   TestMOEADOptimizer_CalculateDiversity
--- PASS: TestMOEADOptimizer_CalculateDiversity (0.00s)
=== RUN   TestMOEADOptimizer_GenerateWeightVectors
--- PASS: TestMOEADOptimizer_GenerateWeightVectors (0.00s)
=== RUN   TestMOEADOptimizer_CalculateNeighborhood
--- PASS: TestMOEADOptimizer_CalculateNeighborhood (0.00s)
=== RUN   TestMOEADOptimizer_Tchebycheff
--- PASS: TestMOEADOptimizer_Tchebycheff (0.00s)
=== RUN   TestMOEADOptimizer_Crossover
--- PASS: TestMOEADOptimizer_Crossover (0.00s)
=== RUN   TestMOEADOptimizer_Mutate
--- PASS: TestMOEADOptimizer_Mutate (0.00s)
=== RUN   TestMOEADOptimizer_Repair
--- PASS: TestMOEADOptimizer_Repair (0.00s)
=== RUN   TestMOEADOptimizer_Optimize
--- PASS: TestMOEADOptimizer_Optimize (0.00s)
=== RUN   TestRoundToTwoDecimals
--- PASS: TestRoundToTwoDecimals (0.00s)
PASS
ok      nutrient-optimizer-benchmark    0.828s
```

### A/B测试验证结果

| 指标 | Bug版本 | 修复版本 | 改进 |
|-----|--------|---------|-----|
| Energy | 1.80e+308 (溢出) | 1074 kcal | ✅ 已修复 |
| Protein | -1.80e+308 (负数溢出) | 32.7 g | ✅ 已修复 |
| Fat | 1.80e+308 (溢出) | 5.85 g | ✅ 已修复 |
| Carbs | 231.15 g | 231.15 g | ✅ 正常 |
| 数值异常 | 存在 | 无 | ✅ 100%修复 |

### 修复效果验证

```json
{
  "improvements": {
    "energy_normalized": true,
    "protein_normalized": true,
    "fat_normalized": true,
    "all_nutrients_valid": true,
    "cost_calculated": true
  }
}
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

### 数值溢出测试场景 ⭐新增
- **描述**: 专门测试数值溢出Bug修复
- **食材**: 小麦、五谷香
- **预期Bug**: Energy=最大值, Protein=最小值, Fat=最大值
- **验证修复**: 所有数值恢复正常范围

## Bug现象说明

### 1. 数值精度问题
- **现象**: 营养素计算出现偏差

### 2. 数值溢出问题
- **现象**: 返回结果中出现NaN/Inf值
- **典型特征**: 
  - Energy = 1.7976931348623157e+308 (float64最大值)
  - Protein = -1.7976931348623157e+308 (float64最小值)
  - Fat = 1.7976931348623157e+308 (float64最大值)
- **修复方案**: 数值保护机制，限制在合理范围内

### 3. 约束越界问题
- **现象**: 食材重量出现负数或超过500g的超大值

### 4. 收敛失败问题
- **现象**: 求解器无法收敛，返回空方案或极端不合理用量

### 5. 结果不稳定问题
- **现象**: 同一参数多次运行结果不一致

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
go run main.go models.go buggy_optimizer.go moead_optimizer.go fixed_optimizer.go
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

### 示例3: Bug修复API请求
```javascript
// POST /api/optimize-with-bugs-fixed?bug_type=numerical_overflow
{
  "count": 2,
  "ingredients": [
    {"id": 2, "name": "小麦", "energy": 338, "protein": 11.9, ...},
    {"id": 3, "name": "五谷香", "energy": 378, "protein": 9.9, ...}
  ]
}
```

### 示例4: A/B测试API请求
```javascript
// POST /api/ab-test-numerical-overflow
{
  "count": 2,
  "ingredients": [
    {"id": 2, "name": "小麦", "energy": 338, "protein": 11.9, ...},
    {"id": 3, "name": "五谷香", "energy": 378, "protein": 9.9, ...}
  ]
}
```

### 示例5: 获取食材列表
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
# 启动完整服务（含Bug修复）
go run main.go models.go buggy_optimizer.go moead_optimizer.go fixed_optimizer.go

# 在浏览器中打开Swagger UI
open http://localhost:8080/swagger/index.html
```

## 许可证

本项目采用MIT许可证。详见LICENSE文件。
