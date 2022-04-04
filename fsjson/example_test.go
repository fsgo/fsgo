// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/4/3

package fsjson_test

import (
	"encoding/json"
	"fmt"

	"github.com/fsgo/fsgo/fsjson"
)

func ExampleStringSlice_UnmarshalJSON() {
	type user struct {
		Alias fsjson.StringSlice
	}

	txtList := []string{
		`{}`,
		`{"Alias":""}`,
		`{"Alias":null}`,
		`{"Alias":123}`,
		`{"Alias":2.3}`,
		`{"Alias":"abc"}`,
		`{"Alias":"abc,def,1"}`,
		`{"Alias":["abc","def",123,0.1]}`,
		`{"Alias":err}`, // not support
	}
	for i := 0; i < len(txtList); i++ {
		var u *user
		err := json.Unmarshal([]byte(txtList[i]), &u)
		fmt.Printf("err=%v user=%#v\n", err != nil, u)
	}
	// Output:
	// err=false user=&fsjson_test.user{Alias:fsjson.StringSlice(nil)}
	// err=false user=&fsjson_test.user{Alias:fsjson.StringSlice(nil)}
	// err=false user=&fsjson_test.user{Alias:fsjson.StringSlice(nil)}
	// err=false user=&fsjson_test.user{Alias:fsjson.StringSlice{"123"}}
	// err=false user=&fsjson_test.user{Alias:fsjson.StringSlice{"2.3"}}
	// err=false user=&fsjson_test.user{Alias:fsjson.StringSlice{"abc"}}
	// err=false user=&fsjson_test.user{Alias:fsjson.StringSlice{"abc", "def", "1"}}
	// err=false user=&fsjson_test.user{Alias:fsjson.StringSlice{"abc", "def", "123", "0.1"}}
	// err=true user=(*fsjson_test.user)(nil)
}

func ExampleInt8Slice_UnmarshalJSON() {
	type user struct {
		IDS fsjson.Int8Slice
	}

	txtList := []string{
		`{}`,
		`{"IDS":""}`,
		`{"IDS":null}`,
		`{"IDS":123}`,
		`{"IDS":["123",456,0]}`, // not support, 456 not int8
		`{"IDS":[1,2,-1]}`,
		`{"IDS":-2}`,
		`{"IDS":2.3}`,         // not support
		`{"IDS":"abc"}`,       // not support
		`{"IDS":"abc,def,1"}`, // not support
		`{"IDS":err}`,         // not support
		`{"IDS":1.0}`,         // not support
		`{"IDS":"1.0"}`,       // not support
	}

	printUser := func(u *user) string {
		if u == nil {
			return "nil"
		}
		if u.IDS == nil {
			return "&{IDS:nil}"
		}
		return fmt.Sprintf("%+v", u)
	}

	for i := 0; i < len(txtList); i++ {
		txt := txtList[i]
		var u *user
		err := json.Unmarshal([]byte(txt), &u)
		fmt.Printf("%-30s -> err=%v user=%s\n", txt, err != nil, printUser(u))
	}
	// Output:
	// {}                             -> err=false user=&{IDS:nil}
	// {"IDS":""}                     -> err=false user=&{IDS:nil}
	// {"IDS":null}                   -> err=false user=&{IDS:nil}
	// {"IDS":123}                    -> err=false user=&{IDS:[123]}
	// {"IDS":["123",456,0]}          -> err=true user=&{IDS:nil}
	// {"IDS":[1,2,-1]}               -> err=false user=&{IDS:[1 2 -1]}
	// {"IDS":-2}                     -> err=false user=&{IDS:[-2]}
	// {"IDS":2.3}                    -> err=true user=&{IDS:nil}
	// {"IDS":"abc"}                  -> err=true user=&{IDS:nil}
	// {"IDS":"abc,def,1"}            -> err=true user=&{IDS:nil}
	// {"IDS":err}                    -> err=true user=nil
	// {"IDS":1.0}                    -> err=true user=&{IDS:nil}
	// {"IDS":"1.0"}                  -> err=true user=&{IDS:nil}
}

