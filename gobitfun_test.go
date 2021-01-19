package gobitfun

import (
	"fmt"
	"math"
	"math/bits"
	"strconv"
	"testing"

	"encoding/binary"

	"reflect"
	"sort"
)

func Test_fib_encode_decode(t *testing.T) {
	//l := []uint64{1, 9, 10, 11, math.MaxUint64 / 2}
	l := []uint64{1, 2, 9, 10, 11}

	b := []byte{}
	bit_len := 0

	for _, v := range l {
		Fib_encode(v, &b, &bit_len)
	}

	s := Get_bit_str(b)
	fmt.Println("fibenc(", l, ") =", s, "bit_len: ", bit_len)

	bit_idx := 0
	restored_l := []uint64{}
	for remaining := len(l); remaining > 0; remaining-- {
		n := Fib_decode(b, &bit_idx)
		restored_l = append(restored_l, n)
	}

	eq := reflect.DeepEqual(l, restored_l)
	if eq {
		fmt.Println("yay")
	} else {
		fmt.Println("fixme")
	}
}

func Test_cont_encode_decode(t *testing.T) {
	return
	ranges := [][]uint64{
		[]uint64{0, 256*256 + 2},
		[]uint64{256*256*256 - 1000, 256*256*256 + 1000},
	}

	tested_count := uint64(0)

	for _, r := range ranges {
		start := r[0]
		end := r[1]

		for i := start; i < end; i++ {
			test_number := i
			cont_encoded_buf := Cont_encode(test_number)

			next_idx := uint64(0)
			restored_2 := Cont_decode(cont_encoded_buf, &next_idx)
			if restored_2 != test_number {
				t.Fail()
			}
			tested_count++
		}
		fmt.Println("tested_count:", tested_count)
	}
}

func Test_sorted_id_list_encode(t *testing.T) {

}

func Test_encode_decode_f64_map(t *testing.T) {
	fmt.Println("hi")
	//Encode_f64_map
}

