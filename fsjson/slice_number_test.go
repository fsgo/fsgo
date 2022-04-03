// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/4/3

package fsjson

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestInt64Slice_UnmarshalJSON(t *testing.T) {
	type user struct {
		IDS Int64Slice
	}
	tests := []struct {
		name    string
		txt     string
		wantIDS Int64Slice
		wantErr bool
	}{
		{
			name:    "case 1",
			txt:     `{"IDS":[1,2,-1]}`,
			wantErr: false,
			wantIDS: []int64{1, 2, -1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &user{}
			if err := json.Unmarshal([]byte(tt.txt), &u); (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.wantIDS, u.IDS) {
				t.Fatalf("wantIDS=%#v got=%#v", tt.wantIDS, u.IDS)
			}
		})
	}
}
