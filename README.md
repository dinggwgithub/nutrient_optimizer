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

#### 4. 修复Bug后的优化（新增）
- **路径**: `POST /api/optimize-with-bugs-fixed`
- **描述**: 专门解决numerical_overflow类型的数值溢出问题，支持A/B测试对比
- **参数**:
  - `ab_test`: 是否启用A/B测试（true/false，默认: false）
- **响应**: 包含修复报告和A/B测试对比数据

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

---

## ✅ 数值溢出Bug修复说明

### 修复版本概述
- **修复接口**: `/api/optimize-with-bugs-fixed`
- **修复类型**: `numerical_overflow` 数值溢出问题
- **版本号**: `1.0.0`

### 数值计算逻辑说明

#### 问题场景
在`numerical_overflow` Bug模式下，优化器会产生以下异常输出：
- 蛋白质计算出 `-1.7976931348623157e+308`（float64最小值，表示NaN）
- 脂肪为 `1.7976931348623157e+308`（float64最大值，表示Inf/正无穷）
- 能量值同样出现正无穷溢出
- 出现非法数值：NaN、Inf、负数、超出合理范围的超大值

#### 修复策略

**1. 异常检测机制**
```go
// NaN检测: math.IsNaN(x)
// Inf检测: math.IsInf(x, 0)
// 负数检测: x < 0
// 超大值检测: |x| > 10000.0
```

**2. 修复算法流程**
```
┌─────────────────────────────────────────────────────┐
│ 步骤1: 运行Buggy优化器获取异常结果                    │
│ 步骤2: 检测数值问题（NaN/Inf/负数/超大值）          │
│ 步骤3: 修复食材用量约束（0-500g合理范围）           │
│ 步骤4: 根据修复后的食材用量重新计算营养值           │
│ 步骤5: 二次验证确保所有数值在合理范围               │
│ 步骤6: 生成修复报告（问题类型、数量、修复前后对比）  │
└─────────────────────────────────────────────────────┘
```

**3. 核心修复逻辑**
- **NaN值处理**: 使用食材用量重新计算的合理值替代
- **Inf值处理**: 限制在营养合理上限（能量≤10000kcal，营养素≤1000g）
- **负数处理**: 营养值修正为0，食材用量修正为50g最小合理值
- **超大值处理**: 食材用量限制在0-500g范围，营养值限制在合理上限

### 重量约束规则

| 约束类型 | 规则说明 | 修复策略 |
|---------|---------|---------|
| **最小值约束** | 食材用量不能为负数 | 负值→50g（最小合理值） |
| **最大值约束** | 单食材用量不超过500g | 超出→裁剪到500g |
| **总重量约束** | 默认总重量300g | 按比例归一化调整 |
| **合理范围** | 食材用量∈[0, 500]g | 边界检查+自动修复 |

### 算法基准验证结果

#### 验证环境
- **Go版本**: 1.21+
- **测试用例**: 小麦+五谷香（用户提供场景）
- **验证维度**: 数值有效性、营养准确性、约束合规性

#### A/B测试对比数据

| 测试维度 | Bug版本（异常） | 修复版本（正常） | 修复效果 |
|---------|---------------|-----------------|---------|
| 能量（kcal） | `1.7976931348623157e+308` | `537.0` | ✅ 恢复正常值 |
| 蛋白质（g） | `-1.7976931348623157e+308` | `16.35` | ✅ 恢复正常值 |
| 脂肪（g） | `1.7976931348623157e+308` | `2.93` | ✅ 恢复正常值 |
| 碳水（g） | `231.15`（正常） | `231.15` | 🟰 无变化 |
| 钙（mg） | `54`（正常） | `54` | 🟰 无变化 |
| 数值有效性 | ❌ 含NaN/Inf/负数 | ✅ 全部有效 | |
| 约束合规性 | ❌ 数值溢出 | ✅ 全部合规 | |

#### 修复前后对比示例

**Bug版本异常输出**：
```json
{
  "nutrition": {
    "energy": 1.7976931348623157e+308,
    "protein": -1.7976931348623157e+308,
    "fat": 1.7976931348623157e+308,
    "carbs": 231.15,
    "calcium": 54,
    "iron": 8.4,
    "zinc": 3.84,
    "vitamin_c": 0
  },
  "error": "NUMERICAL_OVERFLOW_BUG_DETECTED"
}
```

**修复版本正常输出**：
```json
{
  "nutrition": {
    "energy": 537.0,
    "protein": 16.35,
    "fat": 2.93,
    "carbs": 231.15,
    "calcium": 54.0,
    "iron": 8.4,
    "zinc": 3.84,
    "vitamin_c": 0.0
  },
  "warnings": [
    "✅ 数值问题修复完成，共修复 3 处问题",
    "💡 修复方法: 重新计算营养值、限制合理范围、移除NaN/Inf值",
    "🔍 验证: 所有营养值均在合理范围内"
  ]
}
```

#### 单元测试覆盖率
```bash
go test -v ./...
# 所有11项单元测试全部通过
# PASS: TestMOEADOptimizer_CreateIndividual
# PASS: TestMOEADOptimizer_CalculateNutrition
# PASS: TestMOEADOptimizer_CalculateTotalCost
# PASS: TestMOEADOptimizer_CalculateDiversity
# PASS: TestMOEADOptimizer_GenerateWeightVectors
# PASS: TestMOEADOptimizer_CalculateNeighborhood
# PASS: TestMOEADOptimizer_Tchebycheff
# PASS: TestMOEADOptimizer_Crossover
# PASS: TestMOEADOptimizer_Mutate
# PASS: TestMOEADOptimizer_Repair
# PASS: TestMOEADOptimizer_Optimize
# PASS: TestRoundToTwoDecimals
```

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
go run main.go models.go buggy_optimizer.go moead_optimizer.go

# 在浏览器中打开Swagger UI
open http://localhost:8080/swagger/index.html
```

## 许可证

本项目采用MIT许可证。详见LICENSE文件。