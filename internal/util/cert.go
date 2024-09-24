package util

import(
	"os"

	"github.com/joho/godotenv"
	"github.com/go-account/internal/core"
)

func GetCertEnv() core.Cert {
	childLogger.Debug().Msg("GetCertEnv")

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Info().Err(err).Msg("env file not found!!!")
	}

	var cert		core.Cert

	if os.Getenv("TLS") !=  "false" {	
		cert.IsTLS = true
		childLogger.Info().Err(err).Msg("*** Loading server_account_B64.crt ***")

		cert.CertPEM, err = os.ReadFile("/var/pod/cert/server_account_B64.crt") // server_account_B64.crt
		if err != nil {
			childLogger.Info().Err(err).Msg("cert certPEM not found")
		} 
	
		childLogger.Info().Err(err).Msg("*** Loading server_account_B64.key ***")

		cert.CertPrivKeyPEM, err = os.ReadFile("/var/pod/cert/server_account_B64.key") // server_account_B64.key
		if err != nil {
			childLogger.Info().Err(err).Msg("cert CertPrivKeyPEM not found")
		}
	}

	return cert
}