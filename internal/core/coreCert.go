package core

type Cert struct {
	IsTLS				bool
	CertPEM 			[]byte 		
	CertPrivKeyPEM	    []byte     
}