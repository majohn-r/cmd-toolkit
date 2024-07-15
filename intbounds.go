package cmd_toolkit

import "sort"

// IntBounds holds the bounds for an int value which has a minimum value, a
// maximum value, and a default that lies within those bounds
type IntBounds struct {
	MinValue     int
	DefaultValue int
	MaxValue     int
}

// NewIntBounds creates an instance of IntBounds, sorting the provided value into
// reasonable fields
func NewIntBounds(v1, v2, v3 int) *IntBounds {
	v := []int{v1, v2, v3}
	sort.Ints(v)
	return &IntBounds{
		MinValue:     v[0],
		DefaultValue: v[1],
		MaxValue:     v[2],
	}
}

func (b *IntBounds) constrainedValue(value int) (i int) {
	switch {
	case value < b.MinValue:
		i = b.MinValue
	case value > b.MaxValue:
		i = b.MaxValue
	default:
		i = value
	}
	return
}
