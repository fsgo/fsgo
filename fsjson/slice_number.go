// Code generated by cmd/slice_number_go.go. DO NOT EDIT.

package fsjson

import (
	"encoding/json"
)

var _ json.Unmarshaler = (*IntSlice)(nil)

// IntSlice 扩展支持 JSON 的 []int 类型
// 其实际值可以是多种格式，比如:
// 	value: ""
// 	value: "123,456"
// 	value: [123,"456",1]
// 	value: null
// 	value: 123
// 	不支持 float 类型，如 "1.2"、1.3 都会失败
type IntSlice []int

// UnmarshalJSON 实现了自定义的 json.Unmarshaler
func (ns *IntSlice) UnmarshalJSON(bf []byte) error {
	vs, err := numberSliceUnmarshalJSON[int](bf, int(0))
	if err != nil {
		return err
	}
	if len(vs) > 0 {
		*ns = vs
	}
	return nil
}

// Slice 返回 []int 的值
func (ns IntSlice) Slice() []int {
	return ns
}

// --------------------------------------------------------------------------------

var _ json.Unmarshaler = (*Int8Slice)(nil)

// Int8Slice 扩展支持 JSON 的 []int8 类型
// 其实际值可以是多种格式，比如:
// 	value: ""
// 	value: "123,456"
// 	value: [123,"456",1]
// 	value: null
// 	value: 123
// 	不支持 float 类型，如 "1.2"、1.3 都会失败
type Int8Slice []int8

// UnmarshalJSON 实现了自定义的 json.Unmarshaler
func (ns *Int8Slice) UnmarshalJSON(bf []byte) error {
	vs, err := numberSliceUnmarshalJSON[int8](bf, int8(0))
	if err != nil {
		return err
	}
	if len(vs) > 0 {
		*ns = vs
	}
	return nil
}

// Slice 返回 []int8 的值
func (ns Int8Slice) Slice() []int8 {
	return ns
}

// --------------------------------------------------------------------------------

var _ json.Unmarshaler = (*Int16Slice)(nil)

// Int16Slice 扩展支持 JSON 的 []int16 类型
// 其实际值可以是多种格式，比如:
// 	value: ""
// 	value: "123,456"
// 	value: [123,"456",1]
// 	value: null
// 	value: 123
// 	不支持 float 类型，如 "1.2"、1.3 都会失败
type Int16Slice []int16

// UnmarshalJSON 实现了自定义的 json.Unmarshaler
func (ns *Int16Slice) UnmarshalJSON(bf []byte) error {
	vs, err := numberSliceUnmarshalJSON[int16](bf, int16(0))
	if err != nil {
		return err
	}
	if len(vs) > 0 {
		*ns = vs
	}
	return nil
}

// Slice 返回 []int16 的值
func (ns Int16Slice) Slice() []int16 {
	return ns
}

// --------------------------------------------------------------------------------

var _ json.Unmarshaler = (*Int32Slice)(nil)

// Int32Slice 扩展支持 JSON 的 []int32 类型
// 其实际值可以是多种格式，比如:
// 	value: ""
// 	value: "123,456"
// 	value: [123,"456",1]
// 	value: null
// 	value: 123
// 	不支持 float 类型，如 "1.2"、1.3 都会失败
type Int32Slice []int32

// UnmarshalJSON 实现了自定义的 json.Unmarshaler
func (ns *Int32Slice) UnmarshalJSON(bf []byte) error {
	vs, err := numberSliceUnmarshalJSON[int32](bf, int32(0))
	if err != nil {
		return err
	}
	if len(vs) > 0 {
		*ns = vs
	}
	return nil
}

// Slice 返回 []int32 的值
func (ns Int32Slice) Slice() []int32 {
	return ns
}

// --------------------------------------------------------------------------------

var _ json.Unmarshaler = (*Int64Slice)(nil)

// Int64Slice 扩展支持 JSON 的 []int64 类型
// 其实际值可以是多种格式，比如:
// 	value: ""
// 	value: "123,456"
// 	value: [123,"456",1]
// 	value: null
// 	value: 123
// 	不支持 float 类型，如 "1.2"、1.3 都会失败
type Int64Slice []int64

// UnmarshalJSON 实现了自定义的 json.Unmarshaler
func (ns *Int64Slice) UnmarshalJSON(bf []byte) error {
	vs, err := numberSliceUnmarshalJSON[int64](bf, int64(0))
	if err != nil {
		return err
	}
	if len(vs) > 0 {
		*ns = vs
	}
	return nil
}

// Slice 返回 []int64 的值
func (ns Int64Slice) Slice() []int64 {
	return ns
}

// --------------------------------------------------------------------------------

var _ json.Unmarshaler = (*UintSlice)(nil)

// UintSlice 扩展支持 JSON 的 []uint 类型
// 其实际值可以是多种格式，比如:
// 	value: ""
// 	value: "123,456,-1"
// 	value: [123,"456",1,-1]
// 	value: null
// 	value: 123
// 	不支持 float 类型，如 "1.2"、1.3 都会失败
type UintSlice []uint

// UnmarshalJSON 实现了自定义的 json.Unmarshaler
func (ns *UintSlice) UnmarshalJSON(bf []byte) error {
	vs, err := numberSliceUnmarshalJSON[uint](bf, uint(0))
	if err != nil {
		return err
	}
	if len(vs) > 0 {
		*ns = vs
	}
	return nil
}

// Slice 返回 []uint 的值
func (ns UintSlice) Slice() []uint {
	return ns
}