func ExampleInt64Slice_UnmarshalJSON() {
	type user struct {
		IDS fsjson.Int64Slice
	}

	txtList := []string{
		`{}`,
		`{"IDS":""}`,
		`{"IDS":null}`,
		`{"IDS":123}`,
		`{"IDS":["123",456,0]}`,
		`{"IDS":[1,2,-1]}`,
		`{"IDS":-2}`,
		`{"IDS":2.3}`,         // not support
		`{"IDS":"abc"}`,       // not support
		`{"IDS":"abc,def,1"}`, // not support
		`{"IDS":err}`,         // not support
		`{"IDS":1.0}`,         // not support
		`{"IDS":"1.0"}`,       // not support
	}

	printUser := func(u *user) string {
		if u == nil {
			return "nil"
		}
		if u.IDS == nil {
			return "&{IDS:nil}"
		}
		return fmt.Sprintf("%+v", u)
	}

	for i := 0; i < len(txtList); i++ {
		txt := txtList[i]
		var u *user
		err := json.Unmarshal([]byte(txt), &u)
		fmt.Printf("%-30s -> err=%v user=%s\n", txt, err != nil, printUser(u))
	}
	// Output:
	// {}                             -> err=false user=&{IDS:nil}
	// {"IDS":""}                     -> err=false user=&{IDS:nil}
	// {"IDS":null}                   -> err=false user=&{IDS:nil}
	// {"IDS":123}                    -> err=false user=&{IDS:[123]}
	// {"IDS":["123",456,0]}          -> err=false user=&{IDS:[123 456 0]}
	// {"IDS":[1,2,-1]}               -> err=false user=&{IDS:[1 2 -1]}
	// {"IDS":-2}                     -> err=false user=&{IDS:[-2]}
	// {"IDS":2.3}                    -> err=true user=&{IDS:nil}
	// {"IDS":"abc"}                  -> err=true user=&{IDS:nil}
	// {"IDS":"abc,def,1"}            -> err=true user=&{IDS:nil}
	// {"IDS":err}                    -> err=true user=nil
	// {"IDS":1.0}                    -> err=true user=&{IDS:nil}
	// {"IDS":"1.0"}                  -> err=true user=&{IDS:nil}
}

func ExampleUint64Slice_UnmarshalJSON() {
	type user struct {
		IDS fsjson.Uint64Slice
	}

	txtList := []string{
		`{}`,
		`{"IDS":""}`,
		`{"IDS":null}`,
		`{"IDS":123}`,
		`{"IDS":["123",456,0]}`,
		`{"IDS":[1,2]}`,
		`{"IDS":2.3}`,         // not support
		`{"IDS":-1}`,          // not support
		`{"IDS":"abc"}`,       // not support
		`{"IDS":"abc,def,1"}`, // not support
		`{"IDS":err}`,         // not support
		`{"IDS":1.0}`,         // not support
		`{"IDS":"1.0"}`,       // not support
	}
	printUser := func(u *user) string {
		if u == nil {
			return "nil"
		}
		if u.IDS == nil {
			return "&{IDS:nil}"
		}
		return fmt.Sprintf("%+v", u)
	}

	for i := 0; i < len(txtList); i++ {
		txt := txtList[i]
		var u *user
		err := json.Unmarshal([]byte(txt), &u)
		fmt.Printf("%-30s -> err=%v user=%s\n", txt, err != nil, printUser(u))
	}
	// Output:
	// {}                             -> err=false user=&{IDS:nil}
	// {"IDS":""}                     -> err=false user=&{IDS:nil}
	// {"IDS":null}                   -> err=false user=&{IDS:nil}
	// {"IDS":123}                    -> err=false user=&{IDS:[123]}
	// {"IDS":["123",456,0]}          -> err=false user=&{IDS:[123 456 0]}
	// {"IDS":[1,2]}                  -> err=false user=&{IDS:[1 2]}
	// {"IDS":2.3}                    -> err=true user=&{IDS:nil}
	// {"IDS":-1}                     -> err=true user=&{IDS:nil}
	// {"IDS":"abc"}                  -> err=true user=&{IDS:nil}
	// {"IDS":"abc,def,1"}            -> err=true user=&{IDS:nil}
	// {"IDS":err}                    -> err=true user=nil
	// {"IDS":1.0}                    -> err=true user=&{IDS:nil}
	// {"IDS":"1.0"}                  -> err=true user=&{IDS:nil}
}

