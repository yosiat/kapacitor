package tick

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStackLen(t *testing.T) {
	assert := assert.New(t)
	st := &stack{}

	assert.Equal(st.Len(), 0, "Initial stack length must be zero")
}

func TestStackEmptyPop(t *testing.T) {
	assert := assert.New(t)
	st := &stack{}

	assert.Panics(func() {
		st.Pop()
	}, "Pop on empty stack should panic")
}

func TestStackPushLenIncreased(t *testing.T) {
	assert := assert.New(t)
	st := &stack{}

	st.Push(1)

	assert.Equal(st.Len(), 1, "Stack should contain one element")
}

func TestStackPushPop(t *testing.T) {
	assert := assert.New(t)
	st := &stack{}

	st.Push(1)
	st.Push(2)

	assert.Equal(st.Len(), 2, "After push-pop, stack should be empty")

	popedValue := st.Pop()
	assert.Equal(popedValue, 2, "Poped value should 1")

	popedValue = st.Pop()
	assert.Equal(popedValue, 1, "Poped value should 1")
}

func TestStackString(t *testing.T) {
	assert := assert.New(t)
	st := &stack{}

	assert.Equal(st.String(), "s[]", "Empty stack")

	st.Push(1)

	assert.Equal(st.String(), "s[int:1,]", "Stack with one element")

}
