// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml

import (
	"html"
	"strings"

	"github.com/fsgo/fsgo/fstypes"
)

type (
	// Attributes HTML 的属性接口定义
	Attributes interface {
		SetAttr(key string, val AttrValue)
		Attr(key string) AttrValue
		MustAttr(key string) AttrValue
		DeleteAttr(key ...string)
		AttrKeys() []string
		Code
	}

	// AttrValue 属性的值的接口定义
	AttrValue interface {
		Code

		Value() []string
		Sep() string
		SetSep(sep string)
		Set(value ...string)
		Add(value ...string)
		Delete(value ...string)
	}
)

// NewAttributes 一个新的属性集合
func NewAttributes() Attributes {
	return &attrs{}
}

var _ Attributes = (*attrs)(nil)

type attrs struct {
	values map[string]AttrValue
	keys   fstypes.StringSlice
}

func (a *attrs) MustAttr(key string) AttrValue {
	if val := a.Attr(key); val != nil {
		return val
	}
	val := NewAttrValue(" ")
	a.SetAttr(key, val)
	return val
}

func (a *attrs) Attr(key string) AttrValue {
	if a.values == nil {
		return nil
	}
	return a.values[key]
}

func (a *attrs) DeleteAttr(keys ...string) {
	if len(a.values) == 0 {
		return
	}
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		delete(a.values, key)
		if a.keys.Has(key) {
			a.keys.Delete(key)
		}
	}
}

func (a *attrs) AttrKeys() []string {
	return a.keys
}

func (a *attrs) SetAttr(key string, val AttrValue) {
	if a.values == nil {
		a.values = map[string]AttrValue{
			key: val,
		}
	} else {
		a.values[key] = val
	}
	if !a.keys.Has(key) {
		a.keys = append(a.keys, key)
	}
}

func (a *attrs) HTML() ([]byte, error) {
	return attrsHTML(a, "=", `"`, " ")
}

func attrsHTML(attrs Attributes, kvSep string, quota string, sep string) ([]byte, error) {
	keys := attrs.AttrKeys()
	if len(keys) == 0 {
		return nil, nil
	}
	bw := newBufWriter()
	for i := 0; i < len(keys); i++ {
		attrKey := keys[i]
		bw.Write(attrKey)
		bf, err := attrs.Attr(attrKey).HTML()
		if err != nil {
			return nil, err
		}
		if len(bf) > 0 {
			bw.Write(kvSep, quota, bf, quota)
		}
		if i != len(keys)-1 {
			bw.Write(sep)
		}
	}
	return bw.HTML()
}

// NewAttrValue 创建一个属性值
// sep 是用于连接多个属性值的分隔符
func NewAttrValue(sep string) AttrValue {
	return &attrValue{
		sep: sep,
	}
}

var _ AttrValue = (*attrValue)(nil)

type attrValue struct {
	values fstypes.StringSlice // 属性值，允许多个
	sep    string              // 多个属性值的分隔符/连接符
}

func (a *attrValue) Sep() string {
	if len(a.sep) == 0 {
		return " "
	}
	return a.sep
}

func (a *attrValue) SetSep(sep string) {
	a.sep = sep
}

func (a *attrValue) Value() []string {
	return a.values
}

func (a *attrValue) Set(value ...string) {
	a.values = fstypes.StringSlice(value).Unique()
}

func (a *attrValue) Add(value ...string) {
	a.values = append(a.values, value...).Unique()
}

func (a *attrValue) Delete(value ...string) {
	a.values.Delete(value...)
}

func (a *attrValue) HTML() ([]byte, error) {
	if len(a.values) == 0 {
		return nil, nil
	}
	txt := strings.Join(a.Value(), a.Sep())
	if len(txt) > 0 {
		return []byte(html.EscapeString(txt)), nil
	}
	return nil, nil
}

const (
	attrClass = "class"
	attrID    = "id"
	attrName  = "name"
	attrStyle = "style"
)

func findOrCreateAttr(w Attributes, key string, sep string) AttrValue {
	attr := w.Attr(key)
	if attr != nil {
		return attr
	}
	attr = NewAttrValue(" ")
	w.SetAttr(key, attr)
	return attr
}

// SetClass 设置 class 属性
func SetClass(w Attributes, class ...string) {
	findOrCreateAttr(w, attrClass, " ").Set(class...)
}

