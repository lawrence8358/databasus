// Package webtls keeps the HTTPS certificate the Web UI uses when the installer opens it to the
// local network: a PEM certificate chain and private key under <data folder>\tls, imported from
// a PFX file or generated as a self-signed certificate. Both files are restricted to the account
// that wrote them, SYSTEM and Administrators.
package webtls

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"databasus-backend/internal/util/securefile"
)

const (
	folderName   = "tls"
	certFileName = "databasus.crt"
	keyFileName  = "databasus.key"
	newFileExt   = ".new"
)

type Files struct {
	CertFile string
	KeyFile  string
}

// Summary describes a certificate in the words the installer log and the launcher log show.
type Summary struct {
	Subject           string
	Names             []string
	NotAfter          time.Time
	SHA256Fingerprint string
	IsSelfSigned      bool
}

func GetFiles(dataFolder string) Files {
	return Files{
		CertFile: filepath.Join(dataFolder, folderName, certFileName),
		KeyFile:  filepath.Join(dataFolder, folderName, keyFileName),
	}
}

// Check loads the stored certificate and key and fails when they are missing, do not match or the
// certificate is outside its validity period.
func Check(dataFolder string, now time.Time) (Summary, error) {
	files := GetFiles(dataFolder)

	certificate, err := tls.LoadX509KeyPair(files.CertFile, files.KeyFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Summary{}, fmt.Errorf("no HTTPS certificate in %s", filepath.Dir(files.CertFile))
		}

		return Summary{}, fmt.Errorf(
			"the HTTPS certificate in %s cannot be used: %w",
			filepath.Dir(files.CertFile),
			err,
		)
	}

	leaf, err := x509.ParseCertificate(certificate.Certificate[0])
	if err != nil {
		return Summary{}, fmt.Errorf("parse %s: %w", files.CertFile, err)
	}

	if err := checkValidityPeriod(leaf, now); err != nil {
		return Summary{}, err
	}

	return summarize(leaf), nil
}

func (summary Summary) String() string {
	kind := "certificate"
	if summary.IsSelfSigned {
		kind = "self-signed certificate"
	}

	return fmt.Sprintf(
		"%s %s for %s, valid until %s, SHA-256 fingerprint %s",
		kind,
		summary.Subject,
		strings.Join(summary.Names, ", "),
		summary.NotAfter.Format(time.DateOnly),
		summary.SHA256Fingerprint,
	)
}

func checkValidityPeriod(leaf *x509.Certificate, now time.Time) error {
	if now.After(leaf.NotAfter) {
		return fmt.Errorf(
			"the HTTPS certificate %s expired on %s",
			leaf.Subject,
			leaf.NotAfter.Format(time.DateOnly),
		)
	}

	if now.Before(leaf.NotBefore) {
		return fmt.Errorf(
			"the HTTPS certificate %s is not valid before %s",
			leaf.Subject,
			leaf.NotBefore.Format(time.DateOnly),
		)
	}

	return nil
}

func summarize(leaf *x509.Certificate) Summary {
	fingerprint := sha256.Sum256(leaf.Raw)

	names := append([]string{}, leaf.DNSNames...)
	for _, ip := range leaf.IPAddresses {
		names = append(names, ip.String())
	}

	return Summary{
		Subject:           leaf.Subject.String(),
		Names:             names,
		NotAfter:          leaf.NotAfter,
		SHA256Fingerprint: strings.ToUpper(hex.EncodeToString(fingerprint[:])),
		IsSelfSigned:      isSelfSigned(leaf),
	}
}

// CheckSignatureFrom would require the CA flag, which a self-signed server certificate lacks.
func isSelfSigned(leaf *x509.Certificate) bool {
	return leaf.Subject.String() == leaf.Issuer.String() &&
		leaf.CheckSignature(leaf.SignatureAlgorithm, leaf.RawTBSCertificate, leaf.Signature) == nil
}

// store replaces the stored pair only after both new files are written, so a failure leaves the
// previous certificate in place.
func store(dataFolder string, certPEM, keyPEM []byte) error {
	if _, err := tls.X509KeyPair(certPEM, keyPEM); err != nil {
		return fmt.Errorf("the certificate and private key do not form a usable pair: %w", err)
	}

	files := GetFiles(dataFolder)
	folder := filepath.Dir(files.CertFile)

	if err := os.MkdirAll(folder, 0o700); err != nil {
		return fmt.Errorf("create %s: %w", folder, err)
	}

	if err := securefile.RestrictDir(folder, 0o700); err != nil {
		return fmt.Errorf("restrict %s: %w", folder, err)
	}

	newFiles := []struct {
		path    string
		content []byte
	}{
		{files.KeyFile, keyPEM},
		{files.CertFile, certPEM},
	}

	for _, file := range newFiles {
		if err := securefile.WriteFile(file.path+newFileExt, file.content, 0o600); err != nil {
			return fmt.Errorf("write %s: %w", file.path+newFileExt, err)
		}
	}

	for _, file := range newFiles {
		if err := os.Rename(file.path+newFileExt, file.path); err != nil {
			return fmt.Errorf("replace %s: %w", file.path, err)
		}
	}

	return nil
}

func encodeCertificates(certificates []*x509.Certificate) []byte {
	var encoded []byte

	for _, certificate := range certificates {
		encoded = append(encoded, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw})...)
	}

	return encoded
}

func encodePrivateKey(key any) ([]byte, error) {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("encode the private key: %w", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), nil
}
