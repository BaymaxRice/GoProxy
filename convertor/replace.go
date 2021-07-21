package convertor

import (
	"math/rand"
	"time"
)

type Replace struct {
	// 配置
	conf conf
}

func (re *Replace) Init() {
	rand.Seed(time.Now().UnixNano())
	intArr := rand.Perm(passwordLen)
	re.conf.EncryptPassword = make([]byte, passwordLen)
	for key, value := range intArr {
		re.conf.EncryptPassword[key] = byte(value)
	}
}

func (re *Replace) GetPW() []byte {
	return re.conf.EncryptPassword[:]
}

func (re *Replace) GenNewPW(newPW []byte) {
	re.conf.EncryptPassword = make([]byte, len(newPW))
	for key, value := range newPW {
		re.conf.EncryptPassword[key] = byte(value)
	}
}

func (re Replace) Encrypt(st []byte) []byte {
	var ret = make([]byte, len(st))
	for k, v := range re.conf.EncryptPassword {
		ret[v] = st[k]
	}
	return ret
}

func (re Replace) Decrypt(st []byte) []byte {
	var ret = make([]byte, len(st))
	for k, v := range re.conf.EncryptPassword {
		ret[k] = st[v]
	}
	return ret
}
