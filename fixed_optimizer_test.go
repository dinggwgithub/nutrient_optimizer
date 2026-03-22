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
		{50.0, 50.0},
		{-10.0, 0.0},
		{600.0, 500.0},
		{250.0, 250.0},
		{0.0, 0.0},
		{500.0, 500.0},
	}

	for _, test := range tests {
		result := optimizer.clampAmount(test.input)
		if result != test.expected {
			t.Errorf("clampAmount(%f) = %f, expected %f", test.input, result, test.expected)
		}
	}
}

func TestFixedOptimizer_ClampNutrition(t *testing.T) {
	optimizer := NewFixedOptimizer()

	tests := []struct {
		input    float64
		name     string
		expected float64
	}{
		{100.0, "energy", 100.0},
		{-50.0, "protein", 0.0},
		{2e6, "fat", MaxNutritionValue},
		{math.NaN(), "carbs", 0.0},
		{math.Inf(1), "calcium", MaxNutritionValue},
		{math.Inf(-1), "iron", 0.0},
	}

	for _, test := range tests {
		result := optimizer.clampNutrition(test.input, test.name)
		if math.IsNaN(test.input) || math.IsInf(test.input, 0) {
			if result != test.expected {
				t.Errorf("clampNutrition(%f, %s) = %f, expected %f", test.input, test.name, result, test.expected)
			}
		} else if result != test.expected {
			t.Errorf("clampNutrition(%f, %s) = %f, expected %f", test.input, test.name, result, test.expected)
		}
	}
}

func TestFixedOptimizer_SafeMultiply(t *testing.T) {
	optimizer := NewFixedOptimizer()

	tests := []struct {
		a, b     float64
		expected float64
	}{
		{10.0, 5.0, 50.0},
		{0.0, 100.0, 0.0},
		{math.NaN(), 5.0, 0.0},
		{5.0, math.NaN(), 0.0},
		{math.Inf(1), 5.0, MaxNutritionValue},
		{5.0, math.Inf(1), MaxNutritionValue},
	}

	for _, test := range tests {
		result := optimizer.safeMultiply(test.a, test.b)
		if result != test.expected {
			t.Errorf("safeMultiply(%f, %f) = %f, expected %f", test.a, test.b, result, test.expected)
		}
	}
}

func TestFixedOptimizer_SafeDivide(t *testing.T) {
	optimizer := NewFixedOptimizer()

	tests := []struct {
		a, b     float64
		expected float64
	}{
		{100.0, 10.0, 10.0},
		{100.0, 0.0, 0.0},
		{math.NaN(), 10.0, 0.0},
		{10.0, math.NaN(), 0.0},
		{math.Inf(1), 10.0, MaxNutritionValue},
	}

	for _, test := range tests {
		result := optimizer.safeDivide(test.a, test.b)
		if result != test.expected {
			t.Errorf("safeDivide(%f, %f) = %f, expected %f", test.a, test.b, result, test.expected)
		}
	}
}

func TestFixedOptimizer_SafeLog(t *testing.T) {
	optimizer := NewFixedOptimizer()

	tests := []struct {
		input    float64
		expected float64
	}{
		{math.E, 1.0},
		{1.0, 0.0},
		{0.0, 0.0},
		{-1.0, 0.0},
		{math.NaN(), 0.0},
	}

	for _, test := range tests {
		result := optimizer.safeLog(test.input)
		if test.input > 0 && !math.IsNaN(test.input) && test.input != math.E {
			if math.Abs(result-test.expected) > 0.0001 {
				t.Errorf("safeLog(%f) = %f, expected %f", test.input, result, test.expected)
			}
		} else if result != test.expected {
			t.Errorf("safeLog(%f) = %f, expected %f", test.input, result, test.expected)
		}
	}
}

func TestFixedOptimizer_SafeExp(t *testing.T) {
	optimizer := NewFixedOptimizer()

	tests := []struct {
		input    float64
		expected float64
	}{
		{0.0, 1.0},
		{1.0, math.E},
		{800.0, MaxNutritionValue},
		{-800.0, 0.0},
		{math.NaN(), 1.0},
	}

	for _, test := range tests {
		result := optimizer.safeExp(test.input)
		if test.input == 1.0 {
			if math.Abs(result-test.expected) > 0.0001 {
				t.Errorf("safeExp(%f) = %f, expected %f", test.input, result, test.expected)
			}
		} else if result != test.expected {
			t.Errorf("safeExp(%f) = %f, expected %f", test.input, result, test.expected)
		}
	}
}

