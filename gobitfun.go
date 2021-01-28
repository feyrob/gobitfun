package gobitfun

import (
	"encoding/binary"
	"math"
	"math/bits"
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

var Po__general_purpose__prefix_1_count_to_value_bit_count []int = []int{
	1,  // 0 and 1 compress nicely // good enough for sparse and good for bool storage
	5,  // id deltas
	8,  // remaining 8 bit values; nice random u8
	12, // less aggressive step to encode human numbers (stuff smaller than 4000)
	16, // nice random u16
	20, // less aggressive step to encode human numbers (stuff smaller than 1 000 000)
	32, // nice random u32
	64, // nice random u64
}
var Po__general_purpose__prefix_1_count_to_value_offset []uint64 = Get_prefix_1_count_to_value_offset(Po__general_purpose__prefix_1_count_to_value_bit_count)

var Po__cer__prefix_1_count_to_value_bit_count []int = []int{
	3,
	8,
	16,
	24,
	32,
	64,
}
var Po__cer__prefix_1_count_to_value_offset []uint64 = Get_prefix_1_count_to_value_offset(Po__cer__prefix_1_count_to_value_bit_count)

func Po_encode_u64(b *[]byte, prefix_1_count_to_value_bit_count []int, prefix_1_count_to_value_offset []uint64, v uint64, Next_bit_offset *uint64) {
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
		Push_bit(b, 1, Next_bit_offset)
	}
	Push_bit(b, 0, Next_bit_offset)

	for current_v_bit_idx := value_bit_count - 1; current_v_bit_idx >= 0; current_v_bit_idx-- {
		b1 := (offset_adjusted_v >> current_v_bit_idx) % 2
		Push_bit(b, int(b1), Next_bit_offset)
	}
}

func Bit_len(b []byte, Next_bit_offset uint64) uint64 {
	free_bit_count := (8 - Next_bit_offset) % 8
	bit_count := uint64(len(b))*8 - free_bit_count
	return bit_count
}

