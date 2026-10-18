package webtls

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"software.sslmate.com/src/go-pkcs12"

	"databasus-backend/internal/util/securefile"
)

var testNow = time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)

func Test_GenerateSelfSigned_WritesPairThatCheckAccepts(t *testing.T) {
	dataFolder := t.TempDir()
	names := HostNames{DNSNames: []string{"localhost", "db-server"}, IPs: []net.IP{net.ParseIP("192.168.1.20")}}

	generated, err := GenerateSelfSigned(dataFolder, names, testNow, securefile.CurrentAccount)
	require.NoError(t, err)

	checked, err := Check(dataFolder, testNow)
	require.NoError(t, err)

	assert.Equal(t, generated.SHA256Fingerprint, checked.SHA256Fingerprint)
	assert.True(t, checked.IsSelfSigned)
	assert.Equal(t, "CN=db-server,O=Databasus", checked.Subject)
	assert.ElementsMatch(t, []string{"localhost", "db-server", "192.168.1.20"}, checked.Names)
	assert.NoFileExists(t, GetFiles(dataFolder).CertFile+newFileExt)
}

func Test_Check_WhenNoCertificateIsStored_ReturnsError(t *testing.T) {
	_, err := Check(t.TempDir(), testNow)

	assert.ErrorContains(t, err, "no HTTPS certificate")
}

func Test_Check_WhenCertificateHasExpired_ReturnsError(t *testing.T) {
	dataFolder := t.TempDir()
	_, err := GenerateSelfSigned(dataFolder, HostNames{DNSNames: []string{"localhost"}}, testNow, securefile.CurrentAccount)
	require.NoError(t, err)

	_, err = Check(dataFolder, testNow.Add(6*365*24*time.Hour))

	assert.ErrorContains(t, err, "expired")
}

func Test_ImportPfx_WhenPfxUsesModernEncryption_StoresChainAndKey(t *testing.T) {
	dataFolder := t.TempDir()
	pfxData := createTestPfx(t, "s3cret", testNow)

	summary, err := ImportPfx(dataFolder, pfxData, "s3cret", testNow, securefile.CurrentAccount)
	require.NoError(t, err)

	files := GetFiles(dataFolder)
	pair, err := tls.LoadX509KeyPair(files.CertFile, files.KeyFile)
	require.NoError(t, err)
	assert.Len(t, pair.Certificate, 1)

	checked, err := Check(dataFolder, testNow)
	require.NoError(t, err)
	assert.Equal(t, summary.SHA256Fingerprint, checked.SHA256Fingerprint)
}

func Test_ImportPfx_WhenPasswordIsWrong_KeepsPreviousCertificate(t *testing.T) {
	dataFolder := t.TempDir()
	previous, err := GenerateSelfSigned(dataFolder, HostNames{DNSNames: []string{"localhost"}}, testNow, securefile.CurrentAccount)
	require.NoError(t, err)

	_, err = ImportPfx(dataFolder, createTestPfx(t, "s3cret", testNow), "wrong", testNow, securefile.CurrentAccount)
	assert.ErrorContains(t, err, "wrong password")

	checked, err := Check(dataFolder, testNow)
	require.NoError(t, err)
	assert.Equal(t, previous.SHA256Fingerprint, checked.SHA256Fingerprint)
}

func Test_ParsePfx_WhenCertificateHasExpired_ReturnsError(t *testing.T) {
	pfxData := createTestPfx(t, "s3cret", testNow.Add(-10*365*24*time.Hour))

	_, _, _, err := ParsePfx(pfxData, "s3cret", testNow)

	assert.ErrorContains(t, err, "expired")
}

func Test_ParsePfx_WhenDataIsNotPfx_ReturnsError(t *testing.T) {
	_, _, _, err := ParsePfx([]byte("not a pfx"), "", testNow)

	assert.Error(t, err)
}

func Test_DiscoverHostNames_IncludesLocalhostAndThisComputer(t *testing.T) {
	hostname, err := os.Hostname()
	require.NoError(t, err)

	names := DiscoverHostNames()

	assert.Contains(t, names.DNSNames, "localhost")
	assert.Contains(t, names.DNSNames, hostname)
}

func createTestPfx(t *testing.T, password string, issuedAt time.Time) []byte {
	t.Helper()

	certPEM, keyPEM, _, err := createSelfSigned(HostNames{DNSNames: []string{"agency.example"}}, issuedAt)
	require.NoError(t, err)

	certBlock, _ := pem.Decode(certPEM)
	certificate, err := x509.ParseCertificate(certBlock.Bytes)
	require.NoError(t, err)

	keyBlock, _ := pem.Decode(keyPEM)
	privateKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	require.NoError(t, err)

	pfxData, err := pkcs12.Modern.Encode(privateKey, certificate, nil, password)
	require.NoError(t, err)

	return pfxData
}
