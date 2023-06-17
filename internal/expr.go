package internal

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"
)

func doEval(expr string, vars EvalVariables) (string, error) {
	// Handle parentheses first. Go from left to right and innermost first and
	// replace each with a result
	var err error
	var matched, matched_something bool
	var result = strings.TrimSpace(expr)

	// If "just some number" (as string), then return that number (as string).
	isNumber := RE_NUM_ONLY.MatchString(result)
	if isNumber {
		return result, nil
	}

	// If the string is a name of existing variable, return that variable's
	// value right now.
	value, was_variable := vars[result]
	if was_variable {
		return value, nil
	}

	matched_something = false

	for {
		matched = false
		result, err = ReplaceAllStringSubmatchFunc(RE_PAR, result, func(groups []string) (string, error) {
			matched = true
			matched_something = true
			return doEval(groups[1], vars)
		}, -1) // Don't limit processing of multiple non-overlapping matches consecutively.

		if err != nil {
			return "", err
		}

		if !matched {
			break
		}
	}

	// Then handle multiplication/division. Go from left to right and replace
	// each with a result
	for {
		matched = false
		result, err = ReplaceAllStringSubmatchFunc(RE_MUL, result, func(groups []string) (string, error) {
			matched = true
			matched_something = true
			return binaryOpEval(doMul, groups, vars)
		}, 1) // Limit to processing only the first match found (associativity is then processed correctly).

		if err != nil {
			return "", err
		}

		if !matched {
			break
		}
	}

	// Then handle addition/subtraction. Go from left to right and replace
	// each with a result.
	for {
		matched = false
		result, err = ReplaceAllStringSubmatchFunc(RE_ADD, result, func(groups []string) (string, error) {
			matched = true
			matched_something = true
			return binaryOpEval(doAdd, groups, vars)
		}, 1) // Limit to processing only the first match found (associativity is then processed correctly).

		if err != nil {
			return "", err
		}

		if !matched {
			break
		}
	}

	// None of the expected rules above matched our string - we don't know
	// what it is.
	if !matched_something {
		return "", fmt.Errorf("cannot parse expression '%s'", expr)
	}

	return result, nil

}

func doTerminal(a string, vars EvalVariables) (string, error) {

	isNumber := RE_NUM_ONLY.MatchString(a)
	if isNumber {
		return a, nil
	}

	var value, ok = vars[a]
	if !ok {
		return "", fmt.Errorf("undefined variable '%s'", a)
	}

	return value, nil

}

func doMul(a string, op string, b string) (string, error) {

	aDecimal, err := decimal.NewFromString(a)
	if err != nil {
		return "", fmt.Errorf("cannot convert '%s' to float", a)
	}

	bDecimal, err := decimal.NewFromString(b)
	if err != nil {
		return "", fmt.Errorf("cannot convert '%s' to float", b)
	}

	var result decimal.Decimal
	if op == "*" {
		result = aDecimal.Mul(bDecimal)
	} else {
		if bDecimal.Cmp(DecimalZero) == 0 {
			return "", errors.New("division by zero")
		}
		result = aDecimal.Div(bDecimal)
	}

	return result.String(), nil

}

func doAdd(a string, op string, b string) (string, error) {

	aDecimal, err := decimal.NewFromString(a)
	if err != nil {
		return "", fmt.Errorf("cannot convert '%s' to float", a)
	}

	bDecimal, err := decimal.NewFromString(b)
	if err != nil {
		return "", fmt.Errorf("cannot convert '%s' to float", b)
	}

	var result decimal.Decimal
	if op == "+" {
		result = aDecimal.Add(bDecimal)
	} else {
		result = aDecimal.Sub(bDecimal)
	}

	return result.String(), nil

}

func binaryOpEval(
	fn binaryOpEvalCallback,
	groups []string,
	vars EvalVariables,
) (string, error) {
	var err error
	var termA string
	var termB string

	termA, err = doTerminal(groups[1], vars)
	if err != nil {
		return "", err
	}

	termB, err = doTerminal(groups[3], vars)
	if err != nil {
		return "", err
	}

	val, err := fn(termA, groups[2], termB)
	if err != nil {
		return "", err
	}

	return val, nil
}

func evalExpression(expr string, vars EvalVariables) (string, error) {
	return doEval(expr, vars)
}

func expandExpressions(str string, vars EvalVariables) (string, error) {
	var err error

	result := RE_BRACED_EXPR.ReplaceAllStringFunc(str, func(match string) string {
		_result, _err := doEval(match[2:len(match)-1], vars)

		// Bubble up the error, if there's any (we can't do that via returning,
		// because regexp.ReplaceAllStringFunc() expects only a string to be
		// returned).
		if _err != nil {
			err = _err
		}

		return _result
	})

	return result, err
}

var DecimalZero, _ = decimal.NewFromString("0")

const __RE_NUM = `-?\d+(?:\.\d+)?`
const __RE_NUM_ONLY = `^` + __RE_NUM + `$`
const __RE_VAR = `[a-zA-Z][a-zA-Z0-9_.]*`
const __RE_OP = `(?:(?:` + __RE_VAR + `)|(?:` + __RE_NUM + `))`
const __RE_MUL = `(` + __RE_OP + `)\s*([*\/])\s*(` + __RE_OP + `)`
const __RE_ADD = `(` + __RE_OP + `)\s*([+-])\s*(` + __RE_OP + `)`
const __RE_PAR = `\(\s*([^\(]*?)\s*\)`
const __RE_BRACED_EXPR = `\$\{.*?\}`

var RE_NUM = regexp.MustCompile(__RE_NUM)
var RE_NUM_ONLY = regexp.MustCompile(__RE_NUM_ONLY)
var RE_VAR = regexp.MustCompile(__RE_VAR)
var RE_OP = regexp.MustCompile(__RE_OP)
var RE_MUL = regexp.MustCompile(__RE_MUL)
var RE_ADD = regexp.MustCompile(__RE_ADD)
var RE_PAR = regexp.MustCompile(__RE_PAR)
var RE_BRACED_EXPR = regexp.MustCompile(__RE_BRACED_EXPR)