func Push_bit(b *[]byte, bit int, Next_bit_offset *uint64) {
	current_bit_offset := *Next_bit_offset
	if current_bit_offset == 0 {
		*b = append(*b, 0)
	}

	*Next_bit_offset = (*Next_bit_offset + 1) % 8

	last_byte := (*b)[len(*b)-1]

	bit_idx_to_set := 7 - current_bit_offset
	bit_to_set := byte(bit << bit_idx_to_set)
	last_byte |= bit_to_set

	(*b)[len(*b)-1] = last_byte
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

func Sorted_id_list_to_delta_list(l []uint64) []uint64 {
	delta_l := []uint64{}

	prev := int64(-1)
	for i := 0; i < len(l); i++ {
		cur := l[i]
		delta := uint64(int64(cur) - prev)
		delta--
		delta_l = append(delta_l, delta)
		prev = int64(cur)
	}
	return delta_l
}

func Delta_list_to_sorted_id_list(delta_l []uint64) []uint64 {
	l := []uint64{}
	cur := int64(-1)
	for i := 0; i < len(delta_l); i++ {
		delta := delta_l[i]
		cur = cur + int64(delta) + 1
		l = append(l, uint64(cur))
	}
	return l
}

func Encode_f64_map(m map[uint32]float64) []byte {
	used_sub_maps_mask := uint8(0)
	c0_key_list := make([]uint64, 0)
	c1_key_list := make([]uint32, 0)
	ce_uint_map := make(map[uint32][]byte)   // continuation encoded uints
	ce_negint_map := make(map[uint32][]byte) // continuation encoded uints * -1
	f32_map := make(map[uint32]float32)
	f64_map := make(map[uint32]float64)

	for key_id, f64 := range m {
		if f64 == 0 {
			c0_key_list = append(c0_key_list, uint64(key_id))
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

	bit_offset := len(out_buf) * 8

	if len(c0_key_list) != 0 {
		Fib_encode(uint64(len(c0_key_list)), &out_buf, &bit_offset)
	}
	if len(c1_key_list) != 0 {
		Fib_encode(uint64(len(c1_key_list)), &out_buf, &bit_offset)
	}
	if len(ce_uint_map) != 0 {
		Fib_encode(uint64(len(ce_uint_map)), &out_buf, &bit_offset)
	}
	if len(ce_negint_map) != 0 {
		Fib_encode(uint64(len(ce_negint_map)), &out_buf, &bit_offset)
	}
	if len(f32_map) != 0 {
		Fib_encode(uint64(len(f32_map)), &out_buf, &bit_offset)
	}
	if len(f64_map) != 0 {
		Fib_encode(uint64(len(f64_map)), &out_buf, &bit_offset)
	}

	// c0 list
	if len(c0_key_list) != 0 {
		sort.Slice(c0_key_list, func(i, j int) bool { return c0_key_list[i] < c0_key_list[j] })

		//Delta_list_to_sorted_id_list()
		// for _, c0_key := range c0_key_list {
		// 	Fib_encode(uint64(c0_key), &out_buf, &bit_offset)
		// }
		c0_key_delta_list := Sorted_id_list_to_delta_list(c0_key_list)
		for _, delta := range c0_key_delta_list {
			Fib_encode(delta, &out_buf, &bit_offset)
		}
	}

	// c1 list
	if len(c1_key_list) != 0 {
		sort.Slice(c1_key_list, func(i, j int) bool { return c1_key_list[i] < c1_key_list[j] })
		for _, c1_key := range c1_key_list {
			Fib_encode(uint64(c1_key), &out_buf, &bit_offset)
		}
	}

	// uint map
	if len(ce_uint_map) != 0 {

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

func Decode_f64_map(b []byte, next *uint64) map[uint32]float64 {

	f64_map := make(map[uint32]float64)

	used_sub_maps_mask := b[*next]
	*next++

	used_subcontainer_count := bits.OnesCount8(uint8(used_sub_maps_mask))

	bit_idx := int((*next) * 8)
	subcontainer_lengths := make([]uint64, used_subcontainer_count)
	for i := 0; i < used_subcontainer_count; i++ {
		n := Fib_decode(b, &bit_idx)
		subcontainer_lengths[i] = n
	}
	cur_subcontainer_idx := 0
	//*next = uint64((bit_idx + 7) / 8) // round up to full bytes

	// c0 list
	if used_sub_maps_mask&1 == 1 {
		c0_list_len := subcontainer_lengths[cur_subcontainer_idx]
		cur_subcontainer_idx++
		remaining_c0_entry_count := c0_list_len

		// for remaining_c0_entry_count > 0 {
		// 	key_id := Fib_decode(b, &bit_idx)
		// 	//key_id := Cont_decode(b, next)
		// 	f64_map[uint32(key_id)] = 0.0
		// 	remaining_c0_entry_count--
		// }
		key_id_delta_list := []uint64{}
		for remaining_c0_entry_count > 0 {
			delta := Fib_decode(b, &bit_idx)
			key_id_delta_list = append(key_id_delta_list, delta)
			remaining_c0_entry_count--
		}
		sorted_id_list := Delta_list_to_sorted_id_list(key_id_delta_list)
		for _, key_id := range sorted_id_list {
			f64_map[uint32(key_id)] = 0.0
		}
	}
	used_sub_maps_mask = used_sub_maps_mask >> 1

	// c1 list
	if used_sub_maps_mask&1 == 1 {
		c1_list_len := subcontainer_lengths[cur_subcontainer_idx]
		cur_subcontainer_idx++

		remaining_c1_entry_count := c1_list_len
		for remaining_c1_entry_count > 0 {
			key_id := Fib_decode(b, &bit_idx)
			//key_id := Cont_decode(b, next)
			f64_map[uint32(key_id)] = 1.0
			remaining_c1_entry_count--
		}
	}
	used_sub_maps_mask = used_sub_maps_mask >> 1

	// done with fib encoded numbers, round up to full bytes:
	*next = uint64((bit_idx + 7) / 8) // round up to full bytes

	// uint map
	if used_sub_maps_mask&1 == 1 {
		uint_map_len := subcontainer_lengths[cur_subcontainer_idx]
		cur_subcontainer_idx++

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
		negint_map_len := subcontainer_lengths[cur_subcontainer_idx]
		cur_subcontainer_idx++

		remaining_uint_entry_count := negint_map_len
		for remaining_uint_entry_count > 0 {
			n := *next
			_ = n
			key_id := Cont_decode(b, next)
			val := Cont_decode(b, next)
			f64_map[uint32(key_id)] = -float64(val)
			remaining_uint_entry_count--
		}
	}
	used_sub_maps_mask = used_sub_maps_mask >> 1

	// f32 map
	if used_sub_maps_mask&1 == 1 {
		f32_map_len := subcontainer_lengths[cur_subcontainer_idx]
		cur_subcontainer_idx++

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
		f64_map_len := subcontainer_lengths[cur_subcontainer_idx]
		cur_subcontainer_idx++

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
