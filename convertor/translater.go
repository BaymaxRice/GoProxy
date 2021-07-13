package convertor

import (
	"errors"
)

const passwordLen = 1024

type Convertor interface {
	Init()
	// Encrypt 加密方法
	Encrypt(st []byte) []byte
	// Decrypt 解密方法
	Decrypt(st []byte) []byte
	// GenNewPW 生成新密码
	GenNewPW(newPW []byte)
	// GetPW 获取密码
	GetPW() []byte
}

type conf struct {
	// 加密密码
	EncryptPassword [passwordLen]byte
	DecryptPassword [passwordLen]byte
}

var TranslateMap = map[string]Convertor{
	"plaintext": &Plaintext{},
	"replace":   &Replace{},
}

func GetNewConvertor(mode string) (Convertor, error) {
	obj, ok := TranslateMap[mode]
	if !ok {
		return nil, errors.New("mode 不存在")
	}
	obj.Init()
	return obj, nil
}
