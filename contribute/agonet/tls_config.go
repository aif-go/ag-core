package agonet

type TLSType = string

const (
	TLSTypeNone TLSType = "none"
	TLSTypeTLS  TLSType = "tls"
	TLSTypeTLCP TLSType = "tlcp"
)

type TLSConfig struct {
}

type TLCPConfig struct {
}
