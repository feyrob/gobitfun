package gobitfun

func Cont_encode(v uint64) []byte {
	r := make([]byte, 0, 4)

	// highest bit set to 1
	end_byte := uint8(v & 0x7f)
	r = append(r, end_byte)

	for remaining := v >> 7; remaining != 0; remaining = remaining >> 7 {
		// highest bit set to 1
		cont_byte := uint8(remaining | 0x80)
		r = append(r, cont_byte)
	}

	for i := len(r)/2 - 1; i >= 0; i-- {
		op := len(r) - 1 - i
		r[i], r[op] = r[op], r[i]
	}
	return r
}

func Cont_decode(bytes_ []byte, next_idx *uint64) uint64 {
	r := uint64(0)
	for {
		b := bytes_[*next_idx]
		*next_idx = (*next_idx) + 1
		r = r << 7
		r = r | uint64(b&0x7f) // set bottom 7 bits
		// if top bit was set continue
		if b < 128 {
			break
		}
	}
	return r
}
