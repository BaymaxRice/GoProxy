package convertor

type Plaintext struct {
	// 配置
	conf conf
}

func (re *Plaintext) Init() {

}

func (re *Plaintext) GetPW() []byte {
	return nil
}

func (re *Plaintext) GenNewPW(newPW []byte) {

}

func (re Plaintext) Encrypt(st []byte) []byte {
	return st
}

func (re Plaintext) Decrypt(st []byte) []byte {
	return st
}
