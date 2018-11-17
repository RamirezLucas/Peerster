package tests

import (
	"Peerster/data"
	"testing"

	"github.com/stretchr/testify/assert"
)

const BitmapSize = 15

func TestNewBitmap(t *testing.T) {

	values := []struct {
		size   uint32
		isNull bool
	}{
		{0, true},
		{8, false},
		{15, false},
	}

	for _, val := range values {
		b := data.NewBitmap(val.size)
		if val.isNull {
			assert.Nil(t, b, "NewBitmap(%d) returned %p, expected nil", val.size, b)
		} else if !val.isNull {
			assert.NotNil(t, b, "NewBitmap(%d) returned %p, expected non-nil", val.size, b)
		}
	}
}

func TestGetSetUnsetBadArguments(t *testing.T) {

	b := data.NewBitmap(BitmapSize)

	values := []struct {
		bitmap *data.Bitmap
		index  uint32
	}{
		{nil, 0},
		{b, BitmapSize},
		{b, BitmapSize + 1},
	}

	for _, val := range values {
		assert.Panics(t, func() { val.bitmap.GetBit(val.index) },
			"%p.GetBit(%d) should have panicked (bitmap size: %d)", val.bitmap, val.index, BitmapSize)
	}
	for _, val := range values {
		assert.Panics(t, func() { val.bitmap.SetBit(val.index) },
			"%p.SetBit(%d) should have panicked (bitmap size: %d)", val.bitmap, val.index, BitmapSize)
	}
	for _, val := range values {
		assert.Panics(t, func() { val.bitmap.UnsetBit(val.index) },
			"%p.UnsetBit(%d) should have panicked (bitmap size: %d)", val.bitmap, val.index, BitmapSize)
	}
}

func TestCountLeadingBitsBadArguments(t *testing.T) {

	b := data.NewBitmap(BitmapSize)

	values := []struct {
		bitmap *data.Bitmap
		index  uint32
	}{
		{nil, 0},
		{b, BitmapSize + 1},
	}

	for _, val := range values {
		assert.Panics(t, func() { val.bitmap.CountLeadingBits(val.index) },
			"%p.CountLeadingBits(%d) should have panicked (bitmap size: %d)", val.bitmap, val.index, BitmapSize)
	}
}

func TestGetAndSet(t *testing.T) {

	b := data.NewBitmap(BitmapSize)

	setValues := []struct {
		index    uint32
		expected bool
	}{
		{0, false},
		{4, false},
		{BitmapSize - 1, false},
	}

	for _, val := range setValues {
		ret := b.SetBit(val.index)
		assert.Equal(t, val.expected, ret, "SetBit(%d) returned %t, expected %t", val.index, ret, val.expected)
	}

	getValues := []struct {
		index    uint32
		expected bool
	}{
		{0, true},
		{3, false},
		{4, true},
		{5, false},
		{BitmapSize - 1, true},
	}

	for _, val := range getValues {
		ret := b.GetBit(val.index)
		assert.Equal(t, val.expected, ret, "GetBit(%d) returned %t, expected %t", val.index, ret, val.expected)
	}
}

func TestSetAndUnset(t *testing.T) {

	b := data.NewBitmap(BitmapSize)

	setValues := []struct {
		index    uint32
		expected bool
	}{
		{4, false},
		{BitmapSize - 1, false},
	}

	for _, val := range setValues {
		ret := b.SetBit(val.index)
		assert.Equal(t, val.expected, ret, "SetBit(%d) returned %t, expected %t", val.index, ret, val.expected)
	}

	unsetValues := []struct {
		index    uint32
		expected bool
	}{
		{4, true},
		{5, false},
		{BitmapSize - 1, true},
	}

	for _, val := range unsetValues {
		ret := b.UnsetBit(val.index)
		assert.Equal(t, val.expected, ret, "GetBit(%d) returned %t, expected %t", val.index, ret, val.expected)
	}

	getValues := []struct {
		index    uint32
		expected bool
	}{
		{4, false},
		{5, false},
		{BitmapSize - 1, false},
	}

	for _, val := range getValues {
		ret := b.GetBit(val.index)
		assert.Equal(t, val.expected, ret, "GetBit(%d) returned %t, expected %t", val.index, ret, val.expected)
	}

}

func TestLeadingCountBits(t *testing.T) {

	b := data.NewBitmap(BitmapSize)

	// Set values
	setValues := []uint32{0, 1, 2, 3, 8, 9, 10, 12, 14}
	for _, val := range setValues {
		b.SetBit(val)
	}

	indexCount := []struct {
		index    uint32
		expected uint32
	}{
		{0, 0},
		{3, 3},
		{5, 4},
		{13, 8},
		{BitmapSize, uint32(len(setValues))},
	}

	for _, val := range indexCount {
		ret := b.CountLeadingBits(val.index)
		assert.Equal(t, val.expected, ret, "CountLeadingBits(%d) returned %d, expected %d", val.index, ret, val.expected)
	}
}
