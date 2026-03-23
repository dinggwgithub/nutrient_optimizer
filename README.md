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
├── buggy_optimizer.go      # 含Bug的优化器（测试用）
├── moead_optimizer.go      # MOEA/D优化器
├── moead_optimizer_test.go # MOEA/D单元测试
├── fixed_optimizer.go      # Bug修复版优化器（推荐）
├── fixed_optimizer_test.go # Bug修复版单元测试
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
# 完整服务（含MOEA/D和Bug修复版优化器）
go run main.go models.go buggy_optimizer.go moead_optimizer.go fixed_optimizer.go

# Swagger文档: http://localhost:8080/swagger/index.html
```

#### ✅ 完整版本（含MOEA/D优化器和Bug修复版，推荐）
### 3. 使用Swagger UI

打开浏览器访问: `http://localhost:8080/swagger/index.html`

## API端点

#### 1. 健康检查
- **路径**: `GET /api/health`

#### 2. 正常优化
- **路径**: `POST /api/optimize`

#### 3. 含Bug优化（测试版本）
- **路径**: `POST /api/optimize-with-bugs`
- **参数**: 
  - `bug_type`: Bug类型（precision_loss, numerical_overflow, constraint_violation, convergence_failure, result_instability）

#### 4. Bug修复版优化（推荐）
- **路径**: `POST /api/optimize-with-bugs-fixed`
- **描述**: 修复了结果不稳定和约束越界两类Bug的优化器
- **修复内容**:
  - 结果不稳定：使用固定随机种子（seed=42），确保相同输入多次调用结果完全一致
  - 约束越界：食材用量严格限制在0-500g范围内，无负数或超大值

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

### 含Bug优化（测试版本）
```http
POST /api/optimize-with-bugs?bug_type=precision_loss
Content-Type: application/json
```

支持的bug_type：
- `precision_loss` - 浮点数精度丢失
- `numerical_overflow` - 数值溢出
- `constraint_violation` - 约束越界（食材用量负数或超大值）
- `convergence_failure` - 收敛失败
- `result_instability` - 结果不稳定（多次运行结果不一致）

### Bug修复版优化（推荐）
```http
POST /api/optimize-with-bugs-fixed
Content-Type: application/json

{
  "ingredients": [
    {"id": 2, "name": "小麦", "energy": 338, "protein": 11.9, ...},
    {"id": 3, "name": "五谷香", "energy": 378, "protein": 9.9, ...}
  ],
  "constraints": [
    {"type": "total_weight", "value": 400}
  ]
}
```

**修复内容**：
- ✅ **结果稳定性**：使用固定随机种子（seed=42），相同输入多次调用结果完全一致
- ✅ **约束合规**：食材用量严格限制在0-500g范围内
- ✅ **异常消除**：warnings字段清空，converged=true稳定

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

## Bug修复说明

### 修复的Bug类型

#### 1. 结果不稳定Bug（result_instability）
**问题原因**：
- 使用 `time.Now().UnixNano()` 作为随机种子，每次运行产生不同的随机数序列
- 随机扰动直接影响优化结果

**修复方案**：
```go
// 修复前：使用当前时间作为随机种子（不稳定）
rand.Seed(time.Now().UnixNano())

// 修复后：使用固定随机种子（稳定）
fixedSeed := int64(42)  // 固定种子值
rand.Seed(fixedSeed)
```

**验证结果**：
- ✅ 相同输入多次调用，食材用量完全一致
- ✅ 相同输入多次调用，营养值完全一致

#### 2. 约束越界Bug（constraint_violation）
**问题原因**：
- 优化过程中未严格限制食材用量范围
- 变异和交叉操作可能产生负数或超过500g的用量

**修复方案**：
```go
// 食材重量约束规则（0-500g）
const MinAmount = 0.0
const MaxAmount = 500.0

// 在变异、交叉、修复等操作中强制约束
func (o *FixedOptimizer) repair(ind *Individual, req OptimizationRequest) {
    // 确保用量在有效范围内
    for i := range ind.Amounts {
        ind.Amounts[i] = math.Max(MinAmount, math.Min(MaxAmount, ind.Amounts[i]))
    }
    // ... 其他修复逻辑
}

// 最终结果验证
func (o *FixedOptimizer) validateAndFixConstraints(result *OptimizationResult) *OptimizationResult {
    for i := range result.Ingredients {
        if result.Ingredients[i].Amount < 0 {
            result.Ingredients[i].Amount = 0
        }
        if result.Ingredients[i].Amount > 500 {
            result.Ingredients[i].Amount = 500
        }
    }
    return result
}
```

**验证结果**：
- ✅ 所有食材用量在0-500g范围内
- ✅ 无负数用量
- ✅ 无超大值用量

### 营养素目标约束校验逻辑

优化器在处理营养素目标时，使用以下逻辑进行约束校验：

```go
// 计算营养目标偏差
calculateNutritionDeviation(nutrition NutritionSummary, goals []NutritionGoal) float64

// 对于每个营养素目标：
// 1. 获取实际值
actual := getActualNutritionValue(nutrition, goal.Nutrient)

// 2. 计算相对偏差
if goal.Target > 0 {
    relDiff := math.Abs(actual - goal.Target) / goal.Target
    deviation += relDiff * goal.Weight
}

// 3. 优化目标是最小化总偏差
```

### A/B测试验证

| 测试项目 | 含Bug版本 | 修复后版本 | 验证结果 |
|---------|----------|-----------|---------|
| 结果稳定性 | 多次调用结果不一致 | 多次调用结果完全一致 | ✅ 通过 |
| 约束合规 | 出现负数/超大值 | 所有用量在0-500g | ✅ 通过 |
| 异常消除 | warnings非空 | warnings清空 | ✅ 通过 |
| 收敛状态 | converged不稳定 | converged=true稳定 | ✅ 通过 |

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

# 仅运行Bug修复版优化器测试
go test -v -run FixedOptimizer
```

### 新增单元测试说明

修复版优化器包含以下测试用例：

| 测试用例 | 描述 | 验证内容 |
|---------|------|---------|
| `TestFixedOptimizer_Stability` | 结果稳定性测试 | 同一参数多次调用，食材用量、营养值完全一致 |
| `TestFixedOptimizer_ConstraintBounds` | 约束边界测试 | 所有食材用量在0-500g范围内 |
| `TestFixedOptimizer_NoWarnings` | 异常消除测试 | warnings字段清空，converged=true稳定 |
| `TestFixedOptimizer_RepairNegativeAmount` | 负数用量修复测试 | 负数用量被修复为0 |
| `TestFixedOptimizer_RepairExcessAmount` | 超大用量修复测试 | 超过500g的用量被修复到500g以内 |
| `TestFixedOptimizer_FixedSeed` | 固定随机种子测试 | 不同优化器实例结果一致 |
| `TestFixedOptimizer_CompleteFlow` | 完整流程测试 | 端到端优化流程验证 |

### 启动服务

```bash
# 启动完整服务
go run main.go models.go buggy_optimizer.go moead_optimizer.go

# 在浏览器中打开Swagger UI
open http://localhost:8080/swagger/index.html
```

## 许可证

本项目采用MIT许可证。详见LICENSE文件。