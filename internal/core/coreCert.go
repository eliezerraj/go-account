package core

type Cert struct {
	CertPEM 			[]byte 		`json:"cert_pem"`
	CertPrivKeyPEM	    []byte     	`json:"cert_priv"`
}