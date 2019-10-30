package ssr_demo

type Converter interface {
	// 加密方法
	Encrypt(st []byte) []byte
	// 解密方法
	Decrypt(st []byte) []byte
	// 生成新密码
	GenNewPW(newPW []byte)
	// 获取密码
	GetPW() []byte
}
