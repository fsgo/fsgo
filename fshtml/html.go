// Copyright(C) 2022 github.com/hidu  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2022/5/4

//
// https://html.spec.whatwg.org/#element-interfaces

package fshtml

// NewDiv 创建一个 <div>
func NewDiv() *Any {
	return NewAny("div")
}

// NewNav 创建一个 <nav>
func NewNav() *Any {
	return NewAny("nav")
}

// NewP 创建一个 <p>
func NewP() *Any {
	return NewAny("p")
}

// NewHead 创建一个 <head>
func NewHead() *Any {
	return NewAny("head")
}

// NewBody 创建一个 <body>
func NewBody() *Any {
	return NewAny("body")
}

// NewDL 创建一个 <dl>
func NewDL() *Any {
	return NewAny("dl")
}

// NewDT 创建一个 <dt>
func NewDT() *Any {
	return NewAny("dt")
}

// NewArticle 创建一个 <article>
func NewArticle() *Any {
	return NewAny("article")
}

// NewUl 转换为 ul
func NewUl(values StringSlice) *Any {
	return &Any{
		Tag:  "ul",
		Body: values.ToElements("li", nil),
	}
}

// NewPre 创建一个 <pre>
func NewPre() *Any {
	return NewAny("pre")
}

// NewCode 创建一个 <code>
func NewCode() *Any {
	return NewAny("code")
}

// NewFigure 创建一个 <figure>
func NewFigure() *Any {
	return NewAny("figure")
}

// NewFigcaption 创建一个 <figcaption>
func NewFigcaption() *Any {
	return NewAny("figcaption")
}

// NewOl 转换为 ol
func NewOl(values StringSlice) *Any {
	return &Any{
		Tag:  "ol",
		Body: values.ToElements("li", nil),
	}
}

type selfCloseTag struct {
	WithAttrs
}

func (m *selfCloseTag) set(key string, value string) {
	attr := &Attr{
		Key:    key,
		Values: []string{value},
	}
	m.MustAttrs().Set(attr)
}

func (m *selfCloseTag) html(begin string) ([]byte, error) {
	bw := newBufWriter()
	bw.Write(begin)
	bw.WriteWithSep(" ", m.Attrs)
	bw.Write("/>")
	return bw.HTML()
}

// NewIMG 创建一个  img  标签
func NewIMG(src string) *IMG {
	return (&IMG{}).SRC(src)
}

// IMG 图片 img 标签
type IMG struct {
	selfCloseTag
}

func (m *IMG) set(key string, value string) *IMG {
	m.selfCloseTag.set(key, value)
	return m
}

// SRC 设置 src 属性
func (m *IMG) SRC(src string) *IMG {
	return m.set("src", src)
}

// ALT 设置 alt 属性
func (m *IMG) ALT(alt string) *IMG {
	return m.set("alt", alt)
}

// HTML 转换为 html
func (m *IMG) HTML() ([]byte, error) {
	return m.html("<img")
}

var _ Element = (*A)(nil)

// NewA 创建一个 <a>
func NewA(href string) *A {
	return (&A{}).Href(href)
}

// A 地址标签 <a>
type A struct {
	selfCloseTag
}

func (a *A) set(key string, value string) *A {
	a.selfCloseTag.set(key, value)
	return a
}

// Href 设置 href 属性
func (a *A) Href(href string) *A {
	return a.set("href", href)
}

// HrefLang 设置 hreflang 属性
func (a *A) HrefLang(hrefLang string) *A {
	return a.set("hreflang", hrefLang)
}

// Title 设置 Title 属性
func (a *A) Title(title string) *A {
	return a.set("title", title)
}

// Target 设置 target 属性
func (a *A) Target(target string) *A {
	return a.set("target", target)
}

// Rel 设置 rel 属性
func (a *A) Rel(rel string) *A {
	return a.set("rel", rel)
}

// Type 设置 type 属性
func (a *A) Type(tp string) *A {
	return a.set("type", tp)
}

// Ping 设置 ping 属性
func (a *A) Ping(ping string) *A {
	return a.set("ping", ping)
}

// HTML 转换为 html
func (a *A) HTML() ([]byte, error) {
	return a.html("<a")
}

// NewMeta 创建一个新的 <meta>
func NewMeta() *Meta {
	return &Meta{}
}

var _ Element = (*Meta)(nil)

// Meta 页面元信息标签 meth
type Meta struct {
	selfCloseTag
}

func (a *Meta) set(key string, value string) *Meta {
	a.selfCloseTag.set(key, value)
	return a
}

// Name 设置 name 属性
func (a *Meta) Name(name string) *Meta {
	return a.set("name", name)
}

// Charset 设置 charset 属性
func (a *Meta) Charset(charset string) *Meta {
	return a.set("charset", charset)
}

// HTTPEquiv 设置 http-equiv 属性
func (a *Meta) HTTPEquiv(equiv string) *Meta {
	return a.set("http-equiv", equiv)
}

// Content 设置 content 属性
func (a *Meta) Content(content string) *Meta {
	return a.set("content", content)
}

// Media 设置 media 属性
func (a *Meta) Media(media string) *Meta {
	return a.set("media", media)
}

// HTML 转换为 html
func (a *Meta) HTML() ([]byte, error) {
	return a.html("<meta")
}

// NewScript script 标签
func NewScript() *Any {
	return NewAny("script")
}

// NewStyle style 标签
func NewStyle() *Any {
	a := NewAny("style")
	SetType(a, "text/css")
	return a
}

// NewLink 创建一个新的 link 标签
func NewLink() *Link {
	return &Link{}
}

// Link 页面元素 link 标签
type Link struct {
	selfCloseTag
}

// Rel 设置 rel 属性
func (a *Link) Rel(rel string) *Link {
	a.set("rel", rel)
	return a
}

// Type 设置 tp 属性
func (a *Link) Type(tp string) *Link {
	a.set("type", tp)
	return a
}

// Href 设置 href 属性
func (a *Link) Href(href string) *Link {
	a.set("href", href)
	return a
}

// HTML 转换为 html
func (a *Link) HTML() ([]byte, error) {
	return a.html("<link")
}

// NewInput 创建一个 input 标签
func NewInput(tp string) *Any {
	input := &Any{
		Tag:       "input",
		SelfClose: true,
	}
	SetType(input, tp)
	return input
}

// NewForm 创建一个 form
func NewForm(method string, action string) *Any {
	f := NewAny("form")
	SetMethod(f, method)
	SetAction(f, action)
	return f
}

// NewSubmit 创建一个 submit 标签
func NewSubmit(value string) *Any {
	s := NewInput("submit")
	SetValue(s, value)
	return s
}

// NewFieldSet 创建一个 fieldset
func NewFieldSet() *Any {
	return NewAny("fieldset")
}

// NewLegend 创建一个 legend
func NewLegend(e Element) *Any {
	a := &Any{
		Tag:  "legend",
		Body: ToElements(e),
	}
	return a
}

// NewLabel 创建一个 label
func NewLabel(e Element) *Any {
	return &Any{
		Tag:  "label",
		Body: ToElements(e),
	}
}
