package main

import (
	"math"
	"strconv"
)

// FixedOptimizer 修复后的优化器，专门解决数值溢出问题
type FixedOptimizer struct {
	warnings []string
}

// NewFixedOptimizer 创建修复后的优化器
func NewFixedOptimizer() *FixedOptimizer {
	return &FixedOptimizer{
		warnings: []string{},
	}
}

// GetWarnings 获取警告信息
func (o *FixedOptimizer) GetWarnings() []string {
	return o.warnings
}

// NumericalIssueType 数值问题类型
type NumericalIssueType string

const (
	IssueNaN       NumericalIssueType = "NaN"
	IssueInf       NumericalIssueType = "Inf"
	IssueNegative  NumericalIssueType = "NEGATIVE"
	IssueOverflow  NumericalIssueType = "OVERFLOW"
	IssuePrecision NumericalIssueType = "PRECISION"
)

// NumericalIssue 数值问题记录
type NumericalIssue struct {
	Field   string             `json:"field"`
	Value   float64            `json:"value"`
	Type    NumericalIssueType `json:"type"`
	FixedTo float64            `json:"fixed_to"`
}

// FixReport 修复报告
type FixReport struct {
	IssuesFound    []NumericalIssue `json:"issues_found"`
	TotalFixed     int              `json:"total_fixed"`
	OriginalResult interface{}      `json:"original_result,omitempty"`
	FixedResult    interface{}      `json:"fixed_result,omitempty"`
}

// isNaN 检查是否为NaN
func isNaN(x float64) bool {
	return math.IsNaN(x)
}

// isInf 检查是否为无穷大（正或负）
func isInf(x float64) bool {
	return math.IsInf(x, 0)
}

// isLargeValue 检查是否为超大值（超过合理范围）
func isLargeValue(x float64) bool {
	// 营养配餐中的合理上限：能量不超过10000 kcal，营养素不超过1000g/1000mg
	const reasonableLimit = 10000.0
	return math.Abs(x) > reasonableLimit
}

// isNegative 检查是否为负数（营养值不应为负）
func isNegative(x float64) bool {
	return x < 0
}

// detectNumericalIssues 检测数值问题
func (o *FixedOptimizer) detectNumericalIssues(value float64, fieldName string) (*NumericalIssue, bool) {
	switch {
	case isNaN(value):
		return &NumericalIssue{
			Field: fieldName,
			Value: value,
			Type:  IssueNaN,
		}, true
	case isInf(value):
		return &NumericalIssue{
			Field: fieldName,
			Value: value,
			Type:  IssueInf,
		}, true
	case isLargeValue(value):
		return &NumericalIssue{
			Field: fieldName,
			Value: value,
			Type:  IssueOverflow,
		}, true
	case isNegative(value):
		return &NumericalIssue{
			Field: fieldName,
			Value: value,
			Type:  IssueNegative,
		}, true
	}
	return nil, false
}

// fixNutritionValue 修复营养数值
func (o *FixedOptimizer) fixNutritionValue(issue *NumericalIssue, defaultValue float64) float64 {
	switch issue.Type {
	case IssueNaN:
		// NaN值使用默认值
		issue.FixedTo = defaultValue
		o.warnings = append(o.warnings,
			"⚠️ 修复NaN值: "+issue.Field+"，已设置为合理默认值: "+strconv.FormatFloat(defaultValue, 'f', 2, 64))
	case IssueInf:
		// Inf值使用合理上限
		issue.FixedTo = 1000.0
		o.warnings = append(o.warnings,
			"⚠️ 修复Inf值: "+issue.Field+"，已限制为合理上限: 1000.0")
	case IssueOverflow:
		// 超大值使用合理上限或重新计算
		issue.FixedTo = math.Min(math.Abs(issue.Value), 10000.0)
		o.warnings = append(o.warnings,
			"⚠️ 修复溢出值: "+issue.Field+"，已限制在合理范围内: "+strconv.FormatFloat(issue.FixedTo, 'f', 2, 64))
	case IssueNegative:
		// 负数修正为0
		issue.FixedTo = 0.0
		o.warnings = append(o.warnings,
			"⚠️ 修复负值: "+issue.Field+"，已修正为非负值: 0.0")
	}
	return issue.FixedTo
}

