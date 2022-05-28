// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml

import (
	"html"
	"strings"

	"github.com/fsgo/fsgo/fstypes"
)

// Attributes 多个属性
type Attributes struct {
	// Sep 多个属性间的连接符，当为空时，使用默认值 " " (一个空格)
	Sep string

	// KVSep key 和 value 之间的连接符，当为空时，使用默认值 =
	KVSep string

	// Quote 属性值的引号，为空时，使用默认值 "
	Quote string

	attrs map[string]*Attribute
	keys  fstypes.StringSlice
}

// GetSep  多个属性间的连接符，当为空时，返回默认值 " " (一个空格)
func (a *Attributes) GetSep() string {
	if len(a.Sep) == 0 {
		return " "
	}
	return a.Sep
}

// GetKVSep key 和 value 之间的连接符，当为空时， 返回默认值 =
func (a *Attributes) GetKVSep() string {
	if len(a.KVSep) == 0 {
		return "="
	}
	return a.KVSep
}

// GetQuote 属性值的引号，为空时，使用默认值 "
func (a *Attributes) GetQuote() string {
	if len(a.Quote) == 0 {
		return `"`
	}
	return a.Quote
}

// Find 返回一个指定的属性，若不存在，返回 nil
func (a *Attributes) Find(key string) *Attribute {
	if len(a.attrs) == 0 {
		return nil
	}
	return a.attrs[key]
}

// FindOrCreate 返回一个指定的属性，若不存在，返回 nil
func (a *Attributes) FindOrCreate(key string) *Attribute {
	if val := a.Find(key); val != nil {
		return val
	}
	attr := &Attribute{
		Key: key,
	}
	a.Set(attr)
	return attr
}

// Delete 删除指定 key 的属性
func (a *Attributes) Delete(keys ...string) {
	if len(a.attrs) == 0 {
		return
	}
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		delete(a.attrs, key)
		a.keys.Delete(key)
	}
}

// Keys 返回所有属性的 key
func (a *Attributes) Keys() []string {
	return a.keys
}

// Set 设置属性值
func (a *Attributes) Set(attr ...*Attribute) {
	if a.attrs == nil {
		a.attrs = make(map[string]*Attribute, len(attr))
	}
	for _, item := range attr {
		if _, has := a.attrs[item.Key]; !has {
			a.keys = append(a.keys, item.Key)
		}
		a.attrs[item.Key] = item
	}
}

// HTML 转换为 HTML
func (a *Attributes) HTML() ([]byte, error) {
	if a == nil {
		return nil, nil
	}
	return attrsHTML(a, a.GetKVSep(), a.GetQuote(), a.GetSep())
}

func attrsHTML(attrs *Attributes, kvSep string, quote string, sep string) ([]byte, error) {
	keys := attrs.Keys()
	if len(keys) == 0 {
		return nil, nil
	}
	bw := newBufWriter()
	for i := 0; i < len(keys); i++ {
		attrKey := keys[i]
		bw.Write(attrKey)
		bf, err := attrs.Find(attrKey).HTML()
		if err != nil {
			return nil, err
		}
		if len(bf) > 0 {
			bw.Write(kvSep, quote, bf, quote)
		}
		if i != len(keys)-1 {
			bw.Write(sep)
		}
	}
	return bw.HTML()
}

// Attribute  一个属性
type Attribute struct {
	// Key 属性的名字
	Key string
	// Values 属性值，可以有多个
	Values fstypes.StringSlice

	// Sep 多个属性值的连接符
	Sep string
}

// GetSep 多个属性值的连接符，当为空时，返回 " "(一个空格)
func (a *Attribute) GetSep() string {
	if len(a.Sep) == 0 {
		return " "
	}
	return a.Sep
}

// Set 设置属性值
func (a *Attribute) Set(value ...string) {
	a.Values = value
}

// First 返回首个属性值
func (a *Attribute) First() string {
	if len(a.Values) == 0 {
		return ""
	}
	return a.Values[0]
}

// Add 添加新的属性值
func (a *Attribute) Add(value ...string) {
	a.Values = append(a.Values, value...).Unique()
}

// Delete 删除属性值
func (a *Attribute) Delete(value ...string) {
	a.Values.Delete(value...)
}