// AddClass 添加 class 属性
func AddClass(w Attributes, class ...string) {
	findOrCreateAttr(w, attrClass, " ").Add(class...)
}

// DeleteClass 删除 class 属性
func DeleteClass(w Attributes, class ...string) {
	attr := w.Attr(attrClass)
	if attr == nil {
		return
	}
	attr.Delete(class...)
}

// SetID 设置元素的 id
func SetID(w Attributes, id string) {
	findOrCreateAttr(w, attrID, " ").Set(id)
}

// SetName 涉足元素的 name
func SetName(w Attributes, name string) {
	findOrCreateAttr(w, attrName, " ").Set(name)
}

// AttrStyle style 属性
type AttrStyle struct {
	attrs Attributes
}

// Set 设置 key 的属性值为 value
func (s *AttrStyle) Set(key, value string) *AttrStyle {
	if s.attrs == nil {
		s.attrs = NewAttributes()
	}
	attr := s.attrs.Attr(key)
	if attr == nil {
		attr = NewAttrValue(";")
		s.attrs.SetAttr(key, attr)
	}
	attr.Set(value)
	return s
}

// Get 获取属性值
func (s *AttrStyle) Get(key string) string {
	if s.attrs == nil {
		return ""
	}
	attr := s.attrs.Attr(key)
	if attr == nil {
		return ""
	}
	vs := attr.Value()
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

// Width 设置宽度
func (s *AttrStyle) Width(w string) *AttrStyle {
	return s.Set("width", w)
}

// MinWidth 设置最小宽度
func (s *AttrStyle) MinWidth(w string) *AttrStyle {
	return s.Set("min-width", w)
}

// MaxWidth 设置最大新宽度
func (s *AttrStyle) MaxWidth(w string) *AttrStyle {
	return s.Set("max-width", w)
}

// Height 设置高度
func (s *AttrStyle) Height(h string) *AttrStyle {
	return s.Set("height", h)
}

// MinHeight 设置最小高度
func (s *AttrStyle) MinHeight(h string) *AttrStyle {
	return s.Set("min-height", h)
}

// MaxHeight 设置最大高度
func (s *AttrStyle) MaxHeight(h string) *AttrStyle {
	return s.Set("max-height", h)
}

// Color 设置前景/字体颜色
func (s *AttrStyle) Color(color string) *AttrStyle {
	return s.Set("color", color)
}

// BackgroundColor 设置背景演示
func (s *AttrStyle) BackgroundColor(color string) *AttrStyle {
	return s.Set("background-color", color)
}

// TextAlign 设置内容对齐方式
func (s *AttrStyle) TextAlign(align string) *AttrStyle {
	return s.Set("text-align", align)
}

// Margin 设置外边距
func (s *AttrStyle) Margin(val string) *AttrStyle {
	return s.Set("margin", val)
}

// Padding 设置内边距
func (s *AttrStyle) Padding(val string) *AttrStyle {
	return s.Set("padding", val)
}

// Font 设置字体
func (s *AttrStyle) Font(val string) *AttrStyle {
	return s.Set("font", val)
}

// FontSize 设置字体大小
func (s *AttrStyle) FontSize(val string) *AttrStyle {
	return s.Set("font-size", val)
}

// FontWeight 设置字体粗细
func (s *AttrStyle) FontWeight(val string) *AttrStyle {
	return s.Set("font-weight", val)
}

// FontFamily 设置字体系列（字体族）
func (s *AttrStyle) FontFamily(val string) *AttrStyle {
	return s.Set("font-family", val)
}

// Border 设置边框属性
func (s *AttrStyle) Border(val string) *AttrStyle {
	return s.Set("border", val)
}

// HTML 实现 Code 接口
func (s *AttrStyle) HTML() ([]byte, error) {
	if s.attrs == nil {
		return nil, nil
	}
	return attrsHTML(s.attrs, ":", "", "; ")
}

// SetTo 将样式信息设置到指定的属性集合
func (s *AttrStyle) SetTo(a Attributes) error {
	code, err := s.HTML()
	if err != nil {
		return err
	}
	if len(code) == 0 {
		return nil
	}
	a.MustAttr(attrStyle).Set(string(code))
	return nil
}
