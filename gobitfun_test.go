package gobitfun

import (
	"fmt"
	"testing"
	//"bytes"
	//"encoding/binary"
	//"math/bits"
)

func Test_0(t *testing.T) {
	fmt.Println("hi3")

	// buf -> int
	// buf -> int_bits
	// f64_bits := binary.LittleEndian.Uint64(f64_buf)

	// int_bits -> float
	// f64 := math.Float64frombits(f64_bits)

	// int -> buf
	// key_id_buf := make([]byte, 2)
	// binary.LittleEndian.PutUint16(key_id_buf, uint16(key_id))

	//orig := uint64(256)
	//orig_buf := make([]byte, 8)
	//binary.LittleEndian.PutUint64(orig_buf, orig)
	//_ = orig_buf

}

func Test_2(t *testing.T) {

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
