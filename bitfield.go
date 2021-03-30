package hypercore

import (
	"github.com/kiambogo/go-hypercore/mempager"
)

type Bitfield struct {
	pager      *mempager.Pager
	byteLength uint64
}

func NewBitfield(pageSize int) *Bitfield {
	pgr := mempager.NewPager(pageSize)
	return &Bitfield{pager: &pgr}
}

// PageSize returns the size of the pages used by the internal pager
func (b Bitfield) PageSize() int {
	return b.pager.PageSize()
}

// ByteLength returns the number of bytes in the bitfield
func (b Bitfield) ByteLength() uint64 {
	return b.byteLength
}

// Len returns the number of bits set in the bitfield
func (b Bitfield) Len() uint64 {
	return b.byteLength * 8
}

// SetBit sets the bit at a particular index within the bitfield
// Returns true if a change was inacted
func (b *Bitfield) SetBit(index int, value bool) bool {
	byteIndex := uint64(index / 8) // 8 bits in a byte
	byteAtOffset := b.GetByte(byteIndex)

	bitIndex := byte(1 << (index % 8))

	if bitAlreadySet := byteAtOffset&bitIndex == 1; bitAlreadySet {
		return false
	}
	updatedByte := byteAtOffset | bitIndex

	return b.SetByte(byteIndex, updatedByte)
}

// SetByte sets the byte at a particular index within the bitfield
// Returns true if a change was inacted
func (b *Bitfield) SetByte(index uint64, value byte) bool {
	pageIndex := index / uint64(b.pager.PageSize())
	offset := index % uint64(b.pager.PageSize())
	page := b.pager.GetOrAlloc(int(pageIndex))
	pageBuffer := page.Buffer()

	// Update the byte length of the bitfield
	if index >= b.byteLength {
		b.byteLength = index + 1
	}

	if (*pageBuffer)[offset] == value {
		return false
	}
	(*pageBuffer)[offset] = value
	return true
}

// GetBit returns the value of the bit at a provided index
func (b *Bitfield) GetBit(index uint64) bool {
	byteAtOffset := b.GetByte((index / 8))
	bitIndex := byte(1 << (index % 8))

	return byteAtOffset&bitIndex > 0
}

// GetByte returns the value of the byte at a provided index
func (b *Bitfield) GetByte(index uint64) byte {
	byteOffset := b.pageMask(index)
	pageIndex := index / uint64(b.pager.PageSize())
	page := b.pager.Get(int(pageIndex))
	if page == nil {
		return byte(0)
	}
	pageBuffer := page.Buffer()
	return (*pageBuffer)[byteOffset]
}

func (b Bitfield) pageMask(index uint64) uint64 {
	return index & (uint64(b.pager.PageSize()) - 1)
}