// --------------------------------------------------------------------------------

var _ json.Unmarshaler = (*Uint8Slice)(nil)

// Uint8Slice 扩展支持 JSON 的 []uint8 类型
// 其实际值可以是多种格式，比如:
// 	value: ""
// 	value: "123,456,-1"
// 	value: [123,"456",1,-1]
// 	value: null
// 	value: 123
// 	不支持 float 类型，如 "1.2"、1.3 都会失败
type Uint8Slice []uint8

// UnmarshalJSON 实现了自定义的 json.Unmarshaler
func (ns *Uint8Slice) UnmarshalJSON(bf []byte) error {
	vs, err := numberSliceUnmarshalJSON[uint8](bf, uint8(0))
	if err != nil {
		return err
	}
	if len(vs) > 0 {
		*ns = vs
	}
	return nil
}

// Slice 返回 []uint8 的值
func (ns Uint8Slice) Slice() []uint8 {
	return ns
}

// --------------------------------------------------------------------------------

var _ json.Unmarshaler = (*Uint16Slice)(nil)

// Uint16Slice 扩展支持 JSON 的 []uint16 类型
// 其实际值可以是多种格式，比如:
// 	value: ""
// 	value: "123,456,-1"
// 	value: [123,"456",1,-1]
// 	value: null
// 	value: 123
// 	不支持 float 类型，如 "1.2"、1.3 都会失败
type Uint16Slice []uint16

// UnmarshalJSON 实现了自定义的 json.Unmarshaler
func (ns *Uint16Slice) UnmarshalJSON(bf []byte) error {
	vs, err := numberSliceUnmarshalJSON[uint16](bf, uint16(0))
	if err != nil {
		return err
	}
	if len(vs) > 0 {
		*ns = vs
	}
	return nil
}

// Slice 返回 []uint16 的值
func (ns Uint16Slice) Slice() []uint16 {
	return ns
}

// --------------------------------------------------------------------------------

var _ json.Unmarshaler = (*Uint32Slice)(nil)

// Uint32Slice 扩展支持 JSON 的 []uint32 类型
// 其实际值可以是多种格式，比如:
// 	value: ""
// 	value: "123,456,-1"
// 	value: [123,"456",1,-1]
// 	value: null
// 	value: 123
// 	不支持 float 类型，如 "1.2"、1.3 都会失败
type Uint32Slice []uint32

// UnmarshalJSON 实现了自定义的 json.Unmarshaler
func (ns *Uint32Slice) UnmarshalJSON(bf []byte) error {
	vs, err := numberSliceUnmarshalJSON[uint32](bf, uint32(0))
	if err != nil {
		return err
	}
	if len(vs) > 0 {
		*ns = vs
	}
	return nil
}

// Slice 返回 []uint32 的值
func (ns Uint32Slice) Slice() []uint32 {
	return ns
}

// --------------------------------------------------------------------------------

var _ json.Unmarshaler = (*Uint64Slice)(nil)

// Uint64Slice 扩展支持 JSON 的 []uint64 类型
// 其实际值可以是多种格式，比如:
// 	value: ""
// 	value: "123,456,-1"
// 	value: [123,"456",1,-1]
// 	value: null
// 	value: 123
// 	不支持 float 类型，如 "1.2"、1.3 都会失败
type Uint64Slice []uint64

// UnmarshalJSON 实现了自定义的 json.Unmarshaler
func (ns *Uint64Slice) UnmarshalJSON(bf []byte) error {
	vs, err := numberSliceUnmarshalJSON[uint64](bf, uint64(0))
	if err != nil {
		return err
	}
	if len(vs) > 0 {
		*ns = vs
	}
	return nil
}

// Slice 返回 []uint64 的值
func (ns Uint64Slice) Slice() []uint64 {
	return ns
}

// --------------------------------------------------------------------------------

var _ json.Unmarshaler = (*Float32Slice)(nil)

// Float32Slice 扩展支持 JSON 的 []float32 类型
// 其实际值可以是多种格式，比如:
// 	value: ""
// 	value: "123,456,-1,1.2"
// 	value: [123,"456",1,"2.1",2.3]
// 	value: null
// 	value: 123
// 	value: 123.1
type Float32Slice []float32

// UnmarshalJSON 实现了自定义的 json.Unmarshaler
func (ns *Float32Slice) UnmarshalJSON(bf []byte) error {
	vs, err := numberSliceUnmarshalJSON[float32](bf, float32(0))
	if err != nil {
		return err
	}
	if len(vs) > 0 {
		*ns = vs
	}
	return nil
}

// Slice 返回 []float32 的值
func (ns Float32Slice) Slice() []float32 {
	return ns
}

// --------------------------------------------------------------------------------

var _ json.Unmarshaler = (*Float64Slice)(nil)

// Float64Slice 扩展支持 JSON 的 []float64 类型
// 其实际值可以是多种格式，比如:
// 	value: ""
// 	value: "123,456,-1,1.2"
// 	value: [123,"456",1,"2.1",2.3]
// 	value: null
// 	value: 123
// 	value: 123.1
type Float64Slice []float64

// UnmarshalJSON 实现了自定义的 json.Unmarshaler
func (ns *Float64Slice) UnmarshalJSON(bf []byte) error {
	vs, err := numberSliceUnmarshalJSON[float64](bf, float64(0))
	if err != nil {
		return err
	}
	if len(vs) > 0 {
		*ns = vs
	}
	return nil
}

// Slice 返回 []float64 的值
func (ns Float64Slice) Slice() []float64 {
	return ns
}
