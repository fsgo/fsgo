// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml_test

import (
	"fmt"

	"github.com/fsgo/fsgo/fshtml"
)

func ExampleNewUl() {
	values := []string{"1", "2", "3"}
	ul := fshtml.NewUl(values)
	got, _ := ul.HTML()
	fmt.Println(string(got))
	// Output:
	// <ul><li>1</li>
	// <li>2</li>
	// <li>3</li>
	// </ul>
}
func ExampleNewOl() {
	values := []string{"1", "2", "3"}
	ul := fshtml.NewOl(values)
	style := &fshtml.StyleAttributes{}
	_ = style.Width("180px").Height("20px").SetTo(ul.MustAttrs())

	got, _ := ul.HTML()
	fmt.Println(string(got))
	// Output:
	// <ol style="width:180px; height:20px"><li>1</li>
	// <li>2</li>
	// <li>3</li>
	// </ol>
}
