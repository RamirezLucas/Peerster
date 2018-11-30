package data

import "fmt"

// Bitmap represents a bitmap of arbitrary size (NOT thread-safe)
type Bitmap struct {
	size uint64 // Bitmap's size in bits
	data []byte // Byte array containing the actual bitmap
}

// NewBitmap creates a new instance of NameIndex with the indicated size
func NewBitmap(size uint64) *Bitmap {

	// Compute required size for data array
	dataSize := size / 8
	if size%8 != 0 {
		dataSize++
	}

	var bitmap Bitmap
	bitmap.size = size
	bitmap.data = make([]byte, dataSize)
	return &bitmap
}

// GetSize returns the bitmap's size
func (bitmap *Bitmap) GetSize() uint64 {
	if bitmap == nil {
		panic("GetSize called on nil Bitmap")
	}
	return bitmap.size
}

// GetBit returns the bit's value at index as a boolean (true for 1, false for 0)
func (bitmap *Bitmap) GetBit(index uint64) bool {
	if bitmap == nil {
		panic("GetBit called on nil Bitmap")
	}
	if index >= bitmap.size {
		panic(fmt.Sprintf("GetBit called on non-existent bit %d in bitmap of size %d", index, bitmap.size))
	}

	return ((1 << (index % 8)) & bitmap.data[index/8]) != 0
}

// SetBit sets the bit at index and returns the old value as a boolean (true for 1, false for 0)
func (bitmap *Bitmap) SetBit(index uint64) bool {
	if bitmap == nil {
		panic("SetBit called on nil Bitmap")
	}
	if index >= bitmap.size {
		panic(fmt.Sprintf("SetBit called on non-existent bit %d in bitmap of size %d\n", index, bitmap.size))
	}

	arrayIndex := index / 8
	bitmask := byte(1 << (index % 8))

	// Set index and return
	oldVal := (bitmap.data[arrayIndex] & bitmask) != 0
	bitmap.data[arrayIndex] = bitmap.data[arrayIndex] | bitmask
	return oldVal
}

// UnsetBit unsets the bit at index and returns the old value as a boolean (true for 1, false for 0)
func (bitmap *Bitmap) UnsetBit(index uint64) bool {
	if bitmap == nil {
		panic("UnsetBit called on nil Bitmap")
	}
	if index >= bitmap.size {
		panic(fmt.Sprintf("UnsetBit called on non-existent bit %d in bitmap of size %d\n", index, bitmap.size))
	}

	arrayIndex := index / 8
	bitmask := byte(1 << (index % 8))

	// Unset index and return
	oldVal := (bitmap.data[arrayIndex] & bitmask) != 0
	bitmap.data[arrayIndex] = bitmap.data[arrayIndex] & (byte(0xFF) ^ bitmask)
	return oldVal
}

// CountLeadingBits returns the number of bits set up to (and excluding) upToIndex
func (bitmap *Bitmap) CountLeadingBits(upToIndex uint64) uint64 {
	if bitmap == nil {
		panic("CountLeadingBits called on nil Bitmap")
	}
	if upToIndex > bitmap.size {
		panic(fmt.Sprintf("CountLeadingBits called on non-existent bit %d in bitmap of size %d\n", upToIndex, bitmap.size))
	}

	count := uint64(0)

	arraySize := upToIndex / 8
	lastElement := int(upToIndex % 8)
	if lastElement != 0 {
		arraySize++
	}

	// Iterate over array elements
	for i := uint64(0); i < arraySize; i++ {

		// Get element and set up bitmask
		data := bitmap.data[i]
		bitmask := byte(1)

		// Set maximum index
		maxIndex := 8
		if i == arraySize-1 && lastElement != 0 {
			maxIndex = lastElement
		}

		// Count the 1's
		for j := 0; j < maxIndex; j++ {
			if (data & bitmask) != 0 {
				count++
			}
			bitmask <<= 1
		}
	}

	return count
}
