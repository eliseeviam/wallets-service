package wallet

type Wallet interface {
	Name() string
}

type DefaultWallet struct {
	name string
}

func NewDefaultWallet(name string) Wallet {
	return &DefaultWallet{
		name: name,
	}
}

func (dw *DefaultWallet) Name() string {
	return dw.name
}
