# 营养配餐多目标优化算法Bug测试套件

## 项目概述

本项目提供了一个完整的营养配餐多目标优化算法测试框架，专门用于复现和测试5类典型科学计算Bug。测试套件包含含Bug的完整Go+Gin可运行代码、测试用例和Bug复现机制，并集成了Swagger UI提供可视化API测试界面。

## 核心特性

### 🐛 5类典型科学计算Bug

1. **浮点数精度丢失** - 多次迭代累加导致营养素计算偏差
2. **数值溢出** - 返回NaN/Inf结果
3. **约束越界** - 食材重量出现负数或超过500g的超大值
4. **收敛失败** - 求解器无法收敛，返回空方案或极端不合理用量
5. **结果不稳定** - 同一参数多次运行结果不一致

### 🔧 技术栈

- **后端框架**: Go + Gin
- **API文档**: Swagger UI (OpenAPI 3.0)
- **优化算法**: 
  - 加权求和多目标优化（基准版本）
  - **MOEA/D多目标进化算法**（新增，基于分解的多目标优化算法）
- **数据存储**: 
  - JSON文件（无需数据库）
  - MySQL数据库支持
  - 数据库导出的JSON食材数据（200种食材）

## 项目结构

```
nutrient-optimizer-benchmark/
├── main.go                  # 主程序入口，Gin路由设置
├── models.go               # 数据模型定义
├── buggy_optimizer.go      # 含Bug的优化器实现（基准版本）
├── moead_optimizer.go      # ✅ MOEA/D多目标优化算法实现
├── moead_optimizer_test.go # ✅ MOEA/D算法单元测试
├── ingredients.json       # 15种常用食材数据（基准）
├── ingredients_db_export.json # ✅ 数据库导出的200种食材数据
├── test_cases.json        # 测试用例配置
├── docs/                  # Swagger文档生成
│   └── docs.go            # Swagger配置
├── go.mod                  # Go模块依赖
└── README.md              # 项目文档
```

## 快速开始

### 1. 环境准备

```bash
# 安装Go依赖
go mod tidy

# 安装Swagger CLI工具 (可选，用于文档生成)
go install github.com/swaggo/swag/cmd/swag@latest

# 无需数据库配置，所有数据存储在JSON文件中
```

### 2. 启动服务

#### 基准版本（仅含原优化器）
```bash
# 启动营养配餐优化服务（仅原优化器）
go run main.go models.go buggy_optimizer.go

# 服务将在 http://localhost:8080 启动
# Swagger文档地址: http://localhost:8080/swagger/index.html
```

#### ✅ 完整版本（含MOEA/D优化器，推荐）
```bash
# 启动含MOEA/D算法的完整服务（推荐，无需数据库）
go run main.go models.go buggy_optimizer.go moead_optimizer.go

# 服务将在 http://localhost:8080 启动
# Swagger文档地址: http://localhost:8080/swagger/index.html
# 包含MOEA/D优化器接口
```

> 💡 **注意**: MOEA/D服务支持**无数据库运行**，自动使用导出的JSON食材数据。数据库连接失败时会自动降级到JSON文件。

### 3. 使用Swagger UI进行测试

打开浏览器访问: `http://localhost:8080/swagger/index.html`

## Swagger UI使用指南

### 📖 API文档界面
Swagger UI提供了完整的API文档和交互式测试界面：

1. **API概览** - 显示所有可用的API端点
2. **参数说明** - 详细的请求参数说明和示例
3. **在线测试** - 直接在浏览器中发送请求并查看响应
4. **模型定义** - 完整的数据结构定义

### 🔍 主要API端点

#### 1. 健康检查
- **路径**: `GET /api/health`
- **功能**: 检查服务状态
- **测试**: 直接在Swagger UI中点击"Try it out" → "Execute"

#### 2. 正常优化（基准版本）
- **路径**: `POST /api/optimize`
- **功能**: 使用基准版本的优化器
- **参数**: 完整的优化请求参数
- **测试**: 使用test_cases.json中的示例数据

#### 3. 含Bug优化（测试版本）
- **路径**: `POST /api/optimize-with-bugs`
- **功能**: 测试特定类型的科学计算Bug
- **参数**: 
  - `bug_type`: Bug类型（precision_loss, numerical_overflow等）
  - 优化请求参数

