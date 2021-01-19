package gobitfun

import (
	"encoding/binary"
	"math"
	"sort"
)

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

func Set_bit(b []byte, idx int) {
	byte_idx := idx / 8
	bit_offset := idx % 8
	bit := byte(1 << (7 - bit_offset))
	b[byte_idx] = b[byte_idx] | bit
}

// does not support 0
// does not support values larger than 2^63
func Fib_encode(n uint64, dst_buf *[]byte, bit_len *int) {
	cur_fib := uint64(1)
	next_fib := uint64(2)
	fib_idx := 0

	for next_fib <= n {
		cur_fib += next_fib
		cur_fib, next_fib = next_fib, cur_fib
		fib_idx++
	}

	bits_needed := fib_idx + 2
	bit_len_needed := bits_needed + *bit_len
	byte_len_needed := (bit_len_needed + 7) / 8

	if byte_len_needed < cap(*dst_buf) {
		*dst_buf = (*dst_buf)[0:byte_len_needed]
	} else {
		// buf has too little capacity
		new_buf := make([]byte, byte_len_needed, byte_len_needed*2)
		copy(new_buf, *dst_buf)
		*dst_buf = new_buf
	}

	// if len(*dst_buf) < byte_len_needed {
	// 	// buf has too little capacity
	// 	new_buf := make([]byte, byte_len_needed, int(float32(byte_len_needed)*1.5))
	// 	copy(new_buf, *dst_buf)
	// 	*dst_buf = new_buf
	// }

	dst_offset := *bit_len

	rest := n
	for fib_idx >= 0 {

		if cur_fib <= rest {
			Set_bit(*dst_buf, dst_offset+fib_idx)
			rest -= cur_fib
		}

		next_fib -= cur_fib
		cur_fib, next_fib = next_fib, cur_fib
		fib_idx--
	}

	new_dst_len := dst_offset + bits_needed

	Set_bit(*dst_buf, new_dst_len-1)

	*bit_len = new_dst_len
}

func Fib_decode(buf []byte, bit_idx *int) uint64 {
	prev_fib := uint64(1)
	cur_fib := uint64(1)

	maybe_terminate := false

	value := uint64(0)

	for {
		byte_idx := *bit_idx / 8
		b := buf[byte_idx]
		bit_offset := *bit_idx % 8

		bit_mask := byte(1) << (7 - bit_offset)
		if (b & bit_mask) == 0 {
			// is 0
			maybe_terminate = false
		} else {
			// is 1
			if maybe_terminate {
				// we have found 11
				*bit_idx++
				break
			} else {
				value += cur_fib
				maybe_terminate = true
			}
		}

		cur_fib = prev_fib + cur_fib
		prev_fib = cur_fib - prev_fib

		// vs
		//prev_fib, cur_fib = cur_fib, prev_fib+cur_fib

		*bit_idx++
	}
	return value
}

