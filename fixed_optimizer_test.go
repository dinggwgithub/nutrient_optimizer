package main

import (
	"math"
	"testing"
)

func TestFixedOptimizer_ClampAmount(t *testing.T) {
	optimizer := NewFixedOptimizer()

	tests := []struct {
		input    float64
		expected float64
	}{
		{-50, 0},
		{0, 0},
		{250, 250},
		{500, 500},
		{600, 500},
		{1000, 500},
	}

	for _, test := range tests {
		result := optimizer.clampAmount(test.input)
		if result != test.expected {
			t.Errorf("clampAmount(%.2f) = %.2f, expected %.2f", test.input, result, test.expected)
		}
	}
}

func TestFixedOptimizer_CreateSmartIndividual(t *testing.T) {
	optimizer := NewFixedOptimizer()

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "小麦", Energy: 338, Protein: 11.9},
			{ID: 2, Name: "五谷香", Energy: 378, Protein: 9.9},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
	}

	ind := optimizer.createSmartIndividual(2, req)

	if ind == nil {
		t.Fatal("个体创建失败")
	}

	if len(ind.Amounts) != 2 {
		t.Errorf("食材用量数组长度错误: 期望2，实际%d", len(ind.Amounts))
	}

	for i, amount := range ind.Amounts {
		if amount < MinIngredientAmount || amount > MaxIngredientAmount {
			t.Errorf("食材%d用量超出范围[0,500]: %.2fg", i, amount)
		}
	}

	total := 0.0
	for _, amount := range ind.Amounts {
		total += amount
	}
	t.Logf("总重量: %.2fg", total)
}

func TestFixedOptimizer_ExtremeAmount1000g(t *testing.T) {
	optimizer := NewFixedOptimizer()

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 2, Name: "小麦", Energy: 338, Protein: 11.9, Fat: 1.3, Carbs: 75.2, Calcium: 34, Iron: 5.1, Zinc: 2.33, VitaminC: 0, Price: 0},
			{ID: 3, Name: "五谷香", Energy: 378, Protein: 9.9, Fat: 2.6, Carbs: 78.9, Calcium: 2, Iron: 0.5, Zinc: 0.23, VitaminC: 0, Price: 0},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
		MaxIterations: 50,
		Tolerance:     0.05,
	}

	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	if !result.Converged {
		t.Error("优化未收敛")
	}

	if result.Error != "" {
		t.Errorf("优化返回错误: %s", result.Error)
	}

	for i, ing := range result.Ingredients {
		if ing.Amount < MinIngredientAmount || ing.Amount > MaxIngredientAmount {
			t.Errorf("食材%d(%s)用量超出范围[0,500]: %.2fg", i, ing.Name, ing.Amount)
		}
		if ing.Amount == 1000 {
			t.Errorf("食材%d(%s)用量仍为极端值1000g，Bug未修复", i, ing.Name)
		}
	}

	t.Logf("优化结果:")
	for _, ing := range result.Ingredients {
		t.Logf("  %s: %.2fg", ing.Name, ing.Amount)
	}
	t.Logf("营养: 能量=%.2f kcal, 蛋白质=%.2f g", result.Nutrition.Energy, result.Nutrition.Protein)
}

func TestFixedOptimizer_BoundaryValues0gAnd500g(t *testing.T) {
	optimizer := NewFixedOptimizer()

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "食材A", Energy: 100, Protein: 10, Fat: 5, Carbs: 20, Price: 1.0},
			{ID: 2, Name: "食材B", Energy: 200, Protein: 20, Fat: 10, Carbs: 30, Price: 2.0},
			{ID: 3, Name: "食材C", Energy: 150, Protein: 15, Fat: 8, Carbs: 25, Price: 1.5},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 500},
			{Type: "ingredient_min", IngredientID: 1, Value: 0},
			{Type: "ingredient_max", IngredientID: 2, Value: 500},
		},
		MaxIterations: 50,
	}

	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	if !result.Converged {
		t.Error("优化未收敛")
	}

	for i, ing := range result.Ingredients {
		if ing.Amount < MinIngredientAmount {
			t.Errorf("食材%d(%s)用量低于下界0g: %.2fg", i, ing.Name, ing.Amount)
		}
		if ing.Amount > MaxIngredientAmount {
			t.Errorf("食材%d(%s)用量超出上界500g: %.2fg", i, ing.Name, ing.Amount)
		}
	}

	t.Logf("边界值测试结果:")
	for _, ing := range result.Ingredients {
		t.Logf("  %s: %.2fg (范围: 0-500g)", ing.Name, ing.Amount)
	}
}

