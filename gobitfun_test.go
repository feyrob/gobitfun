package gobitfun

import (
	"fmt"
	//"math"
	//	"math/bits"
	"strconv"
	"testing"

	//"encoding/binary"

	"os"
	"reflect"
	//	"sort"
)

// TODO dict_list - delta dicts
// a delta dict can copy a previous dict via other_dict_idx
// a delta dict can not remove or add any dict fields
// a delta dict can overwrite any number of fields with new values

// TODO
type Telemetry_event struct {
	epoch_millis int64
	// epoch starts at 1970
	// int64 supports up to 2^63 future milliseconds
	// / 1000 -> seconds
	// / 60 -> minutes
	// / 60 -> hours
	// / 24 -> days
	// / 365 -> ~years
	// = (2^64)/1000/60/60/24/365 = 584942417.355 years

	f64_fields map[string]float64
	str_fields map[string]string
}

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

func Test_create_chart(t *testing.T) {
	f, _ := os.Create(`C:\Users\feyro\test.xml`)
	defer f.Close()
	f.Write([]byte("<blah>"))
	for i := 1; i < 10000; i = int((float64(i) * 1.1) + 1.0) {
		s := fmt.Sprintf("<blupp foo=\"%d\" bar=\"%d\" />", i, i*i)
		f.Write([]byte(s))
	}

	f.Write([]byte("</blah>"))

}

func Test_sorted_id_list_encode(t *testing.T) {

}

func Test_encode_decode_f64_map(t *testing.T) {
	fmt.Println("hi")
	//Encode_f64_map

	orig := map[uint32]float64{
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
		90123: 0.0,

		100000: 63,
		100001: 65,
		100002: 100002,
		100003: 100003,
		100004: 100004,
		100005: 100005,
		100006: 100006,
		100007: 100007,
	}

	encoded := Encode_f64_map(orig)

	bit_str := Get_bit_str(encoded)
	fmt.Println("encoded:", bit_str)

	next := uint64(0)
	restored := Decode_f64_map(encoded, &next)

	eq := reflect.DeepEqual(orig, restored)
	if eq {
		fmt.Println("yay")
	} else {
		fmt.Println("meh")
	}
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

	s := Encode_f64_map(orig_m)

	next_idx := uint64(0)
	restored_m := Decode_f64_map(s, &next_idx)

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

	compressed_l_bytes := []byte{}
	Next_bit_offset := uint64(0)
	for _, v := range l {
		Po_encode_u64(&compressed_l_bytes, prefix_1_count_to_value_bit_count, prefix_1_count_to_value_offset, v, &Next_bit_offset)
	}

	bit_len := Bit_len(compressed_l_bytes, Next_bit_offset)
	fmt.Println("bit_len:", bit_len)

	//compressed_l_bytes, _ := Po_encode_u64_list(l, prefix_1_count_to_value_bit_count, prefix_1_count_to_value_offset)

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

func Test_delta_fib_encode_sorted_id_list(t *testing.T) {
	l := []uint64{0, 2, 10, 11} // sorted and each entry is unique

	delta_l := Sorted_id_list_to_delta_list(l)
	restored_l := Delta_list_to_sorted_id_list(delta_l)

	eq := reflect.DeepEqual(l, restored_l)
	if eq {
		fmt.Println("yay")
	} else {
		fmt.Println("meh")
	}

	fmt.Println("delta_l:", delta_l)
	return

}

func Test_p1_compress(t *testing.T) {

	bit_counts := Po__general_purpose__prefix_1_count_to_value_bit_count

	value_offsets := Po__general_purpose__prefix_1_count_to_value_offset

	fmt.Println("===")

	fmt.Println("prefix_1_count_to_value_bit_count:")
	for k, v := range bit_counts {
		fmt.Println("-", k, "->", v)
	}
	fmt.Println("prefix_1_count_to_value_offset:")
	for prefix_1_count, value_offset := range Po__general_purpose__prefix_1_count_to_value_offset {
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
