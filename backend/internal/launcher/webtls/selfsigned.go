package webtls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"slices"
	"strings"
	"time"
)

const selfSignedValidity = 5 * 365 * 24 * time.Hour

// HostNames are the names and addresses colleagues may type to reach this computer.
type HostNames struct {
	DNSNames []string
	IPs      []net.IP
}

// DiscoverHostNames returns this computer's host name, its DNS name when DNS knows one, localhost
// and the addresses of its network adapters, so the certificate matches whichever one is typed.
func DiscoverHostNames() HostNames {
	names := HostNames{
		DNSNames: []string{"localhost"},
		IPs:      []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		names.DNSNames = appendName(names.DNSNames, hostname)

		if canonicalName, err := net.LookupCNAME(hostname); err == nil {
			names.DNSNames = appendName(names.DNSNames, strings.TrimSuffix(canonicalName, "."))
		}
	}

	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return names
	}

	for _, address := range addresses {
		network, isIPNet := address.(*net.IPNet)
		if !isIPNet || network.IP.IsLoopback() || network.IP.IsLinkLocalUnicast() {
			continue
		}

		names.IPs = append(names.IPs, network.IP)
	}

	return names
}

// GenerateSelfSigned replaces the stored certificate with a new self-signed one for names.
func GenerateSelfSigned(dataFolder string, names HostNames, now time.Time) (Summary, error) {
	certPEM, keyPEM, summary, err := createSelfSigned(names, now)
	if err != nil {
		return Summary{}, err
	}

	if err := store(dataFolder, certPEM, keyPEM); err != nil {
		return Summary{}, err
	}

	return summary, nil
}

func createSelfSigned(names HostNames, now time.Time) (certPEM, keyPEM []byte, summary Summary, err error) {
	if len(names.DNSNames) == 0 {
		return nil, nil, Summary{}, errors.New("a self-signed certificate needs at least one host name")
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, Summary{}, fmt.Errorf("generate a private key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 127))
	if err != nil {
		return nil, nil, Summary{}, fmt.Errorf("generate a serial number: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: getSubjectHostName(names), Organization: []string{"Databasus"}},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(selfSignedValidity),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              names.DNSNames,
		IPAddresses:           names.IPs,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, Summary{}, fmt.Errorf("create the self-signed certificate: %w", err)
	}

	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, Summary{}, err
	}

	keyPEM, err = encodePrivateKey(privateKey)
	if err != nil {
		return nil, nil, Summary{}, err
	}

	return encodeCertificates([]*x509.Certificate{leaf}), keyPEM, summarize(leaf), nil
}

// The first name other than localhost is the computer's own, which is what a browser shows.
func getSubjectHostName(names HostNames) string {
	for _, name := range names.DNSNames {
		if name != "localhost" {
			return name
		}
	}

	return names.DNSNames[0]
}

func appendName(names []string, name string) []string {
	isKnown := slices.ContainsFunc(names, func(existing string) bool { return strings.EqualFold(existing, name) })
	if name == "" || isKnown {
		return names
	}

	return append(names, name)
}
