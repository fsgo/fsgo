// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

package fshtml

import (
	"strings"

	"github.com/fsgo/fsgo/fstypes"
)

const (
	// onlyKey 属性只需要 key，不需要 value
	onlyKey = ":only-key"
)

// Attrs 多个属性
type Attrs struct {
	// Sep 多个属性间的连接符，当为空时，使用默认值 " " (一个空格)
	Sep string

	// KVSep key 和 value 之间的连接符，当为空时，使用默认值 =
	KVSep string

	// Quote 属性值的引号，为空时，使用默认值 "
	Quote string

	attrs map[string]*Attr
	keys  fstypes.StringSlice
}

// GetSep  多个属性间的连接符，当为空时，返回默认值 " " (一个空格)
func (a *Attrs) GetSep() string {
	if len(a.Sep) == 0 {
		return " "
	}
	return a.Sep
}

// GetKVSep key 和 value 之间的连接符，当为空时， 返回默认值 =
func (a *Attrs) GetKVSep() string {
	if len(a.KVSep) == 0 {
		return "="
	}
	return a.KVSep
}

// GetQuote 属性值的引号，为空时，使用默认值 "
func (a *Attrs) GetQuote() string {
	if len(a.Quote) == 0 {
		return `"`
	}
	return a.Quote
}

// Attr 返回一个指定的属性，若不存在，返回 nil
func (a *Attrs) Attr(key string) *Attr {
	if len(a.attrs) == 0 {
		return nil
	}
	return a.attrs[key]
}

// MustAttr 返回一个指定的属性，若不存在，返回 nil
func (a *Attrs) MustAttr(key string) *Attr {
	if val := a.Attr(key); val != nil {
		return val
	}
	attr := &Attr{
		Key: key,
	}
	a.Set(attr)
	return attr
}

