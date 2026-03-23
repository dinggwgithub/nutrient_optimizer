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
# 仅原优化器（含Bug）
go run main.go models.go buggy_optimizer.go

# 完整服务（含MOEA/D + 修复版优化器，推荐）
go run main.go models.go buggy_optimizer.go moead_optimizer.go fixed_optimizer.go

# Swagger文档: http://localhost:8080/swagger/index.html
```

#### ✅ 完整版本（含MOEA/D + 修复版优化器，推荐）
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

#### 5. 修复版优化接口 ✅
- **路径**: `POST /api/optimize-with-bugs-fixed`
- **描述**: 修复了结果不稳定和约束越界问题的优化器
- **参数**:
  - `population_size`: 种群大小（可选，默认: 50）
  - `max_iterations`: 最大迭代次数（可选，默认: 100）
- **修复特性**:
  - ✅ 固定随机种子，确保同一参数多次调用结果一致
  - ✅ 强制执行食材重量约束（0-500g范围内）
  - ✅ 约束冲突预校验，避免无可行解情况
  - ✅ 收敛状态稳定，无异常警告

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

## 🔧 Bug修复技术说明

### 1. 算法随机种子固定方案
**问题根源**: 
原优化器使用 `rand.Seed(time.Now().UnixNano())` 作为随机种子，导致每次运行结果不同。

**修复方案**:
```go
// FixedOptimizer 使用固定随机种子42，确保结果可重复
type FixedOptimizer struct {
    randomSeed int64 // = 42，固定随机种子
    // ...
}

// 在所有随机操作中使用独立的随机数生成器
rng := rand.New(rand.NewSource(o.randomSeed + int64(generation)))
```

**技术细节**:
- 种子值：使用固定值 `42`（银河系漫游指南中"生命、宇宙以及任何事情的终极答案"）
- 确定性：同一输入参数永远产生相同输出
- 可重复性：便于调试和测试

### 2. 食材重量约束规则（0-500g）
**约束层级**:
```
┌─────────────────────────────────────────────────┐
│ 约束优先级（从高到低）                           │
├─────────────────────────────────────────────────┤
│ 1. 用户指定约束（如：ingredient_min = 100g）     │
│ 2. 通用边界约束（0g ≤ 用量 ≤ 500g）              │
│ 3. 总重量约束（如：total_weight = 400g）         │
└─────────────────────────────────────────────────┘
```

**执行步骤**:
1. **初始化阶段**：创建个体时即在约束范围内随机生成用量
2. **变异操作后**：立即进行边界检查和裁剪
3. **归一化阶段**：在调整未约束食材用量时保持约束食材用量不变
4. **最终校验**：算法结束前进行最终边界检查

### 3. 营养素目标约束校验逻辑
**预校验流程**:
```
开始
  ↓
检查是否存在食材约束冲突
（如 min > max）
  ↓
计算最小可能总重量
（所有食材取最小值之和）
  ↓
计算最大可能总重量
（所有食材取最大值之和）
  ↓
验证目标总重量是否在可行范围内
  ├─ 是 → 继续执行优化
  └─ 否 → 发出警告信息（但继续执行）
结束
```

**设计原则**:
- 容错性：约束冲突时继续执行，不直接返回错误
- 用户体验：通过 `warnings` 字段反馈潜在问题
- 自愈能力：算法内部尝试寻找可行解

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

# 仅运行Bug修复相关测试 ✅
go test -v -run FixedOptimizer
```

### 启动服务

```bash
# 启动完整服务（含修复版优化器）
go run main.go models.go buggy_optimizer.go moead_optimizer.go fixed_optimizer.go

# 在浏览器中打开Swagger UI
open http://localhost:8080/swagger/index.html
```

## 许可证

本项目采用MIT许可证。详见LICENSE文件。