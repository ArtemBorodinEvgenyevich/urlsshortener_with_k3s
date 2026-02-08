package domain

type Provider string

const ProviderSID Provider = "SID"

func (p Provider) IsValid() bool {
	switch p {
	case ProviderSID:
		return true
	default:
		return false
	}
}
