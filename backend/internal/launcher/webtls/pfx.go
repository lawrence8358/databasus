package webtls

import (
	"crypto/x509"
	"fmt"
	"time"

	"software.sslmate.com/src/go-pkcs12"

	"databasus-backend/internal/util/securefile"
)

// ParsePfx decodes a PFX (PKCS #12) file as Windows and OpenSSL export it, including the AES
// encryption their current versions use, and checks that the certificate can serve HTTPS now.
func ParsePfx(pfxData []byte, password string, now time.Time) (certPEM, keyPEM []byte, summary Summary, err error) {
	privateKey, leaf, chain, err := pkcs12.DecodeChain(pfxData, password)
	if err != nil {
		return nil, nil, Summary{}, fmt.Errorf(
			"cannot read the PFX file (wrong password, or not a PFX file with a private key): %w",
			err,
		)
	}

	if err := checkValidityPeriod(leaf, now); err != nil {
		return nil, nil, Summary{}, err
	}

	if len(leaf.ExtKeyUsage) > 0 && !hasServerAuthUsage(leaf) {
		return nil, nil, Summary{}, fmt.Errorf(
			"the certificate %s is not issued for server authentication",
			leaf.Subject,
		)
	}

	keyPEM, err = encodePrivateKey(privateKey)
	if err != nil {
		return nil, nil, Summary{}, err
	}

	certPEM = encodeCertificates(append([]*x509.Certificate{leaf}, chain...))

	return certPEM, keyPEM, summarize(leaf), nil
}

func ImportPfx(
	dataFolder string,
	pfxData []byte,
	password string,
	now time.Time,
	readers securefile.Readers,
) (Summary, error) {
	certPEM, keyPEM, summary, err := ParsePfx(pfxData, password, now)
	if err != nil {
		return Summary{}, err
	}

	if err := store(dataFolder, certPEM, keyPEM, readers); err != nil {
		return Summary{}, err
	}

	return summary, nil
}

func hasServerAuthUsage(certificate *x509.Certificate) bool {
	for _, usage := range certificate.ExtKeyUsage {
		if usage == x509.ExtKeyUsageServerAuth || usage == x509.ExtKeyUsageAny {
			return true
		}
	}

	return false
}