func TestFixedOptimizer_ConvergenceDeviation(t *testing.T) {
	optimizer := NewFixedOptimizer()

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "鸡胸肉", Energy: 165, Protein: 31, Fat: 3.6, Carbs: 0, Price: 0.8},
			{ID: 2, Name: "西兰花", Energy: 34, Protein: 2.8, Fat: 0.4, Carbs: 6.6, Price: 0.3},
			{ID: 3, Name: "米饭", Energy: 130, Protein: 2.6, Fat: 0.3, Carbs: 28, Price: 0.1},
		},
		NutritionGoals: []NutritionGoal{
			{Nutrient: "energy", Target: 550, Min: 500, Max: 700, Weight: 0.5},
			{Nutrient: "protein", Target: 30, Min: 20, Max: 40, Weight: 0.5},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
		MaxIterations: 100,
		Tolerance:     0.05,
	}

	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	if !result.Converged {
		t.Error("优化未收敛")
	}

	energyDeviation := math.Abs(result.Nutrition.Energy-req.NutritionGoals[0].Target) / req.NutritionGoals[0].Target
	proteinDeviation := math.Abs(result.Nutrition.Protein-req.NutritionGoals[1].Target) / req.NutritionGoals[1].Target

	avgDeviation := (energyDeviation + proteinDeviation) / 2
	if avgDeviation > 0.05 {
		t.Errorf("平均收敛偏差超过5%%: 能量%.2f%%, 蛋白质%.2f%%, 平均%.2f%%",
			energyDeviation*100, proteinDeviation*100, avgDeviation*100)
	}

	t.Logf("收敛偏差测试:")
	t.Logf("  能量目标: %.2f kcal, 实际: %.2f kcal, 偏差: %.2f%%",
		req.NutritionGoals[0].Target, result.Nutrition.Energy, energyDeviation*100)
	t.Logf("  蛋白质目标: %.2f g, 实际: %.2f g, 偏差: %.2f%%",
		req.NutritionGoals[1].Target, result.Nutrition.Protein, proteinDeviation*100)
}

func TestFixedOptimizer_DeterministicResults(t *testing.T) {
	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "食材A", Energy: 100, Protein: 10, Fat: 5, Carbs: 20, Price: 1.0},
			{ID: 2, Name: "食材B", Energy: 200, Protein: 20, Fat: 10, Carbs: 30, Price: 2.0},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 300},
		},
		MaxIterations: 50,
	}

	var results []*OptimizationResult
	for i := 0; i < 3; i++ {
		optimizer := NewFixedOptimizer()
		result, err := optimizer.Optimize(req)
		if err != nil {
			t.Fatalf("第%d次优化失败: %v", i+1, err)
		}
		results = append(results, result)
	}

	for i := 1; i < len(results); i++ {
		if len(results[0].Ingredients) != len(results[i].Ingredients) {
			t.Errorf("第%d次运行结果食材数量不一致", i+1)
			continue
		}

		for j := range results[0].Ingredients {
			if math.Abs(results[0].Ingredients[j].Amount-results[i].Ingredients[j].Amount) > 0.01 {
				t.Errorf("第%d次运行结果不一致: 食材%d用量 %.2f vs %.2f",
					i+1, j, results[0].Ingredients[j].Amount, results[i].Ingredients[j].Amount)
			}
		}
	}

	t.Logf("确定性测试: 多次运行结果一致")
}

