package tick

import (
	"errors"
	"fmt"
	"math"
	"regexp"
)

var ErrInvalidExpr = errors.New("expression is invalid, could not evaluate")

// Expression functions are stateful. Their state is updated with
// each call to the function. A StatefulExpr is a Node
// and its associated function state.
type StatefulExpr struct {
	Node  Node
	Funcs Funcs
}

func NewStatefulExpr(n Node) *StatefulExpr {
	return &StatefulExpr{
		Node:  n,
		Funcs: NewFunctions(),
	}
}

// Reset the state
func (s *StatefulExpr) Reset() {
	for _, f := range s.Funcs {
		f.Reset()
	}
}

func (s *StatefulExpr) EvalBool(scope *Scope) (bool, error) {
	stck := &stack{}
	err := s.eval(s.Node, scope, stck)
	if err != nil {
		return false, err
	}
	if stck.Len() == 1 {
		valueItem := stck.PopItem()

		switch {
		case valueItem.IsBool:
			return valueItem.BoolValue, nil
		default:
			return false, fmt.Errorf("expression returned unexpected type %T", valueItem.Value())
		}
	}

	return false, ErrInvalidExpr
}

func (s *StatefulExpr) EvalNum(scope *Scope) (interface{}, error) {
	stck := &stack{}
	err := s.eval(s.Node, scope, stck)
	if err != nil {
		return math.NaN(), err
	}
	if stck.Len() == 1 {
		value := stck.Pop()
		// Resolve reference
		if ref, ok := value.(*ReferenceNode); ok {
			value, err = scope.Get(ref.Reference)
			if err != nil {
				return math.NaN(), err
			}
		}
		switch value.(type) {
		case float64, int64:
			return value, nil
		default:
			return math.NaN(), fmt.Errorf("expression returned unexpected type %T", value)
		}
	}
	return math.NaN(), ErrInvalidExpr
}

func (s *StatefulExpr) eval(n Node, scope *Scope, stck *stack) (err error) {
	switch node := n.(type) {

	case *ReferenceNode:
		refValue, err := scope.Get(node.Reference)
		if err != nil {
			return err
		}
		stck.Push(refValue)
	case *BoolNode:
		stck.PushBool(node.Bool)
	case *NumberNode:
		if node.IsInt {
			stck.PushInt64(node.Int64)
		} else {
			stck.PushFloat64(node.Float64)
		}
	case *DurationNode:
		stck.Push(node.Dur)
	case *StringNode:
		stck.PushString(node.Literal)
	case *RegexNode:
		stck.PushRegex(node.Regex)
	case *UnaryNode:
		err = s.eval(node.Node, scope, stck)
		if err != nil {
			return
		}
		s.evalUnary(node.Operator, scope, stck)
	case *BinaryNode:
		err = s.eval(node.Left, scope, stck)
		if err != nil {
			return
		}
		err = s.eval(node.Right, scope, stck)
		if err != nil {
			return
		}
		err = s.evalBinary(node.Operator, scope, stck)
		if err != nil {
			return
		}
	case *FunctionNode:
		args := make([]interface{}, len(node.Args))
		for i, arg := range node.Args {
			err = s.eval(arg, scope, stck)
			if err != nil {
				return
			}
			a := stck.Pop()
			if r, ok := a.(*ReferenceNode); ok {
				a, err = scope.Get(r.Reference)
				if err != nil {
					return err
				}
			}
			args[i] = a
		}
		// Call function
		f := s.Funcs[node.Func]
		if f == nil {
			return fmt.Errorf("undefined function %s", node.Func)
		}
		ret, err := f.Call(args...)
		if err != nil {
			return fmt.Errorf("error calling %s: %s", node.Func, err)
		}
		stck.Push(ret)
	default:
		stck.Push(node)
	}
	return nil
}

func (s *StatefulExpr) evalUnary(op TokenType, scope *Scope, stck *stack) error {
	v := stck.Pop()
	switch op {
	case TokenMinus:
		switch n := v.(type) {
		case float64:
			stck.Push(-1 * n)
		case int64:
			stck.Push(-1 * n)
		default:
			return fmt.Errorf("invalid arugument to '-' %v", v)
		}
	case TokenNot:
		if b, ok := v.(bool); ok {
			stck.Push(!b)
		} else {
			return fmt.Errorf("invalid arugument to '!' %v", v)
		}
	}
	return nil
}

func errMismatched(op TokenType, l, r interface{}) error {
	return fmt.Errorf("mismatched type to binary operator. got %T %v %T. see bool(), int(), float()", l, op, r)
}