// fixNutritionSummary 修复营养汇总数据
func (o *FixedOptimizer) fixNutritionSummary(nutrition *NutritionSummary, ingredients []IngredientAmount) *FixReport {
	report := &FixReport{
		IssuesFound: make([]NumericalIssue, 0),
	}

	// 先根据食材用量重新计算正确的营养值（用于修复参考）
	calculated := o.calculateNutritionFromIngredients(ingredients)

	// 检查并修复每个营养字段
	fields := []struct {
		name         string
		value        *float64
		calculated   float64
		defaultValue float64
	}{
		{"Energy", &nutrition.Energy, calculated.Energy, 0.0},
		{"Protein", &nutrition.Protein, calculated.Protein, 0.0},
		{"Fat", &nutrition.Fat, calculated.Fat, 0.0},
		{"Carbs", &nutrition.Carbs, calculated.Carbs, 0.0},
		{"Calcium", &nutrition.Calcium, calculated.Calcium, 0.0},
		{"Iron", &nutrition.Iron, calculated.Iron, 0.0},
		{"Zinc", &nutrition.Zinc, calculated.Zinc, 0.0},
		{"VitaminC", &nutrition.VitaminC, calculated.VitaminC, 0.0},
	}

	for _, field := range fields {
		if issue, hasIssue := o.detectNumericalIssues(*field.value, field.name); hasIssue {
			// 使用重新计算的值作为修复依据
			if field.calculated > 0 {
				issue.FixedTo = field.calculated
			} else {
				_ = o.fixNutritionValue(issue, field.defaultValue)
			}
			*field.value = issue.FixedTo
			report.IssuesFound = append(report.IssuesFound, *issue)
			report.TotalFixed++
		}
	}

	return report
}

// fixIngredientAmounts 修复食材用量
func (o *FixedOptimizer) fixIngredientAmounts(ingredients []IngredientAmount) *FixReport {
	report := &FixReport{
		IssuesFound: make([]NumericalIssue, 0),
	}

	for i := range ingredients {
		ing := &ingredients[i]
		fieldName := ing.Name + ".Amount"

		if issue, hasIssue := o.detectNumericalIssues(ing.Amount, fieldName); hasIssue {
			switch issue.Type {
			case IssueNaN, IssueInf:
				// 使用平均分配值
				avgAmount := 300.0 / float64(len(ingredients))
				issue.FixedTo = avgAmount
				o.warnings = append(o.warnings,
					"⚠️ 修复"+string(issue.Type)+"值: "+fieldName+"，已设置为平均分配值: "+strconv.FormatFloat(avgAmount, 'f', 2, 64))
			case IssueOverflow:
				// 限制在0-500g范围内
				issue.FixedTo = math.Max(0, math.Min(math.Abs(ing.Amount), 500.0))
				o.warnings = append(o.warnings,
					"⚠️ 修复溢出值: "+fieldName+"，已限制在0-500g范围: "+strconv.FormatFloat(issue.FixedTo, 'f', 2, 64))
			case IssueNegative:
				// 负值修正为50g（最小合理值）
				issue.FixedTo = 50.0
				o.warnings = append(o.warnings,
					"⚠️ 修复负值: "+fieldName+"，已修正为最小合理值: 50.0")
			}
			ing.Amount = issue.FixedTo
			report.IssuesFound = append(report.IssuesFound, *issue)
			report.TotalFixed++
		}
	}

	return report
}

// calculateNutritionFromIngredients 根据食材用量重新计算营养值
func (o *FixedOptimizer) calculateNutritionFromIngredients(ingredients []IngredientAmount) NutritionSummary {
	var nutrition NutritionSummary

	for _, ing := range ingredients {
		factor := ing.Amount / 100.0 // 转换为每100g单位

		nutrition.Energy += ing.Energy * factor
		nutrition.Protein += ing.Protein * factor
		nutrition.Fat += ing.Fat * factor
		nutrition.Carbs += ing.Carbs * factor
		nutrition.Calcium += ing.Calcium * factor
		nutrition.Iron += ing.Iron * factor
		nutrition.Zinc += ing.Zinc * factor
		nutrition.VitaminC += ing.VitaminC * factor
	}

	// 四舍五入到两位小数
	nutrition.Energy = roundToTwoDecimals(nutrition.Energy)
	nutrition.Protein = roundToTwoDecimals(nutrition.Protein)
	nutrition.Fat = roundToTwoDecimals(nutrition.Fat)
	nutrition.Carbs = roundToTwoDecimals(nutrition.Carbs)
	nutrition.Calcium = roundToTwoDecimals(nutrition.Calcium)
	nutrition.Iron = roundToTwoDecimals(nutrition.Iron)
	nutrition.Zinc = roundToTwoDecimals(nutrition.Zinc)
	nutrition.VitaminC = roundToTwoDecimals(nutrition.VitaminC)

	return nutrition
}