func Deserialize_f64_map(b []byte, next *uint64) map[uint32]float64 {

	f64_map := make(map[uint32]float64)

	used_sub_maps_mask := b[*next]
	*next++

	// c0 list
	if used_sub_maps_mask&1 == 1 {
		c0_list_len := Cont_decode(b, next)
		remaining_c0_entry_count := c0_list_len
		for remaining_c0_entry_count > 0 {
			key_id := Cont_decode(b, next)
			f64_map[uint32(key_id)] = 0.0
			remaining_c0_entry_count--
		}
	}
	used_sub_maps_mask = used_sub_maps_mask >> 1

	// c1 list
	if used_sub_maps_mask&1 == 1 {
		c1_list_len := Cont_decode(b, next)
		remaining_c1_entry_count := c1_list_len
		for remaining_c1_entry_count > 0 {
			key_id := Cont_decode(b, next)
			f64_map[uint32(key_id)] = 1.0
			remaining_c1_entry_count--
		}
	}
	used_sub_maps_mask = used_sub_maps_mask >> 1

	// uint map
	if used_sub_maps_mask&1 == 1 {
		uint_map_len := Cont_decode(b, next)
		remaining_uint_entry_count := uint_map_len
		for remaining_uint_entry_count > 0 {
			key_id := Cont_decode(b, next)
			val := Cont_decode(b, next)
			f64_map[uint32(key_id)] = float64(val)
			remaining_uint_entry_count--
		}
	}
	used_sub_maps_mask = used_sub_maps_mask >> 1

	// negint map
	if used_sub_maps_mask&1 == 1 {
		negint_map_len := Cont_decode(b, next)
		remaining_uint_entry_count := negint_map_len
		for remaining_uint_entry_count > 0 {
			key_id := Cont_decode(b, next)
			val := Cont_decode(b, next)
			f64_map[uint32(key_id)] = -float64(val)
			remaining_uint_entry_count--
		}
	}
	used_sub_maps_mask = used_sub_maps_mask >> 1

	// f32 map
	if used_sub_maps_mask&1 == 1 {
		f32_map_len := Cont_decode(b, next)
		remaining_f32_entry_count := f32_map_len
		for remaining_f32_entry_count > 0 {
			key_id := Cont_decode(b, next)
			f32_val_buf := b[*next : *next+4]
			*next += 4
			f32_val_bits := binary.LittleEndian.Uint32(f32_val_buf)
			f32_val := math.Float32frombits(f32_val_bits)
			f64_map[uint32(key_id)] = float64(f32_val)
			remaining_f32_entry_count--
		}
	}
	used_sub_maps_mask = used_sub_maps_mask >> 1

	// f64 map
	if used_sub_maps_mask&1 == 1 {
		f64_map_len := Cont_decode(b, next)
		remaining_f64_entry_count := f64_map_len
		for remaining_f64_entry_count > 0 {
			key_id := Cont_decode(b, next)
			f64_val_buf := b[*next : *next+8]
			*next += 8
			f64_val_bits := binary.LittleEndian.Uint64(f64_val_buf)
			f64_val := math.Float64frombits(f64_val_bits)
			f64_map[uint32(key_id)] = f64_val
			remaining_f64_entry_count--
		}
	}
	used_sub_maps_mask = used_sub_maps_mask >> 1

	return f64_map
}

