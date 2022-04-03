// Copyright(C) 2022 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/4/3

package fsjson

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObject_UnmarshalJSON(t *testing.T) {
	type class struct {
		Name string
	}

	type user struct {
		Cl *Object
	}

	t.Run("empty-as-array", func(t *testing.T) {
		vs := []string{
			`{"Cl":[]}`,
			`{"Cl":""}`,
		}
		for i, str := range vs {
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				var u *user
				err := json.Unmarshal([]byte(str), &u)
				require.NoError(t, err)
				require.NotNil(t, u.Cl)
				var c *class
				require.NoError(t, u.Cl.UnmarshalTo(&c))
				require.Nil(t, c)
			})
		}
	})

	t.Run("empty-null", func(t *testing.T) {
		var u *user
		err := json.Unmarshal([]byte(`{"Cl":null}`), &u)
		require.NoError(t, err)
		require.NotNil(t, u)
		require.Nil(t, u.Cl)
	})

	t.Run("has value", func(t *testing.T) {
		var u *user
		content := []byte(`{"Cl":{"Name":"hello"}}`)
		err := json.Unmarshal(content, &u)
		require.NoError(t, err)
		require.NotNil(t, u.Cl)
		var c *class
		require.NoError(t, u.Cl.UnmarshalTo(&c))
		wantC := &class{Name: "hello"}
		require.Equal(t, c, wantC)
	})
}

func TestObject_MarshalJSON(t *testing.T) {
	type class struct {
		Name string
	}

	type user struct {
		Cl *Object
	}

	t.Run("value nil", func(t *testing.T) {
		u := &user{}
		bf, err := json.Marshal(u)
		require.NoError(t, err)
		require.Equal(t, `{"Cl":null}`, string(bf))
	})
	t.Run("value not nil", func(t *testing.T) {
		u := &user{
			Cl: NewStruct(&class{Name: "hello"}),
		}
		bf, err := json.Marshal(u)
		require.NoError(t, err)
		require.Equal(t, `{"Cl":{"Name":"hello"}}`, string(bf))
	})
}
