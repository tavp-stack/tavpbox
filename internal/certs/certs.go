package certs

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
)

func certsDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".tavpbox", "certs")
	os.MkdirAll(dir, 0755)
	return dir
}

type LegoUser struct {
	Email        string
	Registration *registration.Resource
	key          *ecdsa.PrivateKey
}

func (u *LegoUser) GetEmail() string                        { return u.Email }
func (u *LegoUser) GetRegistration() *registration.Resource { return u.Registration }
func (u *LegoUser) GetPrivateKey() crypto.PrivateKey { return u.key }

// GenerateWildcardCert generates a wildcard cert using lego with Cloudflare DNS
func GenerateWildcardCert(domain, cfToken string) (certPath, keyPath string, err error) {
	dir := certsDir()
	certPath = filepath.Join(dir, domain+".pem")
	keyPath = filepath.Join(dir, domain+"-key.pem")

	// Check if cert exists and is valid
	if isCertValid(certPath) {
		return certPath, keyPath, nil
	}

	// Create user key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	user := &LegoUser{
		Email: "admin@" + domain,
		key:   privateKey,
	}

	// Create lego config
	config := lego.NewConfig(user)
	config.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"

	client, err := lego.NewClient(config)
	if err != nil {
		return "", "", err
	}

	// Setup Cloudflare DNS provider
	cfConfig := cloudflare.NewDefaultConfig()
	cfConfig.AuthToken = cfToken
	cfConfig.PropagationTimeout = 2 * time.Minute

	provider, err := cloudflare.NewDNSProviderConfig(cfConfig)
	if err != nil {
		return "", "", fmt.Errorf("cloudflare provider: %w", err)
	}

	client.Challenge.SetDNS01Provider(provider)

	// Register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return "", "", fmt.Errorf("register: %w", err)
	}
	user.Registration = reg

	// Request certificate
	certResource, err := client.Certificate.Obtain(certificate.ObtainRequest{
		Domains: []string{"*." + domain, domain},
		Bundle:  true,
	})
	if err != nil {
		return "", "", fmt.Errorf("obtain cert: %w", err)
	}

	// Save cert and key
	os.MkdirAll(dir, 0755)
	if err := os.WriteFile(certPath, certResource.Certificate, 0644); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(keyPath, certResource.PrivateKey, 0644); err != nil {
		return "", "", err
	}

	return certPath, keyPath, nil
}

// GetWildcardCert returns the path to the wildcard cert
func GetWildcardCert(domain string) (certPath, keyPath string) {
	dir := certsDir()
	certPath = filepath.Join(dir, domain+".pem")
	keyPath = filepath.Join(dir, domain+"-key.pem")

	if isCertValid(certPath) {
		return certPath, keyPath
	}
	return "", ""
}

func isCertValid(certPath string) bool {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return false
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return false
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}
	return time.Until(cert.NotAfter) > 30*24*time.Hour
}
