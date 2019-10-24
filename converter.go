package ssr_demo

type Converter interface {
	// 数据转换方法
	TransLater(st []byte) []byte
}

