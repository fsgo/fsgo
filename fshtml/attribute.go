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
	Attributes interface {
		SetAttr(key string, val AttrValue)
		Attr(key string) AttrValue
		MustAttr(key string) AttrValue
		DeleteAttr(key string)
		AttrKeys() []string
		HTML
	}

	AttrValue interface {
		HTML

		Value() []string
		Sep() string
		SetSep(sep string)
		Set(value ...string)
		Add(value ...string)
		Delete(value ...string)
	}
)

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

func (a *attrs) DeleteAttr(key string) {
	if len(a.values) == 0 {
		return
	}
	delete(a.values, key)
	if a.keys.Has(key) {
		a.keys.Delete(key)
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
	if len(a.keys) == 0 {
		return nil, nil
	}
	bw := newBufWriter()
	for i := 0; i < len(a.keys); i++ {
		attrKey := a.keys[i]
		bw.Write(attrKey)
		bf, err := a.values[attrKey].HTML()
		if err != nil {
			return nil, err
		}
		if len(bf) > 0 {
			bw.Write(`="`, bf, `"`)
		}
		if i != len(a.keys)-1 {
			bw.Write(" ")
		}
	}
	return bw.HTML()
}

func NewAttrValue(sep string) AttrValue {
	return &attrValue{
		sep: sep,
	}
}

var _ AttrValue = (*attrValue)(nil)

type attrValue struct {
	values fstypes.StringSlice
	sep    string
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

func SetClass(w Attributes, class ...string) {
	findOrCreateAttr(w, attrClass, " ").Set(class...)
}

func AddClass(w Attributes, class ...string) {
	findOrCreateAttr(w, attrClass, " ").Add(class...)
}

func DeleteClass(w Attributes, class ...string) {
	attr := w.Attr(attrClass)
	if attr == nil {
		return
	}
	attr.Delete(class...)
}

func SetID(w Attributes, id string) {
	findOrCreateAttr(w, attrID, " ").Set(id)
}

func SetName(w Attributes, name string) {
	findOrCreateAttr(w, attrName, " ").Set(name)
}

type AttrStyle struct {
	attrs Attributes
}

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

func (s *AttrStyle) Width(w string) *AttrStyle {
	return s.Set("width", w)
}

func (s *AttrStyle) MinWidth(w string) *AttrStyle {
	return s.Set("min-width", w)
}

func (s *AttrStyle) MaxWidth(w string) *AttrStyle {
	return s.Set("max-width", w)
}

func (s *AttrStyle) Height(h string) *AttrStyle {
	return s.Set("height", h)
}

func (s *AttrStyle) MinHeight(h string) *AttrStyle {
	return s.Set("min-height", h)
}
func (s *AttrStyle) MaxHeight(h string) *AttrStyle {
	return s.Set("max-height", h)
}

func (s *AttrStyle) Color(color string) *AttrStyle {
	return s.Set("color", color)
}

func (s *AttrStyle) BackgroundColor(color string) *AttrStyle {
	return s.Set("background-color", color)
}

func (s *AttrStyle) TextAlign(align string) *AttrStyle {
	return s.Set("text-align", align)
}

func (s *AttrStyle) Margin(val string) *AttrStyle {
	return s.Set("margin", val)
}

func (s *AttrStyle) Padding(val string) *AttrStyle {
	return s.Set("padding", val)
}

func (s *AttrStyle) Font(val string) *AttrStyle {
	return s.Set("font", val)
}

func (s *AttrStyle) FontSize(val string) *AttrStyle {
	return s.Set("font-size", val)
}

func (s *AttrStyle) FontWeight(val string) *AttrStyle {
	return s.Set("font-weight", val)
}

func (s *AttrStyle) FontFamily(val string) *AttrStyle {
	return s.Set("font-family", val)
}

func (s *AttrStyle) Border(val string) *AttrStyle {
	return s.Set("border", val)
}

func (s *AttrStyle) HTML() ([]byte, error) {
	if s.attrs == nil {
		return nil, nil
	}
	return s.attrs.HTML()
}
