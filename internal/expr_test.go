package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testEvalExpr(
	t *testing.T,
	vars EvalVariables,
	expr string,
	result string,
) {

	r, err := evalExpression(expr, vars)

	if err != nil {
		t.Logf("Expr '%s' returned error '%s'", expr, err.Error())
		t.FailNow()
	}
	if r != result {
		t.Logf("Expr '%s' should evaluate to '%s'", expr, result)
		t.FailNow()
	}

}

func TestEvalExprOk(t *testing.T) {

	var vars = make(EvalVariables)

	testEvalExpr(t, vars, "1", "1")
	testEvalExpr(t, vars, "  1  ", "1")
	testEvalExpr(t, vars, "  1+1  ", "2")
	testEvalExpr(t, vars, "  1     +            2  ", "3")
	testEvalExpr(t, vars, "  (1     +2) *           2  ", "6")
	testEvalExpr(t, vars, "28+782-287*27+87-287*827*78+7-72", "-18520139")
	testEvalExpr(t, vars, "1 + 2 - (3 * 4) + 5 - (6 * 7 * 8) + 9 - 10", "-341")
	testEvalExpr(t, vars, "28 + 782 - (287 * 27) + 87 - (287 * 827 * 78) + 7 - 72", "-18520139")
	testEvalExpr(t, vars, "1+2 * 3 / 4*-3 +1*-12/-4", "-0.5")
	testEvalExpr(t, vars, "(1 + 2) * 6", "18")
	testEvalExpr(t, vars, "1 + (2 * 6)", "13")
	testEvalExpr(t, vars, "(1 + (2) * 6)", "13")

	vars["hey"] = "HEY"
	vars["hey.dude"] = "HEY...DUDE"
	testEvalExpr(t, vars, "hey", "HEY")
	testEvalExpr(t, vars, "hey.dude", "HEY...DUDE")

}

func TestEvalExprParseError(t *testing.T) {

	var vars = make(EvalVariables)

	_, err := evalExpression("1.=", vars)
	assert.ErrorContains(t, err, "cannot parse expression")
	_, err = evalExpression(".", vars)
	assert.ErrorContains(t, err, "cannot parse expression")
	_, err = evalExpression("-", vars)
	assert.ErrorContains(t, err, "cannot parse expression")
	_, err = evalExpression("a a", vars)
	assert.ErrorContains(t, err, "cannot parse expression")

}

func TestEvalExprMissingVars(t *testing.T) {

	var vars = make(EvalVariables)
	vars["ahoj"] = "1"

	_, err := evalExpression("1 + hello", vars)
	assert.EqualError(t, err, "undefined variable 'hello'")

}

func TestEvalExprDivZero(t *testing.T) {

	var vars = make(EvalVariables)
	var err error

	_, err = evalExpression("1 / 0", vars)
	assert.EqualError(t, err, "division by zero")

	_, err = evalExpression("1 / (4 * 4 - 16)", vars)
	assert.EqualError(t, err, "division by zero")

}

func BenchmarkEvalExpr_simple_number(b *testing.B) {
	const expr = "123456789"
	for i := 0; i < b.N; i++ {
		_, _ = evalExpression(expr, make(EvalVariables))
	}
	b.ReportAllocs()
}

func BenchmarkEvalExpr_simple_binaryAdd_Longer(b *testing.B) {
	const expr = "123456789 + 987654321"
	for i := 0; i < b.N; i++ {
		_, _ = evalExpression(expr, make(EvalVariables))
	}
	b.ReportAllocs()
}

func BenchmarkEvalExpr_complex(b *testing.B) {
	const expr = "28 + 782 - (287 * 27) + 87 - (287 * 827 * 78) + 7 - 72 - 89 * (123 / 4 - 75 / (1 + 2) * (45 + 89 / 3) - (12 + 7 * 3))"
	for i := 0; i < b.N; i++ {
		_, _ = evalExpression(expr, make(EvalVariables))
	}
	b.ReportAllocs()
}