func Serialize_f64_map(m map[uint32]float64) []byte {
	used_sub_maps_mask := uint8(0)
	c0_key_list := make([]uint32, 0)
	c1_key_list := make([]uint32, 0)
	ce_uint_map := make(map[uint32][]byte)   // continuation encoded uints
	ce_negint_map := make(map[uint32][]byte) // continuation encoded uints * -1
	f32_map := make(map[uint32]float32)
	f64_map := make(map[uint32]float64)

	for key_id, f64 := range m {
		if f64 == 0 {
			c0_key_list = append(c0_key_list, key_id)
			continue
		}
		if f64 == 1 {
			c1_key_list = append(c1_key_list, key_id)
			continue
		}

		// if it is a positive integer and it can be stored in less than 4 bytes, then do so
		u32 := uint32(f64)
		if f64 == float64(u32) {
			b := Cont_encode(uint64(u32))
			if len(b) < 4 {
				ce_uint_map[key_id] = b
				continue
			}
		}

		// if it is a negative integer and it can be stored in less than 4 bytes, then do so
		neg32 := uint32(-f64)
		if f64 == -float64(neg32) {
			b := Cont_encode(uint64(neg32))
			if len(b) < 4 {
				ce_negint_map[key_id] = b
				continue
			}
		}

		// if it can be stored as f32
		f32 := float32(f64)
		if f64 == float64(f32) {
			f32_map[key_id] = f32
			continue

		}

		// if it is a positive integer and it can be stored in less than 8 bytes, then do so
		u64 := uint64(f64)
		if f64 == float64(u64) {
			b := Cont_encode(u64)
			if len(b) < 8 {
				ce_uint_map[key_id] = b
				continue
			}
		}

		// if it is a negative integer and it can be stored in less than 8 bytes, then do so
		neg64 := uint64(-f64)
		if f64 == -float64(neg64) {
			b := Cont_encode(neg64)
			if len(b) < 8 {
				ce_negint_map[key_id] = b
				continue
			}
		}

		f64_map[key_id] = f64
	}

	out_buf := make([]byte, 0, 256)

	// remember which sub maps are used
	if len(f64_map) != 0 {
		used_sub_maps_mask |= 1
	}

	used_sub_maps_mask = used_sub_maps_mask << 1
	if len(f32_map) != 0 {
		used_sub_maps_mask |= 1
	}

	used_sub_maps_mask = used_sub_maps_mask << 1
	if len(ce_negint_map) != 0 {
		used_sub_maps_mask |= 1
	}

	used_sub_maps_mask = used_sub_maps_mask << 1
	if len(ce_uint_map) != 0 {
		used_sub_maps_mask |= 1
	}

	used_sub_maps_mask = used_sub_maps_mask << 1
	if len(c1_key_list) != 0 {
		used_sub_maps_mask |= 1
	}

	used_sub_maps_mask = used_sub_maps_mask << 1
	if len(c0_key_list) != 0 {
		used_sub_maps_mask |= 1
	}

	out_buf = append(out_buf, byte(used_sub_maps_mask))

	// c0 list
	if len(c0_key_list) != 0 {
		c0_len_buf := Cont_encode(uint64(len(c0_key_list)))
		out_buf = append(out_buf, c0_len_buf...)

		sort.Slice(c0_key_list, func(i, j int) bool { return c0_key_list[i] < c0_key_list[j] })
		for _, c0_key := range c0_key_list {
			key_id_buf := Cont_encode(uint64(c0_key))
			out_buf = append(out_buf, key_id_buf...)
		}
	}

	// c1 list
	if len(c1_key_list) != 0 {
		c1_len_buf := Cont_encode(uint64(len(c1_key_list)))
		out_buf = append(out_buf, c1_len_buf...)
		sort.Slice(c1_key_list, func(i, j int) bool { return c1_key_list[i] < c1_key_list[j] })
		for _, c1_key := range c1_key_list {
			key_id_buf := Cont_encode(uint64(c1_key))
			out_buf = append(out_buf, key_id_buf...)
		}
	}

	// uint map
	if len(ce_uint_map) != 0 {
		uint_len_buf := Cont_encode(uint64(len(ce_uint_map)))
		out_buf = append(out_buf, uint_len_buf...)

		uint_keys := make([]uint32, 0)
		for k, _ := range ce_uint_map {
			uint_keys = append(uint_keys, k)
		}
		sort.Slice(uint_keys, func(i, j int) bool { return uint_keys[i] < uint_keys[j] })

		for _, k := range uint_keys {
			v := ce_uint_map[k]
			key_id_buf := Cont_encode(uint64(k))
			out_buf = append(out_buf, key_id_buf...)
			out_buf = append(out_buf, v...)
		}
	}

	// negint map
	if len(ce_negint_map) != 0 {
		uint_len_buf := Cont_encode(uint64(len(ce_negint_map)))
		out_buf = append(out_buf, uint_len_buf...)

		uint_keys := make([]uint32, 0)
		for k, _ := range ce_negint_map {
			uint_keys = append(uint_keys, k)
		}
		sort.Slice(uint_keys, func(i, j int) bool { return uint_keys[i] < uint_keys[j] })

		for _, k := range uint_keys {
			v := ce_negint_map[k]
			key_id_buf := Cont_encode(uint64(k))
			out_buf = append(out_buf, key_id_buf...)
			out_buf = append(out_buf, v...)
		}
	}

	// f32 map
	if len(f32_map) != 0 {
		f32_len_buf := Cont_encode(uint64(len(f32_map)))
		out_buf = append(out_buf, f32_len_buf...)
		f32_keys := make([]uint32, 0)
		for k, _ := range f32_map {
			f32_keys = append(f32_keys, k)
		}
		sort.Slice(f32_keys, func(i, j int) bool { return f32_keys[i] < f32_keys[j] })
		for _, k := range f32_keys {
			v := f32_map[k]
			key_id_buf := Cont_encode(uint64(k))
			out_buf = append(out_buf, key_id_buf...)

			f32_buf := make([]byte, 4)
			binary.LittleEndian.PutUint32(f32_buf, math.Float32bits(v))
			out_buf = append(out_buf, f32_buf...)
		}
	}

	// f64 map
	if len(f64_map) != 0 {
		f64_len_buf := Cont_encode(uint64(len(f64_map)))
		out_buf = append(out_buf, f64_len_buf...)
		f64_keys := make([]uint32, 0)
		for k, _ := range f64_map {
			f64_keys = append(f64_keys, k)
		}
		sort.Slice(f64_keys, func(i, j int) bool { return f64_keys[i] < f64_keys[j] })
		for _, k := range f64_keys {
			v := f64_map[k]
			key_id_buf := Cont_encode(uint64(k))
			out_buf = append(out_buf, key_id_buf...)

			f64_buf := make([]byte, 8)
			binary.LittleEndian.PutUint64(f64_buf, math.Float64bits(v))
			out_buf = append(out_buf, f64_buf...)
		}
	}
	return out_buf
}