#### ✅ 4. MOEA/D多目标优化（新增）
- **路径**: `POST /api/optimize-moead`
- **功能**: 使用MOEA/D多目标进化算法进行配餐优化
- **算法**: 基于分解的多目标进化算法
- **优势**: 更好的多目标优化能力，Pareto最优解集
- **参数**:
  - `population_size`: 种群大小（可选，默认: 50）
  - `max_iterations`: 最大迭代次数（可选，默认: 100）
  - 优化请求参数
  - **支持无数据库运行**（自动加载JSON食材数据）

#### ✅ 5. 获取食材列表（新增）
- **路径**: `GET /api/ingredients`
- **功能**: 获取可用的食材列表
- **参数**:
  - `source`: 数据源（可选，db或json，默认: db）
  - `limit`: 返回数量限制（可选，默认: 20）
  - `filepath`: JSON文件路径（可选）
- **自动降级**: 数据库失败时自动使用JSON文件

### 🧪 在Swagger UI中测试Bug

#### 测试浮点数精度丢失
1. 选择 `/api/optimize-with-bugs` 端点
2. 设置 `bug_type` 为 `precision_loss`
3. 使用test_cases.json中的请求体
4. 执行请求，观察营养素计算偏差

#### 测试数值溢出
1. 设置 `bug_type` 为 `numerical_overflow`
2. 执行请求，观察返回结果中的NaN/Inf值

#### 测试约束越界
1. 设置 `bug_type` 为 `constraint_violation`
2. 执行请求，观察食材用量出现负数或超大值

## API接口详情

### 健康检查
```http
GET /api/health
```

### 正常优化（基准版本）
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

{
  // 同上
}
```

支持的bug_type参数：
- `precision_loss` - 浮点数精度丢失
- `numerical_overflow` - 数值溢出
- `constraint_violation` - 约束越界
- `convergence_failure` - 收敛失败
- `result_instability` - 结果不稳定

### ✅ MOEA/D多目标优化
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

> 💡 **提示**: 如不提供 `ingredients` 字段，系统将**自动从数据库或JSON文件加载食材数据（数据库连接失败时自动降级到JSON文件）。

### ✅ 获取食材列表
```http
# 从数据库获取（默认）
GET /api/ingredients?limit=20