func TestFixedOptimizer_Optimize(t *testing.T) {
	optimizer := NewFixedOptimizer()

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "小麦", Energy: 338, Protein: 11.9, Fat: 1.3, Carbs: 75.2, Calcium: 34, Iron: 5.1, Zinc: 2.33, VitaminC: 0, Price: 0},
			{ID: 2, Name: "五谷香", Energy: 378, Protein: 9.9, Fat: 2.6, Carbs: 78.9, Calcium: 2, Iron: 0.5, Zinc: 0.23, VitaminC: 0, Price: 0},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 300},
		},
	}

	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("Optimize failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if !result.Converged {
		t.Error("Result should be converged")
	}

	if len(result.Ingredients) != 2 {
		t.Errorf("Expected 2 ingredients, got %d", len(result.Ingredients))
	}

	if math.IsNaN(result.Nutrition.Energy) {
		t.Error("Energy should not be NaN")
	}
	if math.IsInf(result.Nutrition.Energy, 0) {
		t.Error("Energy should not be Inf")
	}
	if result.Nutrition.Protein < 0 {
		t.Error("Protein should not be negative")
	}
	if result.Nutrition.Energy < 0 {
		t.Error("Energy should not be negative")
	}
	if result.Nutrition.Fat < 0 {
		t.Error("Fat should not be negative")
	}

	t.Logf("Energy: %.2f", result.Nutrition.Energy)
	t.Logf("Protein: %.2f", result.Nutrition.Protein)
	t.Logf("Fat: %.2f", result.Nutrition.Fat)
	t.Logf("Carbs: %.2f", result.Nutrition.Carbs)
}

func TestFixedOptimizer_OptimizeWithEmptyIngredients(t *testing.T) {
	optimizer := NewFixedOptimizer()

	req := OptimizationRequest{
		Ingredients: []Ingredient{},
	}

	_, err := optimizer.Optimize(req)
	if err == nil {
		t.Error("Expected error for empty ingredients")
	}
}

func TestFixedOptimizer_GenerateABTestReport(t *testing.T) {
	optimizer := NewFixedOptimizer()

	buggyResult := &OptimizationResult{
		Nutrition: NutritionSummary{
			Energy:  math.Inf(1),
			Protein: math.NaN(),
			Fat:     math.Inf(1),
			Carbs:   231.15,
		},
		Error: "NUMERICAL_OVERFLOW_BUG_DETECTED",
	}

	fixedResult := &OptimizationResult{
		Nutrition: NutritionSummary{
			Energy:  500.0,
			Protein: 30.0,
			Fat:     15.0,
			Carbs:   80.0,
		},
	}

	report := optimizer.GenerateABTestReport(buggyResult, fixedResult)

	if report == nil {
		t.Fatal("Report is nil")
	}

	buggyData, ok := report["buggy"].(map[string]interface{})
	if !ok {
		t.Fatal("buggy data is not a map")
	}

	if !buggyData["has_inf"].(bool) {
		t.Error("buggy result should have Inf values")
	}
	if !buggyData["has_nan"].(bool) {
		t.Error("buggy result should have NaN values")
	}

	fixedData, ok := report["fixed"].(map[string]interface{})
	if !ok {
		t.Fatal("fixed data is not a map")
	}

	if fixedData["has_nan"].(bool) {
		t.Error("fixed result should not have NaN values")
	}
	if fixedData["has_inf"].(bool) {
		t.Error("fixed result should not have Inf values")
	}
	if fixedData["has_negative"].(bool) {
		t.Error("fixed result should not have negative values")
	}

	comparison, ok := report["comparison"].(map[string]interface{})
	if !ok {
		t.Fatal("comparison data is not a map")
	}

	if !comparison["all_valid"].(bool) {
		t.Error("fixed result should be all valid")
	}
}

func TestFixedOptimizer_IsNumericalSafe(t *testing.T) {
	optimizer := NewFixedOptimizer()

	tests := []struct {
		input    float64
		expected bool
	}{
		{100.0, true},
		{0.0, true},
		{-50.0, true},
		{math.NaN(), false},
		{math.Inf(1), false},
		{math.Inf(-1), false},
	}

	for _, test := range tests {
		result := optimizer.IsNumericalSafe(test.input)
		if result != test.expected {
			t.Errorf("IsNumericalSafe(%f) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestFixedOptimizer_ValidateInput(t *testing.T) {
	optimizer := NewFixedOptimizer()

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "测试食材", Energy: 100, Protein: 10, Fat: 5, Carbs: 20},
			{ID: 2, Name: "无效食材", Energy: math.NaN(), Protein: -1, Fat: math.Inf(1), Carbs: 0},
		},
	}

	errors := optimizer.ValidateInput(req)

	if len(errors) == 0 {
		t.Error("Expected validation errors for invalid input")
	}

	t.Logf("Validation errors: %v", errors)
}