// HTML 转换为 HTML
func (a *Attribute) HTML() ([]byte, error) {
	if len(a.Values) == 0 {
		return nil, nil
	}
	txt := strings.Join(a.Values, a.GetSep())
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

func findOrCreateAttr(w *Attributes, key string, sep string) *Attribute {
	attr := w.Find(key)
	if attr != nil {
		return attr
	}
	attr = &Attribute{
		Key: key,
		Sep: sep,
	}
	w.Set(attr)
	return attr
}

// SetClass 设置 class 属性
func SetClass(w *Attributes, class ...string) {
	findOrCreateAttr(w, attrClass, " ").Set(class...)
}

// AddClass 添加 class 属性
func AddClass(w *Attributes, class ...string) {
	findOrCreateAttr(w, attrClass, " ").Add(class...)
}

// DeleteClass 删除 class 属性
func DeleteClass(w *Attributes, class ...string) {
	attr := w.Find(attrClass)
	if attr == nil {
		return
	}
	attr.Delete(class...)
}

// SetID 设置元素的 id
func SetID(w *Attributes, id string) {
	findOrCreateAttr(w, attrID, " ").Set(id)
}

// SetName 涉足元素的 name
func SetName(w *Attributes, name string) {
	findOrCreateAttr(w, attrName, " ").Set(name)
}

// StyleAttributes style 属性
type StyleAttributes struct {
	attrs *Attributes
}

// Set 设置 key 的属性值为 value
func (s StyleAttributes) Set(key, value string) StyleAttributes {
	if s.attrs == nil {
		s.attrs = &Attributes{}
	}
	attr := s.attrs.Find(key)
	if attr == nil {
		attr = &Attribute{
			Key:    key,
			Values: []string{value},
		}
		s.attrs.Set(attr)
	}
	attr.Set(value)
	return s
}

// Get 获取属性值
func (s StyleAttributes) Get(key string) string {
	if s.attrs == nil {
		return ""
	}
	attr := s.attrs.Find(key)
	if attr == nil {
		return ""
	}
	if len(attr.Values) == 0 {
		return ""
	}
	return attr.Values[0]
}

// Width 设置宽度
func (s StyleAttributes) Width(w string) StyleAttributes {
	return s.Set("width", w)
}

// MinWidth 设置最小宽度
func (s StyleAttributes) MinWidth(w string) StyleAttributes {
	return s.Set("min-width", w)
}

// MaxWidth 设置最大新宽度
func (s StyleAttributes) MaxWidth(w string) StyleAttributes {
	return s.Set("max-width", w)
}

// Height 设置高度
func (s StyleAttributes) Height(h string) StyleAttributes {
	return s.Set("height", h)
}

// MinHeight 设置最小高度
func (s StyleAttributes) MinHeight(h string) StyleAttributes {
	return s.Set("min-height", h)
}

// MaxHeight 设置最大高度
func (s StyleAttributes) MaxHeight(h string) StyleAttributes {
	return s.Set("max-height", h)
}

// Color 设置前景/字体颜色
func (s StyleAttributes) Color(color string) StyleAttributes {
	return s.Set("color", color)
}

// BackgroundColor 设置背景演示
func (s StyleAttributes) BackgroundColor(color string) StyleAttributes {
	return s.Set("background-color", color)
}

// TextAlign 设置内容对齐方式
func (s StyleAttributes) TextAlign(align string) StyleAttributes {
	return s.Set("text-align", align)
}

// Margin 设置外边距
func (s StyleAttributes) Margin(val string) StyleAttributes {
	return s.Set("margin", val)
}

// Padding 设置内边距
func (s StyleAttributes) Padding(val string) StyleAttributes {
	return s.Set("padding", val)
}

// Font 设置字体
func (s StyleAttributes) Font(val string) StyleAttributes {
	return s.Set("font", val)
}

// FontSize 设置字体大小
func (s StyleAttributes) FontSize(val string) StyleAttributes {
	return s.Set("font-size", val)
}

// FontWeight 设置字体粗细
func (s StyleAttributes) FontWeight(val string) StyleAttributes {
	return s.Set("font-weight", val)
}

// FontFamily 设置字体系列（字体族）
func (s StyleAttributes) FontFamily(val string) StyleAttributes {
	return s.Set("font-family", val)
}

// Border 设置边框属性
func (s StyleAttributes) Border(val string) StyleAttributes {
	return s.Set("border", val)
}

// HTML 实现 Element 接口
func (s StyleAttributes) HTML() ([]byte, error) {
	if s.attrs == nil {
		return nil, nil
	}
	return attrsHTML(s.attrs, ":", "", "; ")
}

// SetTo 将样式信息设置到指定的属性集合
func (s StyleAttributes) SetTo(a *Attributes) error {
	code, err := s.HTML()
	if err != nil {
		return err
	}
	if len(code) == 0 {
		return nil
	}
	a.FindOrCreate(attrStyle).Set(string(code))
	return nil
}