# 从JSON文件获取（无数据库）
GET /api/ingredients?source=json&limit=20
```

## 测试用例

### 常规场景测试
- **描述**: 正常营养配餐优化场景
- **食材**: 鸡胸肉、西兰花、米饭、鸡蛋
- **营养目标**: 能量600kcal，蛋白质30g等
- **预期Bug**: 无

### 边界场景测试（低热量目标）
- **描述**: 极低热量目标的边界情况
- **食材**: 黄瓜、西红柿、生菜等低热量食材
- **营养目标**: 能量100kcal，蛋白质5g
- **预期Bug**: 精度丢失、收敛失败

### 极端数值测试
- **描述**: 测试极端数值情况下的稳定性
- **食材**: 高能量、高蛋白、高微量营养素食材
- **营养目标**: 能量2000kcal，蛋白质100g等
- **预期Bug**: 数值溢出、约束越界、结果不稳定

## Bug复现指南

### 1. 数值精度问题
- **现象**: 多次迭代累加时营养素计算出现偏差
- **复现方法**: 使用float32进行浮点运算，模拟多次迭代累加

### 2. 数值溢出问题
- **现象**: 返回结果中出现NaN/Inf值
- **复现方法**: 故意制造除零操作和对数运算异常

### 3. 约束越界问题
- **现象**: 食材重量出现负数或超过500g的超大值
- **复现方法**: 设置矛盾的约束条件，缺失边界约束

### 4. 收敛失败问题
- **现象**: 求解器无法收敛，返回空方案或极端不合理用量
- **复现方法**: 设置极严格收敛阈值，构造不合理初始解

### 5. 结果不稳定问题
- **现象**: 同一参数多次运行结果不一致
- **复现方法**: 添加随机扰动，使用非确定性算法

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
包含3个测试场景：
- **常规场景**: 正常营养配餐优化
- **边界场景**: 极低热量目标
- **极端数值**: 测试数值稳定性

## ✅ MOEA/D算法使用示例

### 示例1: 无数据库快速测试（推荐）
```bash
# 启动服务（自动使用JSON食材数据）
go run main.go models.go buggy_optimizer.go moead_optimizer.go
```

无需配置MySQL，直接使用导出的JSON文件中的200种食材。

### 示例2: 使用MOEA/D API进行优化（请求体中不包含ingredients时自动加载JSON中的200种真实食材数据）
```javascript
// POST /api/optimize-moead
{
  "population_size": 30,
  "max_iterations": 100,
  "nutrition_goals": [
    {"nutrient": "energy", "target": 600, "min": 500, "max": 700, "weight": 0.3},
    {"nutrient": "protein", "target": 30, "min": 20, "max": 40, "weight": 0.4},
    {"nutrient": "fat", "target": 15, "min": 10, "max": 25, "weight": 0.2},
    {"nutrient": "carbs", "target": 80, "min": 60, "max": 100, "weight": 0.1}
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

### 示例3: 获取JSON文件中的食材列表
```http
GET /api/ingredients?source=json&limit=10
```

## 测试任务要求

### 根因定位 → 代码改进 → 回归测试 → 验证报告

请完成以下测试任务：
1. **根因定位**: 准确识别每类Bug的根本原因
2. **代码改进**: 改进所有相关代码路径，不引入新问题
3. **回归测试**: 确保改进后所有功能正常
4. **验证报告**: 提供完整的验证结果

## 扩展使用

### ✅ 运行MOEA/D单元测试
```bash
# 运行所有单元测试（含MOEA/D算法测试）
go test -v ./...

# 仅运行MOEA/D优化器测试
go test -v -run MOEAD
```

### 自定义测试用例
编辑 `test_cases.json` 文件添加新的测试场景：

```json
{
  "name": "自定义测试",
  "description": "自定义测试场景描述",
  "request": {
    // 自定义请求参数
  }
}
```

### 集成测试

**Windows PowerShell:**
```powershell
# 启动完整服务（含MOEA/D优化器，推荐）
Start-Process powershell -ArgumentList "go run main.go models.go buggy_optimizer.go moead_optimizer.go"

# 或启动基准版本（仅原优化器）
# Start-Process powershell -ArgumentList "go run main.go models.go buggy_optimizer.go"

# 等待服务启动
Start-Sleep -Seconds 5

# 在浏览器中打开Swagger UI
Start-Process "http://localhost:8080/swagger/index.html"
```

**Linux/macOS:**
```bash
# 启动完整服务（含MOEA/D优化器，推荐）
go run main.go models.go buggy_optimizer.go moead_optimizer.go &
sleep 5

# 在浏览器中打开Swagger UI
open http://localhost:8080/swagger/index.html
```

## 故障排除

### 常见问题

1. **Swagger UI无法访问**
   - 检查服务是否正常启动
   - 确认端口8080未被占用
   - 访问 `http://localhost:8080/swagger/index.html`

2. **JSON文件读取失败**
   - 确认ingredients.json和test_cases.json文件存在
   - 检查文件权限是否正确
   - 验证JSON格式是否有效

3. **MOEA/D服务问题**
   - 确认 ingredients_db_export.json 文件存在（200种食材数据）
   - 数据库连接失败时会自动降级到JSON文件
   - 可通过 ?source=json 参数强制使用JSON数据源

4. **API请求失败**
   - 检查请求体JSON格式
   - 确认参数类型和范围正确
   - 查看服务日志获取详细错误信息

### 调试技巧

- 使用Swagger UI的"Try it out"功能进行交互式测试
- 查看浏览器开发者工具的网络请求详情
- 使用服务日志了解后端处理过程
- 对比含Bug版本和基准版本的输出差异

## 许可证

本项目采用MIT许可证。详见LICENSE文件。

---

**注意**: 本测试套件专为Bug复现和测试设计，包含故意引入的Bug用于测试目的。正确的实现需要由测试者自行完成。

**Swagger UI优势**: 提供可视化测试界面，便于多个模型对比测试时的统一标准和交互式调试。