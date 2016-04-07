package tick

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"

	//"log"
)

var ErrEmptyStack = errors.New("stack is empty")

type stackItem struct {
	IsGeneric    bool
	GenericValue interface{}

	IsInt    bool
	IntValue int64

	IsFloat    bool
	FloatValue float64

	IsString    bool
	StringValue string

	IsBool    bool
	BoolValue bool

	IsRegex    bool
	RegexValue *regexp.Regexp
}

// Value returns the value of the stack item as generic value
func (s stackItem) Value() interface{} {
	if s.IsFloat {
		return s.FloatValue
	}

	if s.IsInt {
		return s.IntValue
	}

	if s.IsBool {
		return s.BoolValue
	}

	if s.IsString {
		return s.StringValue
	}

	if s.IsRegex {
		return s.RegexValue
	}

	if s.IsGeneric {
		return s.GenericValue
	}

	return nil
}

func newStackItem(v interface{}) stackItem {
	switch value := v.(type) {
	case bool:
		return stackItem{
			IsBool:    true,
			BoolValue: value,
		}

	case int:
		return stackItem{
			IsInt:    true,
			IntValue: int64(value),
		}
	case int64:
		return stackItem{
			IsInt:    true,
			IntValue: value,
		}
	case float64:
		return stackItem{
			IsFloat:    true,
			FloatValue: value,
		}
	case string:
		return stackItem{
			IsString:    true,
			StringValue: value,
		}
	default:
		return stackItem{
			IsGeneric:    true,
			GenericValue: v,
		}
	}
}

type stack struct {
	data []stackItem
}

func (s *stack) Len() int {
	return len(s.data)
}

func (s *stack) Push(v interface{}) {
	s.data = append(s.data, newStackItem(v))
}

func (s *stack) PushInt64(v int64) {
	s.data = append(s.data, stackItem{IsInt: true, IntValue: v})
}

func (s *stack) PushFloat64(v float64) {
	s.data = append(s.data, stackItem{IsFloat: true, FloatValue: v})
}

func (s *stack) PushBool(v bool) {
	s.data = append(s.data, stackItem{IsBool: true, BoolValue: v})
}

func (s *stack) PushString(v string) {
	s.data = append(s.data, stackItem{IsString: true, StringValue: v})
}

func (s *stack) PushRegex(v *regexp.Regexp) {
	s.data = append(s.data, stackItem{IsRegex: true, RegexValue: v})
}

func (s *stack) popStackItem() (stackItem, error) {
	if s.Len() > 0 {
		l := s.Len() - 1
		v := s.data[l]
		s.data = s.data[:l]
		return v, nil
	}

	return stackItem{}, ErrEmptyStack
}

func (s *stack) PopItem() stackItem {
	stackItem, err := s.popStackItem()
	if err != nil {
		panic(err)
	}

	return stackItem
}

func (s *stack) Pop() interface{} {
	stackItem, err := s.popStackItem()
	if err != nil {
		panic(err)
	}

	return stackItem.Value()
}

func (s *stack) String() string {
	var str bytes.Buffer
	str.Write([]byte("s["))
	for i := len(s.data) - 1; i >= 0; i-- {
		fmt.Fprintf(&str, "%T:%v,", s.data[i].Value(), s.data[i].Value())
	}
	str.Write([]byte("]"))
	return str.String()
}