// the resulting byte array does not encode length
// past end will contain extra junk values
// PO prefixed ones
func Po_encode_u64_list(l []uint64, prefix_1_count_to_value_bit_count []int, prefix_1_count_to_value_offset []uint64) ([]byte, int) {
	a := Bit_array{}
	for _, v := range l {

		prefix_1_count := 0
		value_offset := uint64(0)

		for i := len(prefix_1_count_to_value_offset) - 1; i >= 0; i-- {
			prefix_1_count = i
			value_offset = prefix_1_count_to_value_offset[i]
			if v >= value_offset {
				break
			}
		}

		offset_adjusted_v := v - value_offset
		value_bit_count := prefix_1_count_to_value_bit_count[prefix_1_count]

		for i := 0; i < prefix_1_count; i++ {
			a.Push_bit(1)
		}
		a.Push_bit(0)

		for current_v_bit_idx := value_bit_count - 1; current_v_bit_idx >= 0; current_v_bit_idx-- {
			b := (offset_adjusted_v >> current_v_bit_idx) % 2
			a.Push_bit(int(b))
		}
	}

	//free_bit_count := (8 - a.Next_bit_offset) % 8
	//bit_count := len(a.Bytes)*8 - free_bit_count
	bit_len := a.Bit_len()
	return a.Bytes, bit_len
}

// TODO

type Bit_array struct {
	Bytes []byte
	// TODO replace with Bit_len
	Next_bit_offset int
}

func (a *Bit_array) Bit_len() int {
	free_bit_count := (8 - a.Next_bit_offset) % 8
	bit_count := len(a.Bytes)*8 - free_bit_count
	return bit_count
}

func (a *Bit_array) Set_bit(idx int) {
	byte_idx := idx / 8
	bit_offset := idx % 8

	bit := byte(1 << (7 - bit_offset))

	a.Bytes[byte_idx] = a.Bytes[byte_idx] | bit
}

func (a *Bit_array) Push_bit(bit int) {
	current_bit_offset := a.Next_bit_offset
	if current_bit_offset == 0 {
		a.Bytes = append(a.Bytes, 0)
	}

	a.Next_bit_offset = (a.Next_bit_offset + 1) % 8

	last_byte := a.Bytes[len(a.Bytes)-1]

	bit_idx_to_set := 7 - current_bit_offset
	bit_to_set := byte(bit << bit_idx_to_set)
	last_byte |= bit_to_set

	a.Bytes[len(a.Bytes)-1] = last_byte
}

func Po_decode_u64_list(b []byte, size int, prefix_1_count_to_value_bit_count []int, prefix_1_count_to_value_offset []uint64) []uint64 {

	r := []uint64{}

	current_byte_idx := 0
	current_bit_idx := 7
	for i := size - 1; i >= 0; i-- {

		// get prefix
		prefix_1_count := 0
		for {
			current_byte := b[current_byte_idx]
			current_bit := (current_byte >> current_bit_idx) % 2

			current_bit_idx--
			if current_bit_idx < 0 {
				current_bit_idx = 7
				current_byte_idx++
			}

			if current_bit == 0 {
				break
			}
			prefix_1_count++
		}

		value_bit_count := prefix_1_count_to_value_bit_count[prefix_1_count]

		// but the bits into a uint64
		offset_adjusted_v := uint64(0)

		for remaining := value_bit_count; remaining > 0; remaining-- {
			current_byte := b[current_byte_idx]
			current_bit := (current_byte >> current_bit_idx) % 2

			current_bit_idx--
			if current_bit_idx < 0 {
				current_bit_idx = 7
				current_byte_idx++
			}

			offset_adjusted_v = offset_adjusted_v << 1
			offset_adjusted_v += uint64(current_bit)
		}

		value_offset := prefix_1_count_to_value_offset[prefix_1_count]
		value := value_offset + offset_adjusted_v

		r = append(r, value)
	}
	return r
}