func TestFixedOptimizer_RepairConstraints(t *testing.T) {
	optimizer := NewFixedOptimizer()

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "食材A", Energy: 100, Protein: 10, Price: 1.0},
			{ID: 2, Name: "食材B", Energy: 200, Protein: 20, Price: 2.0},
			{ID: 3, Name: "食材C", Energy: 150, Protein: 15, Price: 1.5},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
			{Type: "ingredient_min", IngredientID: 1, Value: 50},
			{Type: "ingredient_max", IngredientID: 2, Value: 100},
		},
	}

	ind := &FixedIndividual{Amounts: []float64{30, 200, 300}}

	optimizer.repair(ind, req)

	if ind.Amounts[0] < 50 {
		t.Errorf("食材1最小值约束未满足: %.2f < 50", ind.Amounts[0])
	}

	if ind.Amounts[1] > 100 {
		t.Errorf("食材2最大值约束未满足: %.2f > 100", ind.Amounts[1])
	}

	for i, amount := range ind.Amounts {
		if amount < MinIngredientAmount || amount > MaxIngredientAmount {
			t.Errorf("食材%d用量超出范围[0,500]: %.2fg", i, amount)
		}
	}

	t.Logf("约束修复测试:")
	t.Logf("  食材1: %.2fg (最小约束: 50g)", ind.Amounts[0])
	t.Logf("  食材2: %.2fg (最大约束: 100g)", ind.Amounts[1])
	t.Logf("  食材3: %.2fg", ind.Amounts[2])
}

func TestFixedOptimizer_NoWarningsOnSuccess(t *testing.T) {
	optimizer := NewFixedOptimizer()

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "鸡胸肉", Energy: 165, Protein: 31, Fat: 3.6, Carbs: 0, Price: 0.8},
			{ID: 2, Name: "西兰花", Energy: 34, Protein: 2.8, Fat: 0.4, Carbs: 6.6, Price: 0.3},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 400},
		},
		MaxIterations: 50,
	}

	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	if result.Error != "" {
		t.Errorf("成功优化不应返回错误: %s", result.Error)
	}

	for _, ing := range result.Ingredients {
		if ing.Amount == 1000 {
			t.Error("不应返回极端值1000g")
		}
	}

	t.Logf("成功优化无错误: converged=%v, error='%s'", result.Converged, result.Error)
}

func TestFixedOptimizer_MultipleIngredients(t *testing.T) {
	optimizer := NewFixedOptimizer()

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "鸡胸肉", Energy: 165, Protein: 31, Fat: 3.6, Carbs: 0, Price: 0.8},
			{ID: 2, Name: "西兰花", Energy: 34, Protein: 2.8, Fat: 0.4, Carbs: 6.6, Price: 0.3},
			{ID: 3, Name: "米饭", Energy: 130, Protein: 2.6, Fat: 0.3, Carbs: 28, Price: 0.1},
			{ID: 4, Name: "胡萝卜", Energy: 41, Protein: 0.9, Fat: 0.2, Carbs: 9.6, Price: 0.2},
			{ID: 5, Name: "鸡蛋", Energy: 155, Protein: 13, Fat: 11, Carbs: 1.1, Price: 0.5},
		},
		NutritionGoals: []NutritionGoal{
			{Nutrient: "energy", Target: 600, Weight: 0.4},
			{Nutrient: "protein", Target: 35, Weight: 0.6},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 500},
		},
		MaxIterations: 100,
	}

	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("优化失败: %v", err)
	}

	if !result.Converged {
		t.Error("优化未收敛")
	}

	totalAmount := 0.0
	for _, ing := range result.Ingredients {
		if ing.Amount < MinIngredientAmount || ing.Amount > MaxIngredientAmount {
			t.Errorf("食材%s用量超出范围: %.2fg", ing.Name, ing.Amount)
		}
		totalAmount += ing.Amount
	}

	t.Logf("多食材优化结果:")
	t.Logf("  食材数量: %d", len(result.Ingredients))
	t.Logf("  总重量: %.2fg", totalAmount)
	t.Logf("  能量: %.2f kcal", result.Nutrition.Energy)
	t.Logf("  蛋白质: %.2f g", result.Nutrition.Protein)
	t.Logf("  成本: %.2f 元", result.Cost)
}