func TestFixedOptimizer_CalculateNutritionSafe(t *testing.T) {
	optimizer := NewFixedOptimizer()

	ingredients := []IngredientAmount{
		{
			Ingredient: Ingredient{ID: 1, Name: "小麦", Energy: 338, Protein: 11.9, Fat: 1.3, Carbs: 75.2, Calcium: 34, Iron: 5.1, Zinc: 2.33, VitaminC: 0},
			Amount:     150,
		},
		{
			Ingredient: Ingredient{ID: 2, Name: "五谷香", Energy: 378, Protein: 9.9, Fat: 2.6, Carbs: 78.9, Calcium: 2, Iron: 0.5, Zinc: 0.23, VitaminC: 0},
			Amount:     150,
		},
	}

	nutrition := optimizer.calculateNutritionSafe(ingredients)

	if math.IsNaN(nutrition.Energy) || math.IsInf(nutrition.Energy, 0) {
		t.Errorf("Energy is invalid: %f", nutrition.Energy)
	}
	if math.IsNaN(nutrition.Protein) || math.IsInf(nutrition.Protein, 0) || nutrition.Protein < 0 {
		t.Errorf("Protein is invalid: %f", nutrition.Protein)
	}
	if math.IsNaN(nutrition.Fat) || math.IsInf(nutrition.Fat, 0) || nutrition.Fat < 0 {
		t.Errorf("Fat is invalid: %f", nutrition.Fat)
	}

	expectedEnergy := 338*1.5 + 378*1.5
	if math.Abs(nutrition.Energy-expectedEnergy) > 0.01 {
		t.Errorf("Energy calculation error: expected %.2f, got %.2f", expectedEnergy, nutrition.Energy)
	}

	t.Logf("Nutrition - Energy: %.2f, Protein: %.2f, Fat: %.2f, Carbs: %.2f",
		nutrition.Energy, nutrition.Protein, nutrition.Fat, nutrition.Carbs)
}

func TestFixedOptimizer_ExtremeValues(t *testing.T) {
	optimizer := NewFixedOptimizer()

	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 1, Name: "极端食材", Energy: 1e10, Protein: 1e10, Fat: 1e10, Carbs: 1e10, Price: 1e10},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 100},
		},
	}

	result, err := optimizer.Optimize(req)
	if err != nil {
		t.Fatalf("Optimize failed: %v", err)
	}

	if result.Nutrition.Energy > MaxNutritionValue {
		t.Errorf("Energy should be clamped to MaxNutritionValue, got %f", result.Nutrition.Energy)
	}
	if result.Nutrition.Protein > MaxNutritionValue {
		t.Errorf("Protein should be clamped to MaxNutritionValue, got %f", result.Nutrition.Protein)
	}
	if result.Nutrition.Fat > MaxNutritionValue {
		t.Errorf("Fat should be clamped to MaxNutritionValue, got %f", result.Nutrition.Fat)
	}

	t.Logf("Clamped values - Energy: %.2f, Protein: %.2f, Fat: %.2f",
		result.Nutrition.Energy, result.Nutrition.Protein, result.Nutrition.Fat)
}

func TestNumericalOverflowComparison(t *testing.T) {
	req := OptimizationRequest{
		Ingredients: []Ingredient{
			{ID: 2, Name: "小麦", Energy: 338, Protein: 11.9, Fat: 1.3, Carbs: 75.2, Calcium: 34, Iron: 5.1, Zinc: 2.33, VitaminC: 0, Price: 0},
			{ID: 3, Name: "五谷香", Energy: 378, Protein: 9.9, Fat: 2.6, Carbs: 78.9, Calcium: 2, Iron: 0.5, Zinc: 0.23, VitaminC: 0, Price: 0},
		},
		Constraints: []Constraint{
			{Type: "total_weight", Value: 300},
		},
	}

	buggyOptimizer := NewBuggyOptimizer("numerical_overflow")
	buggyResult, _ := buggyOptimizer.Optimize(req)

	fixedOptimizer := NewFixedOptimizer()
	fixedResult, _ := fixedOptimizer.Optimize(req)

	t.Log("=== A/B Test Comparison ===")
	t.Logf("Buggy Energy: %e", buggyResult.Nutrition.Energy)
	t.Logf("Fixed Energy: %.2f", fixedResult.Nutrition.Energy)
	t.Logf("Buggy Protein: %e", buggyResult.Nutrition.Protein)
	t.Logf("Fixed Protein: %.2f", fixedResult.Nutrition.Protein)
	t.Logf("Buggy Fat: %e", buggyResult.Nutrition.Fat)
	t.Logf("Fixed Fat: %.2f", fixedResult.Nutrition.Fat)

	if math.IsInf(buggyResult.Nutrition.Energy, 0) && !math.IsInf(fixedResult.Nutrition.Energy, 0) {
		t.Log("✅ Energy overflow fixed")
	}
	if buggyResult.Nutrition.Protein < 0 && fixedResult.Nutrition.Protein >= 0 {
		t.Log("✅ Protein negative value fixed")
	}
	if math.IsInf(buggyResult.Nutrition.Fat, 0) && !math.IsInf(fixedResult.Nutrition.Fat, 0) {
		t.Log("✅ Fat overflow fixed")
	}

	if fixedResult.Nutrition.Energy < 0 || fixedResult.Nutrition.Energy > MaxNutritionValue {
		t.Error("Fixed Energy is out of valid range")
	}
	if fixedResult.Nutrition.Protein < 0 || fixedResult.Nutrition.Protein > MaxNutritionValue {
		t.Error("Fixed Protein is out of valid range")
	}
	if fixedResult.Nutrition.Fat < 0 || fixedResult.Nutrition.Fat > MaxNutritionValue {
		t.Error("Fixed Fat is out of valid range")
	}
}
