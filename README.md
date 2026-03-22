# 营养配餐多目标优化算法Bug测试套件

## 项目概述

本项目提供营养配餐多目标优化算法测试框架，包含5类典型科学计算Bug的复现机制及修复方案。

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
  - FixedOptimizer（数值溢出修复版）
- **数据存储**: JSON文件 / MySQL数据库

## 项目结构

```
nutrient-optimizer-benchmark/
├── main.go                  # 主程序入口
├── models.go               # 数据模型
├── buggy_optimizer.go      # 含Bug的优化器
├── fixed_optimizer.go      # 修复后的优化器
├── moead_optimizer.go      # MOEA/D优化器
├── moead_optimizer_test.go # MOEA/D单元测试
├── fixed_optimizer_test.go # 修复优化器单元测试
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
# 完整服务（含所有优化器）
go run .

# Swagger文档: http://localhost:8080/swagger/index.html
```

### 3. 使用Swagger UI

打开浏览器访问: `http://localhost:8080/swagger/index.html`

## API端点

#### 1. 健康检查
- **路径**: `GET /api/health`

#### 2. 含Bug优化
- **路径**: `POST /api/optimize-with-bugs`
- **参数**: 
  - `bug_type`: Bug类型（precision_loss, numerical_overflow等）

#### 3. 修复后优化（新增）
- **路径**: `POST /api/optimize-with-bugs-fixed`
- **参数**: 
  - `bug_type`: Bug类型（可选，用于A/B对比测试）

#### 4. MOEA/D多目标优化
- **路径**: `POST /api/optimize-moead`
- **参数**:
  - `population_size`: 种群大小（可选，默认: 50）
  - `max_iterations`: 最大迭代次数（可选，默认: 100）

#### 5. 获取食材列表
- **路径**: `GET /api/ingredients`
- **参数**:
  - `source`: 数据源（可选，db或json，默认: db）
  - `limit`: 返回数量限制（可选，默认: 20）

## 数值计算逻辑说明

### 营养素计算公式

每种营养素的计算采用标准公式：

```
营养素总量 = Σ(食材用量(g) / 100 × 食材营养素含量(每100g))
```

示例：
```
能量(kcal) = Σ(amount_i / 100 × energy_i)
蛋白质(g) = Σ(amount_i / 100 × protein_i)
脂肪(g) = Σ(amount_i / 100 × fat_i)
碳水化合物(g) = Σ(amount_i / 100 × carbs_i)
```

### FixedOptimizer 安全计算机制

为防止数值溢出和异常值，实现了以下安全计算函数：

#### 1. 安全乘法 (safeMultiply)
```go
func safeMultiply(a, b float64) float64 {
    if math.IsNaN(a) || math.IsNaN(b) { return 0 }
    if math.IsInf(a, 0) || math.IsInf(b, 0) { return MaxNutritionValue }
    result := a * b
    if math.IsNaN(result) || math.IsInf(result, 0) { return MaxNutritionValue }
    return result
}
```

#### 2. 安全除法 (safeDivide)
```go
func safeDivide(a, b float64) float64 {
    if math.Abs(b) < Epsilon { return 0 }  // 除零保护
    if math.IsNaN(a) || math.IsNaN(b) { return 0 }
    result := a / b
    if math.IsNaN(result) || math.IsInf(result, 0) { return MaxNutritionValue }
    return result
}
```

#### 3. 安全对数 (safeLog)
```go
func safeLog(value float64) float64 {
    if value <= 0 { return 0 }  // 非法对数操作保护
    return math.Log(value)
}
```

#### 4. 安全指数 (safeExp)
```go
func safeExp(value float64) float64 {
    if value > 700 { return MaxNutritionValue }  // 指数溢出保护
    if value < -700 { return 0 }
    return math.Exp(value)
}
```

## 重量约束规则

### 食材用量约束
| 约束类型 | 最小值 | 最大值 | 说明 |
|---------|--------|--------|------|
| 单个食材用量 | 0g | 500g | 单个食材用量上限 |
| 总重量约束 | - | 用户指定 | 默认300g |

### 营养素值约束
| 营养素 | 最小值 | 最大值 | 说明 |
|--------|--------|--------|------|
| 能量 | 0 | 1,000,000 | kcal |
| 蛋白质 | 0 | 1,000,000 | g |
| 脂肪 | 0 | 1,000,000 | g |
| 碳水化合物 | 0 | 1,000,000 | g |
| 钙 | 0 | 1,000,000 | mg |
| 铁 | 0 | 1,000,000 | mg |
| 锌 | 0 | 1,000,000 | mg |
| 维生素C | 0 | 1,000,000 | mg |

### 约束处理流程

1. **食材约束检查**: 确保每个食材用量在 [0, 500]g 范围内
2. **总重量归一化**: 调整食材用量以满足总重量约束
3. **营养素值钳制**: 确保所有营养素值在有效范围内
4. **异常值检测**: 自动检测并修正 NaN/Inf/负数/超大值

## 算法基准验证结果

### A/B测试对比数据