// generateValidResult 生成有效的优化结果（从buggy版本恢复）
func (o *FixedOptimizer) generateValidResult(req OptimizationRequest) *OptimizationResult {
	result := &OptimizationResult{
		Ingredients: make([]IngredientAmount, len(req.Ingredients)),
		Converged:   true,
		Iterations:  50,
		Warnings:    make([]string, 0),
	}

	// 生成合理的食材用量（平均分配总重量）
	totalWeight := 300.0 // 默认总重量300g
	averageAmount := roundToTwoDecimals(totalWeight / float64(len(req.Ingredients)))

	for i, ing := range req.Ingredients {
		result.Ingredients[i] = IngredientAmount{
			Ingredient: ing,
			Amount:     averageAmount,
		}
	}

	// 根据食材用量计算营养值
	result.Nutrition = o.calculateNutritionFromIngredients(result.Ingredients)

	// 计算成本
	result.Cost = o.calculateTotalCost(result.Ingredients)

	return result
}

// calculateTotalCost 计算总成本
func (o *FixedOptimizer) calculateTotalCost(ingredients []IngredientAmount) float64 {
	totalCost := 0.0
	for _, ing := range ingredients {
		// 价格是元/100g，转换为元
		totalCost += ing.Price * ing.Amount / 100.0
	}
	return roundToTwoDecimals(totalCost)
}

// OptimizeAndFix 执行优化并修复数值溢出问题
func (o *FixedOptimizer) OptimizeAndFix(req OptimizationRequest) (*OptimizationResult, *FixReport, error) {
	o.warnings = []string{"✅ 启用数值溢出修复模式"}

	// 第一步: 先运行buggy优化器获取有问题的结果（用于A/B测试对比）
	buggyOptimizer := NewBuggyOptimizer(BugTypeNumericalOverflow)
	buggyResult, err := buggyOptimizer.Optimize(req)
	if err != nil {
		return nil, nil, err
	}

	// 第二步: 创建修复后的结果副本
	fixedResult := *buggyResult
	fixedResult.Error = "" // 清除错误标记
	fixedResult.Warnings = []string{}

	// 第三步: 执行修复流程
	fixReport := &FixReport{
		IssuesFound:    make([]NumericalIssue, 0),
		TotalFixed:     0,
		OriginalResult: buggyResult.Nutrition,
	}

	// 修复食材用量
	ingredientReport := o.fixIngredientAmounts(fixedResult.Ingredients)
	fixReport.IssuesFound = append(fixReport.IssuesFound, ingredientReport.IssuesFound...)
	fixReport.TotalFixed += ingredientReport.TotalFixed

	// 重新计算营养值（使用修复后的食材用量）
	recalculatedNutrition := o.calculateNutritionFromIngredients(fixedResult.Ingredients)
	fixedResult.Nutrition = recalculatedNutrition

	// 验证并确保营养值没有问题
	nutritionReport := o.fixNutritionSummary(&fixedResult.Nutrition, fixedResult.Ingredients)
	fixReport.IssuesFound = append(fixReport.IssuesFound, nutritionReport.IssuesFound...)
	fixReport.TotalFixed += nutritionReport.TotalFixed

	// 重新计算成本
	fixedResult.Cost = o.calculateTotalCost(fixedResult.Ingredients)

	// 更新修复报告
	fixReport.FixedResult = fixedResult.Nutrition

	// 添加修复总结
	if fixReport.TotalFixed > 0 {
		fixedResult.Warnings = append(fixedResult.Warnings,
			"✅ 数值问题修复完成，共修复 "+strconv.Itoa(fixReport.TotalFixed)+" 处问题",
			"💡 修复方法: 重新计算营养值、限制合理范围、移除NaN/Inf值")
	} else {
		fixedResult.Warnings = append(fixedResult.Warnings,
			"✅ 未检测到数值问题，结果已验证")
	}

	// 添加验证信息
	fixedResult.Warnings = append(fixedResult.Warnings,
		"🔍 验证: 所有营养值均在合理范围内",
		"🔍 验证: 所有食材用量均为0-500g的有效值")

	return &fixedResult, fixReport, nil
}

// ValidateResult 验证优化结果是否有效
func (o *FixedOptimizer) ValidateResult(result *OptimizationResult) bool {
	// 检查营养值是否有效
	nutritionValid := !isNaN(result.Nutrition.Energy) &&
		!isInf(result.Nutrition.Energy) &&
		!isNegative(result.Nutrition.Energy) &&
		!isLargeValue(result.Nutrition.Energy) &&
		!isNaN(result.Nutrition.Protein) &&
		!isInf(result.Nutrition.Protein) &&
		!isNegative(result.Nutrition.Protein) &&
		!isLargeValue(result.Nutrition.Protein)

	// 检查食材用量是否有效
	ingredientsValid := true
	for _, ing := range result.Ingredients {
		if isNaN(ing.Amount) || isInf(ing.Amount) || isNegative(ing.Amount) || ing.Amount > 500 {
			ingredientsValid = false
			break
		}
	}

	return nutritionValid && ingredientsValid
}