// Delete 删除指定 key 的属性
func (a *Attrs) Delete(keys ...string) {
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
func (a *Attrs) Keys() []string {
	return a.keys
}

// Set 设置属性值
func (a *Attrs) Set(attr ...*Attr) {
	if a.attrs == nil {
		a.attrs = make(map[string]*Attr, len(attr))
	}
	for _, item := range attr {
		if _, has := a.attrs[item.Key]; !has {
			a.keys = append(a.keys, item.Key)
		}
		a.attrs[item.Key] = item
	}
}

// HTML 转换为 HTML
func (a *Attrs) HTML() ([]byte, error) {
	if a == nil {
		return nil, nil
	}
	return attrsHTML(a, a.GetKVSep(), a.GetQuote(), a.GetSep())
}

func attrsHTML(attrs *Attrs, kvSep string, quote string, sep string) ([]byte, error) {
	keys := attrs.Keys()
	if len(keys) == 0 {
		return nil, nil
	}
	bw := newBufWriter()
	for i := 0; i < len(keys); i++ {
		attrKey := keys[i]
		vs := attrs.Attr(attrKey).Values
		if len(vs) == 0 {
			continue
		}
		bw.Write(attrKey)
		if vs[0] != onlyKey {
			bw.Write(kvSep, quote, strings.Join(vs, " "), quote)
		}
		if i != len(keys)-1 {
			bw.Write(sep)
		}
	}
	return bw.HTML()
}

// AttrsMapper 具有 AttrsMapper 方法
type AttrsMapper interface {
	MustAttrs() *Attrs
	FindAttrs() *Attrs
}

var _ AttrsMapper = (*WithAttrs)(nil)

// WithAttrs 具有 attrs 属性
type WithAttrs struct {
	Attrs *Attrs
}

// FindAttrs 返回当前的 Attrs
func (w *WithAttrs) FindAttrs() *Attrs {
	return w.Attrs
}

// MustAttrs 若 attrs 不存在，则创建 并返回 attrs
func (w *WithAttrs) MustAttrs() *Attrs {
	if w.Attrs == nil {
		w.Attrs = &Attrs{}
	}
	return w.Attrs
}

// DeleteAttr 删除指定的属性值
func DeleteAttr(w AttrsMapper, key string, values ...string) {
	as := w.FindAttrs()
	if as == nil {
		return
	}
	attr := as.Attr(key)
	if attr == nil {
		return
	}
	attr.Delete(values...)
}

// Attr  一个属性
type Attr struct {
	// Key 属性的名字
	Key string

	// Values 属性值，可以有多个
	Values fstypes.StringSlice
}

// Set 设置属性值
func (a *Attr) Set(value ...string) {
	a.Values = value
}

// First 返回首个属性值
func (a *Attr) First() string {
	if len(a.Values) == 0 {
		return ""
	}
	return a.Values[0]
}

// Add 添加新的属性值
func (a *Attr) Add(value ...string) {
	a.Values = append(a.Values, value...).Unique()
}

// Delete 删除属性值
func (a *Attr) Delete(value ...string) {
	a.Values.Delete(value...)
}

func findOrCreateAttr(w AttrsMapper, key string, sep string) *Attr {
	as := w.MustAttrs()
	attr := as.Attr(key)
	if attr != nil {
		return attr
	}
	attr = &Attr{
		Key: key,
	}
	as.Set(attr)
	return attr
}

// SetAttr 设置属性值
func SetAttr(w AttrsMapper, key string, value ...string) {
	findOrCreateAttr(w, key, " ").Set(value...)
}

// SetAttrNoValue 设置只有 key，不需要 value 的属性
func SetAttrNoValue(w AttrsMapper, key string) {
	findOrCreateAttr(w, key, " ").Set(onlyKey)
}

// SetAsync 设置 async  属性
func SetAsync(w AttrsMapper) {
	SetAttrNoValue(w, "async")
}

// SetClass 设置 class 属性
func SetClass(w AttrsMapper, class ...string) {
	findOrCreateAttr(w, "class", " ").Set(class...)
}

// AddClass 添加 class 属性
func AddClass(w AttrsMapper, class ...string) {
	findOrCreateAttr(w, "class", " ").Add(class...)
}

// DeleteClass 删除 class 属性
func DeleteClass(w AttrsMapper, class ...string) {
	DeleteAttr(w, "class", class...)
}

// SetID 设置元素的 id
func SetID(w AttrsMapper, id string) {
	findOrCreateAttr(w, "id", " ").Set(id)
}

// SetName 设置元素的 name
func SetName(w AttrsMapper, name string) {
	findOrCreateAttr(w, "name", " ").Set(name)
}

// SetWidth 设置元素的 width
func SetWidth(w AttrsMapper, width string) {
	findOrCreateAttr(w, "width", " ").Set(width)
}

// SetHeight 设置元素的 height
func SetHeight(w AttrsMapper, height string) {
	findOrCreateAttr(w, "height", " ").Set(height)
}

// SetLang 设置元素的 lang 属性
// 	如 en-US、zh-CN
func SetLang(w AttrsMapper, lang string) {
	findOrCreateAttr(w, "lang", " ").Set(lang)
}

// SetTitle 设置 title 属性
func SetTitle(w AttrsMapper, title string) {
	findOrCreateAttr(w, "title", " ").Set(title)
}

// SetSrc 设置 src 属性
func SetSrc(w AttrsMapper, src string) {
	findOrCreateAttr(w, "src", " ").Set(src)
}

// SetTarget 设置 target 属性
func SetTarget(w AttrsMapper, target string) {
	findOrCreateAttr(w, "target", " ").Set(target)
}

// SetType 设置 type 属性
func SetType(w AttrsMapper, tp string) {
	findOrCreateAttr(w, "type", " ").Set(tp)
}

// StyleAttr style 属性
type StyleAttr struct {
	WithAttrs
}

// set 设置 key 的属性值为 value
func (s *StyleAttr) set(key, value string) *StyleAttr {
	attrs := s.MustAttrs()
	attr := attrs.Attr(key)
	if attr == nil {
		attr = &Attr{
			Key:    key,
			Values: []string{value},
		}
		attrs.Set(attr)
	}
	attr.Set(value)
	return s
}

func attrFirstValue(w AttrsMapper, key string) string {
	attrs := w.FindAttrs()
	if attrs == nil {
		return ""
	}
	attr := attrs.Attr(key)
	if attr == nil {
		return ""
	}
	return attr.First()
}

// Get 获取属性值
func (s *StyleAttr) Get(key string) string {
	return attrFirstValue(s, key)
}

// Width 设置宽度
func (s *StyleAttr) Width(w string) *StyleAttr {
	return s.set("width", w)
}

// MinWidth 设置最小宽度
func (s *StyleAttr) MinWidth(w string) *StyleAttr {
	return s.set("min-width", w)
}

// MaxWidth 设置最大新宽度
func (s *StyleAttr) MaxWidth(w string) *StyleAttr {
	return s.set("max-width", w)
}

// Height 设置高度
func (s *StyleAttr) Height(h string) *StyleAttr {
	return s.set("height", h)
}

// MinHeight 设置最小高度
func (s *StyleAttr) MinHeight(h string) *StyleAttr {
	return s.set("min-height", h)
}

// MaxHeight 设置最大高度
func (s *StyleAttr) MaxHeight(h string) *StyleAttr {
	return s.set("max-height", h)
}

// Color 设置前景/字体颜色
func (s *StyleAttr) Color(color string) *StyleAttr {
	return s.set("color", color)
}

// BackgroundColor 设置背景演示
func (s *StyleAttr) BackgroundColor(color string) *StyleAttr {
	return s.set("background-color", color)
}

// TextAlign 设置内容对齐方式
func (s *StyleAttr) TextAlign(align string) *StyleAttr {
	return s.set("text-align", align)
}

// Margin 设置外边距
func (s *StyleAttr) Margin(val string) *StyleAttr {
	return s.set("margin", val)
}

// Padding 设置内边距
func (s *StyleAttr) Padding(val string) *StyleAttr {
	return s.set("padding", val)
}

// Font 设置字体
func (s *StyleAttr) Font(val string) *StyleAttr {
	return s.set("font", val)
}

// FontSize 设置字体大小
func (s *StyleAttr) FontSize(val string) *StyleAttr {
	return s.set("font-size", val)
}

// FontWeight 设置字体粗细
func (s *StyleAttr) FontWeight(val string) *StyleAttr {
	return s.set("font-weight", val)
}

// FontFamily 设置字体系列（字体族）
func (s *StyleAttr) FontFamily(val string) *StyleAttr {
	return s.set("font-family", val)
}

// Border 设置边框属性
func (s *StyleAttr) Border(val string) *StyleAttr {
	return s.set("border", val)
}

// HTML 实现 Element 接口
func (s *StyleAttr) HTML() ([]byte, error) {
	attrs := s.FindAttrs()
	if attrs == nil {
		return nil, nil
	}
	return attrsHTML(attrs, ":", "", "; ")
}

// SetTo 将样式信息设置到指定的属性集合
func (s *StyleAttr) SetTo(a AttrsMapper) error {
	code, err := s.HTML()
	if err != nil {
		return err
	}
	if len(code) == 0 {
		return nil
	}
	a.MustAttrs().MustAttr("style").Set(string(code))
	return nil
}
