package visual

type bitmap []uint64

const BITMAP_SHIFT = 6
const BITMAP_MASK = 63

func wordsForBits(bits uint) uint {
	return bits/64 + 1
}

func newBitmap(bits uint) bitmap {
	return make([]uint64, wordsForBits(bits))
}

func (b bitmap) clear() {
	for i := range b {
		b[i] = 0
	}
}

func (b bitmap) set(index uint, state bool) {
	if state {
		b[index>>BITMAP_SHIFT] |= 1 << (index & BITMAP_MASK)
	} else {
		b[index>>BITMAP_SHIFT] &^= 1 << (index & BITMAP_MASK)
	}
}

func (b bitmap) get(index uint) bool {
	return (b[index>>BITMAP_SHIFT]>>(index&BITMAP_MASK))&1 > 0
}