使用相同输入参数测试 `numerical_overflow` Bug：

#### 输入参数
```json
{
  "ingredients": [
    {"id": 2, "name": "小麦", "energy": 338, "protein": 11.9, "fat": 1.3, "carbs": 75.2},
    {"id": 3, "name": "五谷香", "energy": 378, "protein": 9.9, "fat": 2.6, "carbs": 78.9}
  ]
}
```

#### 对比结果

| 指标 | Buggy版本 | Fixed版本 | 状态 |
|------|-----------|-----------|------|
| Energy | 1.797693e+308 | 1074.00 | ✅ 已修复 |
| Protein | -1.797693e+308 | 32.70 | ✅ 已修复 |
| Fat | 1.797693e+308 | 5.85 | ✅ 已修复 |
| Carbs | 231.15 | 231.15 | ✅ 正常 |
| has_nan | false | false | ✅ 正常 |
| has_inf | false | false | ✅ 正常 |
| has_negative | true | false | ✅ 已修复 |
| all_valid | false | true | ✅ 已修复 |

### 单元测试覆盖率

```
=== FixedOptimizer Tests ===
✅ TestFixedOptimizer_ClampAmount         - 食材用量边界约束
✅ TestFixedOptimizer_ClampNutrition      - 营养素值边界约束
✅ TestFixedOptimizer_SafeMultiply        - 安全乘法测试
✅ TestFixedOptimizer_SafeDivide          - 安全除法测试
✅ TestFixedOptimizer_SafeLog             - 安全对数测试
✅ TestFixedOptimizer_SafeExp             - 安全指数测试
✅ TestFixedOptimizer_Optimize            - 完整优化流程
✅ TestFixedOptimizer_ExtremeValues       - 极端值处理
✅ TestNumericalOverflowComparison        - A/B对比测试

=== MOEA/D Tests ===
✅ TestMOEADOptimizer_CreateIndividual    - 个体创建
✅ TestMOEADOptimizer_CalculateNutrition  - 营养计算
✅ TestMOEADOptimizer_CalculateTotalCost  - 成本计算
✅ TestMOEADOptimizer_CalculateDiversity  - 多样性计算
✅ TestMOEADOptimizer_Optimize            - 完整优化流程
```

### 极端值测试场景

| 场景 | 输入值 | 期望结果 | 实际结果 |
|------|--------|----------|----------|
| 超大能量值 | 1e10 | 钳制到1e6 | ✅ 通过 |
| NaN输入 | NaN | 修正为0 | ✅ 通过 |
| Inf输入 | +Inf | 钳制到1e6 | ✅ 通过 |
| 负值输入 | -100 | 修正为0 | ✅ 通过 |
| 除零操作 | x/0 | 返回0 | ✅ 通过 |
| 非法对数 | log(-1) | 返回0 | ✅ 通过 |
| 指数溢出 | exp(800) | 钳制到1e6 | ✅ 通过 |

## API接口详情

### 含Bug优化
```http
POST /api/optimize-with-bugs?bug_type=numerical_overflow
Content-Type: application/json
```

支持的bug_type：
- `precision_loss` - 浮点数精度丢失
- `numerical_overflow` - 数值溢出
- `constraint_violation` - 约束越界
- `convergence_failure` - 收敛失败
- `result_instability` - 结果不稳定

### 修复后优化（A/B测试）
```http
POST /api/optimize-with-bugs-fixed?bug_type=numerical_overflow
Content-Type: application/json
```

返回数据包含：
- `result`: 修复后的优化结果
- `ab_test`: A/B测试对比数据
- `fix_applied`: 应用的修复措施列表
- `fix_summary`: 修复摘要

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

## Bug现象说明

### 1. 数值精度问题
- **现象**: 营养素计算出现偏差
- **修复**: 使用float64替代float32，添加精度保护

### 2. 数值溢出问题
- **现象**: 返回结果中出现NaN/Inf值
- **修复**: 实现安全计算函数，自动检测并修正异常值

### 3. 约束越界问题
- **现象**: 食材重量出现负数或超过500g的超大值
- **修复**: 添加边界约束检查和自动修正

### 4. 收敛失败问题
- **现象**: 求解器无法收敛，返回空方案或极端不合理用量
- **修复**: 优化初始解构造，放宽收敛阈值

### 5. 结果不稳定问题
- **现象**: 同一参数多次运行结果不一致
- **修复**: 固定随机种子，确保结果可复现

## 数据存储

项目使用JSON文件存储数据，无需数据库配置：

### ingredients.json
包含15种常用食材的营养成分数据。

### ingredients_db_export.json
从数据库导出的200种真实食材数据。

## 扩展使用

### 运行单元测试
```bash
# 运行所有单元测试
go test -v ./...

# 仅运行FixedOptimizer测试
go test -v -run Fixed

# 仅运行MOEA/D优化器测试
go test -v -run MOEAD
```

### 生成Swagger文档
```bash
swag init
```

## 许可证

本项目采用MIT许可证。详见LICENSE文件。