func ExampleFloat64Slice_UnmarshalJSON() {
	type user struct {
		IDS fsjson.Float64Slice
	}

	txtList := []string{
		`{}`,
		`{"IDS":""}`,
		`{"IDS":null}`,
		`{"IDS":123}`,
		`{"IDS":["123", 456,0,2.1," 3.2"]}`,
		`{"IDS":[1,2]}`,
		`{"IDS":2.3}`,
		`{"IDS":-1}`,
		`{"IDS":"abc"}`,       // not support
		`{"IDS":"abc,def,1"}`, // not support
		`{"IDS":err}`,         // not support
		`{"IDS":1.0}`,
		`{"IDS":"1.0"}`,
	}
	printUser := func(u *user) string {
		if u == nil {
			return "nil"
		}
		if u.IDS == nil {
			return "&{IDS:nil}"
		}
		return fmt.Sprintf("%+v", u)
	}

	for i := 0; i < len(txtList); i++ {
		txt := txtList[i]
		var u *user
		err := json.Unmarshal([]byte(txt), &u)
		fmt.Printf("%-30s -> err=%v user=%s\n", txt, err != nil, printUser(u))
	}
	// Output:
	// {}                             -> err=false user=&{IDS:nil}
	// {"IDS":""}                     -> err=false user=&{IDS:nil}
	// {"IDS":null}                   -> err=false user=&{IDS:nil}
	// {"IDS":123}                    -> err=false user=&{IDS:[123]}
	// {"IDS":["123", 456,0,2.1," 3.2"]} -> err=false user=&{IDS:[123 456 0 2.1 3.2]}
	// {"IDS":[1,2]}                  -> err=false user=&{IDS:[1 2]}
	// {"IDS":2.3}                    -> err=false user=&{IDS:[2.3]}
	// {"IDS":-1}                     -> err=false user=&{IDS:[-1]}
	// {"IDS":"abc"}                  -> err=true user=&{IDS:nil}
	// {"IDS":"abc,def,1"}            -> err=true user=&{IDS:nil}
	// {"IDS":err}                    -> err=true user=nil
	// {"IDS":1.0}                    -> err=false user=&{IDS:[1]}
	// {"IDS":"1.0"}                  -> err=false user=&{IDS:[1]}
}

func ExampleObject_UnmarshalJSON() {
	type class struct {
		Name string
	}

	type user struct {
		Cl *fsjson.Object
	}

	txtList := []string{
		`{"Cl":[]}`,
		`{"Cl":null}`,
		`{"Cl":""}`,
		`{"Cl":{"Name":"hello"}}`,
		`{"Cl":"abc"}`, // not support
	}

	for i := 0; i < len(txtList); i++ {
		txt := txtList[i]
		c := &class{}
		u := &user{
			Cl: fsjson.NewObject(c),
		}
		err := json.Unmarshal([]byte(txt), &u)
		fmt.Printf("%-30s -> err=%v class=%+v\n", txt, err != nil, c)
	}
	// Output:
	// {"Cl":[]}                      -> err=false class=&{Name:}
	// {"Cl":null}                    -> err=false class=&{Name:}
	// {"Cl":""}                      -> err=false class=&{Name:}
	// {"Cl":{"Name":"hello"}}        -> err=false class=&{Name:hello}
	// {"Cl":"abc"}                   -> err=true class=&{Name:}
}
