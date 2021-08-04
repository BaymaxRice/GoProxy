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
	re.conf.DecryptPassword = make([]byte, passwordLen)
	for key, value := range intArr {
		re.conf.EncryptPassword[key] = byte(value)
		re.conf.DecryptPassword[value] = byte(key)
	}
}

func (re *Replace) GetPW() []byte {
	return re.conf.EncryptPassword[:]
}

func (re *Replace) GenNewPW(newPW []byte) {
	if len(newPW) > 256 {
		panic("replace 密码不能长于256")
	}
	re.conf.EncryptPassword = make([]byte, len(newPW))
	re.conf.DecryptPassword = make([]byte, len(newPW))
	for key, value := range newPW {
		re.conf.EncryptPassword[key] = byte(value)
		re.conf.DecryptPassword[value] = byte(key)
	}
}

func (re Replace) Encrypt(st []byte) []byte {
	for k, v := range st {
		st[k] = re.conf.EncryptPassword[v]
	}
	return st
}

func (re Replace) Decrypt(st []byte) []byte {
	for k, v := range st {
		st[k] = re.conf.DecryptPassword[v]
	}
	return st
}