func (s *StatefulExpr) evalBinary(op TokenType, scope *Scope, stck *stack) (err error) {
	rItem := stck.PopItem()
	lItem := stck.PopItem()

	var v interface{}
	switch {

	case isMathOperator(op):
		// No tests, no changes!
		r := rItem.Value()
		l := lItem.Value()
		switch ln := l.(type) {
		case int64:
			rn, ok := r.(int64)
			if !ok {
				return errMismatched(op, l, r)
			}
			v, err = doIntMath(op, ln, rn)
		case float64:
			rn, ok := r.(float64)
			if !ok {
				return errMismatched(op, l, r)
			}
			v, err = doFloatMath(op, ln, rn)
		default:
			return errMismatched(op, l, r)
		}

		if err != nil {
			return
		}

		stck.Push(v)
		return
	case isCompOperator(op):
		var compareResult bool
		switch {
		case lItem.IsBool:
			if rItem.IsBool {
				compareResult, err = doBoolComp(op, lItem.BoolValue, rItem.BoolValue)
			} else {
				return errMismatched(op, lItem.BoolValue, rItem.Value())
			}

		case lItem.IsString:
			if rItem.IsString {
				compareResult, err = doStringComp(op, lItem.StringValue, rItem.StringValue)
			} else if rItem.IsRegex {
				compareResult, err = doRegexComp(op, lItem.StringValue, rItem.RegexValue)
			} else {
				return errMismatched(op, lItem.StringValue, rItem.Value())
			}

		case lItem.IsFloat:
			ln := lItem.FloatValue
			var rf float64
			switch {
			case rItem.IsInt:
				rf = float64(rItem.IntValue)
			case rItem.IsFloat:
				rf = rItem.FloatValue
			default:
				return errMismatched(op, lItem.FloatValue, rItem.Value())
			}
			compareResult, err = doFloatComp(op, ln, rf)

		case lItem.IsInt:
			switch {

			// If both sides are int64, we will do int64 comparsion
			case rItem.IsInt:
				// Left and right are int64
				compareResult, err = doIntComp(op, lItem.IntValue, rItem.IntValue)

			// The right side is float64, we will do float64 comparison
			case rItem.IsFloat:
				lf := float64(lItem.IntValue)
				rf := rItem.FloatValue
				compareResult, err = doFloatComp(op, lf, rf)
			default:
				return errMismatched(op, lItem.IntValue, rItem.Value())
			}

		default:
			return errMismatched(op, lItem.Value(), rItem.Value())
		}

		if err != nil {
			return
		}

		stck.PushBool(compareResult)
		return
	default:
		return fmt.Errorf("return: unknown operator %v", op)
	}

}

func doIntMath(op TokenType, l, r int64) (v int64, err error) {
	switch op {
	case TokenPlus:
		v = l + r
	case TokenMinus:
		v = l - r
	case TokenMult:
		v = l * r
	case TokenDiv:
		v = l / r
	case TokenMod:
		v = l % r
	default:
		return 0, fmt.Errorf("invalid integer math operator %v", op)
	}
	return
}

func doFloatMath(op TokenType, l, r float64) (v float64, err error) {
	switch op {
	case TokenPlus:
		v = l + r
	case TokenMinus:
		v = l - r
	case TokenMult:
		v = l * r
	case TokenDiv:
		v = l / r
	default:
		return math.NaN(), fmt.Errorf("invalid float math operator %v", op)
	}
	return
}

func doBoolComp(op TokenType, l, r bool) (v bool, err error) {
	switch op {
	case TokenEqual:
		v = l == r
	case TokenNotEqual:
		v = l != r
	case TokenAnd:
		v = l && r
	case TokenOr:
		v = l || r
	default:
		err = fmt.Errorf("invalid boolean comparison operator %v", op)
	}
	return
}

func doFloatComp(op TokenType, l, r float64) (v bool, err error) {
	switch op {
	case TokenEqual:
		v = l == r
	case TokenNotEqual:
		v = l != r
	case TokenLess:
		v = l < r
	case TokenGreater:
		v = l > r
	case TokenLessEqual:
		v = l <= r
	case TokenGreaterEqual:
		v = l >= r
	default:
		err = fmt.Errorf("invalid float comparison operator %v", op)
	}
	return
}

func doIntComp(op TokenType, l, r int64) (v bool, err error) {
	switch op {
	case TokenEqual:
		v = l == r
	case TokenNotEqual:
		v = l != r
	case TokenLess:
		v = l < r
	case TokenGreater:
		v = l > r
	case TokenLessEqual:
		v = l <= r
	case TokenGreaterEqual:
		v = l >= r
	default:
		err = fmt.Errorf("invalid int comparison operator %v", op)
	}
	return
}

func doStringComp(op TokenType, l, r string) (v bool, err error) {
	switch op {
	case TokenEqual:
		v = l == r
	case TokenNotEqual:
		v = l != r
	case TokenLess:
		v = l < r
	case TokenGreater:
		v = l > r
	case TokenLessEqual:
		v = l <= r
	case TokenGreaterEqual:
		v = l >= r
	default:
		err = fmt.Errorf("invalid string comparison operator %v", op)
	}
	return
}

func doRegexComp(op TokenType, l string, r *regexp.Regexp) (v bool, err error) {
	switch op {
	case TokenRegexEqual:
		v = r.MatchString(l)
	case TokenRegexNotEqual:
		v = !r.MatchString(l)
	default:
		err = fmt.Errorf("invalid regex comparison operator %v", op)
	}
	return
}