func Encode_f64_map(m map[uint32]float64) []byte {
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
		for _, c0_key := range c0_key_list {
			key_id_buf := Cont_encode(uint64(c0_key))
			out_buf = append(out_buf, key_id_buf...)
		}
	}

	// c1 list
	if len(c1_key_list) != 0 {

		sort.Slice(c1_key_list, func(i, j int) bool { return c1_key_list[i] < c1_key_list[j] })
		for _, c1_key := range c1_key_list {
			key_id_buf := Cont_encode(uint64(c1_key))
			out_buf = append(out_buf, key_id_buf...)
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
	*next = uint64((bit_idx + 7) / 8) // round up to full bytes

	// c0 list
	if used_sub_maps_mask&1 == 1 {
		c0_list_len := subcontainer_lengths[cur_subcontainer_idx]
		cur_subcontainer_idx++
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
		c1_list_len := subcontainer_lengths[cur_subcontainer_idx]
		cur_subcontainer_idx++

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

func Get_bit_str(byts []byte) string {
	s := ""
	for _, b := range byts {
		s += fmt.Sprintf("%08b ", b)
	}
	return s
}

// TODO
// - rework compress f64 map to use fib_encode
//  - keys are sorted and delta encoded
//  - entry count is prefixed to key_list
// - count strings in source
// 	- order by count

// - string compression - replace numbers with placeholders

func Test_4(t *testing.T) {
	return
	orig_m := map[uint32]float64{
		10:    0.0,
		20:    1.0,
		25:    1.0,
		30:    float64(float32(1.0 / 3)),
		35:    1.0 / 3.0,
		40:    -1.0,
		50:    -0.0,
		60:    10000000000,
		700:   10000000000.1,
		80000: 123,
		80255: 123,
		80256: 123,
	}

	s := Serialize_f64_map(orig_m)

	next_idx := uint64(0)
	restored_m := Deserialize_f64_map(s, &next_idx)

	eq := reflect.DeepEqual(orig_m, restored_m)
	if !eq {
		t.Fail()
	}
}

func Demo_compress(l []uint64, prefix_1_count_to_value_bit_count []int, prefix_1_count_to_value_offset []uint64) {

	s := "{"
	for _, e := range l {
		s += strconv.FormatUint(e, 10) + " "
	}
	s += "}"

	s += " -> "

	fmt.Println(l)
	compressed_l_bytes, _ := Po_encode_u64_list(l, prefix_1_count_to_value_bit_count, prefix_1_count_to_value_offset)

	bit_str := Get_bit_str(compressed_l_bytes)
	fmt.Println(bit_str)

	restored_l := Po_decode_u64_list(compressed_l_bytes, len(l), prefix_1_count_to_value_bit_count, prefix_1_count_to_value_offset)

	eq := reflect.DeepEqual(l, restored_l)
	if !eq {
		fmt.Println("fixme")
	}

	compressed_bytes_bit_str := Get_bit_str(compressed_l_bytes)

	s += compressed_bytes_bit_str
	s += "(byte_count:" + strconv.FormatInt(int64(len(compressed_l_bytes)), 10) + ")"
	fmt.Println(s)
	//fmt.Println("restored_l:", restored_l)

}

func Get_prefix_1_count_to_value_offset(prefix_1_count_to_value_bit_count []int) []uint64 {
	current_representable_max := uint64(0)
	r := make([]uint64, len(prefix_1_count_to_value_bit_count))
	for k, v := range prefix_1_count_to_value_bit_count {
		r[k] = current_representable_max
		current_representable_max += 1 << v
	}
	return r
}

func Test_p1_compress(t *testing.T) {
	return

	prefix_1_count_to_value_bit_count := []int{
		1,  // 0 and 1 compress nicely // good enough for sparse and good for bool storage
		5,  // id deltas
		8,  // remaining 8 bit values; nice random u8
		12, // less aggressive step to encode human numbers (stuff smaller than 4000)
		16, // nice random u16
		20, // less aggressive step to encode human numbers (stuff smaller than 1 000 000)
		32, // nice random u32
		63, // most u64 (not at 7, 15, or 31, since too much cost would be passed on)
		64, // nice random u64
	}
	bit_counts := prefix_1_count_to_value_bit_count

	prefix_1_count_to_value_offset := Get_prefix_1_count_to_value_offset(prefix_1_count_to_value_bit_count)
	value_offsets := prefix_1_count_to_value_offset

	fmt.Println("===")

	fmt.Println("prefix_1_count_to_value_bit_count:")
	for k, v := range prefix_1_count_to_value_bit_count {
		fmt.Println("-", k, "->", v)
	}
	fmt.Println("prefix_1_count_to_value_offset:")
	for prefix_1_count, value_offset := range prefix_1_count_to_value_offset {
		fmt.Println("-", prefix_1_count, "->", value_offset)
	}
	fmt.Println("")

	Demo_compress([]uint64{0, 1, 21, 23, 24, 100, 106}, bit_counts, value_offsets)
	Demo_compress([]uint64{0, 0, 19, 1, 0, 75, 5}, bit_counts, value_offsets)
	Demo_compress([]uint64{1, 4, 7, 15, 16, 23, 42}, bit_counts, value_offsets)
	Demo_compress([]uint64{128}, bit_counts, value_offsets)

	fmt.Println("====")
	Demo_compress([]uint64{
		0, 0, 0, // 0
		9, 12, 12, // 4
		114, 148, // 8
		416, 5368, // 12
		15132,               // 13
		29615, 35206, 53896, // 16
		566910, // 20
	}, bit_counts, value_offsets)

	Demo_compress([]uint64{
		0, 0, 0,
		9, 3, 0,
		102, 34,
		268, 4952, 9764,
		14483, 5591, 18690,
		513014,
	}, bit_counts, value_offsets)

	//

	Demo_compress([]uint64{0, 0, 0, 0}, bit_counts, value_offsets)
	Demo_compress([]uint64{1, 1, 1, 1}, bit_counts, value_offsets)
	Demo_compress([]uint64{2, 2}, bit_counts, value_offsets)
	Demo_compress([]uint64{4, 4}, bit_counts, value_offsets)
	Demo_compress([]uint64{16, 16, 16}, bit_counts, value_offsets)
	Demo_compress([]uint64{16, 16, 16, 16}, bit_counts, value_offsets)
	Demo_compress([]uint64{16, 16, 1, 6, 20, 120, 240, 2000, 0, 0, 3, 1}, bit_counts, value_offsets)

	// Demo_compress([]uint64{2, 2, 2})

}
