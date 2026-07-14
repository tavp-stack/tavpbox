package acme

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const acmeDirectoryURL = "https://acme-v02.api.letsencrypt.org/directory"

type Client struct {
	cfToken    string
	cfZone     string
	domain     string
	accountKey *ecdsa.PrivateKey
	dir        map[string]string
	accountURL string
	nonce      string
	http       *http.Client
}

func NewClient(cfToken, cfZone, domain string) *Client {
	return &Client{
		cfToken: cfToken,
		cfZone:  cfZone,
		domain:  domain,
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Generate() (certPath, keyPath string, err error) {
	dir := certsDir()
	certPath = filepath.Join(dir, c.domain+".pem")
	keyPath = filepath.Join(dir, c.domain+"-key.pem")

	// Check existing cert
	if data, err := os.ReadFile(certPath); err == nil {
		block, _ := pem.Decode(data)
		if block != nil {
			if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
				if time.Until(cert.NotAfter) > 30*24*time.Hour {
					return certPath, keyPath, nil
				}
			}
		}
	}

	// Generate account key
	c.accountKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	// Step 1: Get directory
	if err := c.getDirectory(); err != nil {
		return "", "", fmt.Errorf("directory: %w", err)
	}

	// Step 2: Create account
	if err := c.createAccount(); err != nil {
		return "", "", fmt.Errorf("account: %w", err)
	}

	// Step 3: Create order
	orderURL, finalizeURL, err := c.createOrder()
	if err != nil {
		return "", "", fmt.Errorf("order: %w", err)
	}

	// Step 4: Get challenge
	challengeURL, keyAuth, err := c.getChallenge(orderURL)
	if err != nil {
		return "", "", fmt.Errorf("challenge: %w", err)
	}

	// Step 5: Create DNS record
	recordID, err := c.createDNSRecord(keyAuth)
	if err != nil {
		return "", "", fmt.Errorf("dns record: %w", err)
	}
	defer c.deleteDNSRecord(recordID)

	// Step 6: Wait for DNS propagation
	fmt.Println("  Waiting for DNS propagation...")
	time.Sleep(15 * time.Second)

	// Step 7: Respond to challenge
	if err := c.respondChallenge(challengeURL); err != nil {
		return "", "", fmt.Errorf("respond: %w", err)
	}

	// Step 8: Wait for validation
	fmt.Println("  Waiting for validation...")
	time.Sleep(5 * time.Second)

	// Step 9: Finalize order
	certURL, err := c.finalizeOrder(finalizeURL)
	if err != nil {
		return "", "", fmt.Errorf("finalize: %w", err)
	}

	// Step 10: Get certificate
	certPEM, keyPEM, err := c.getCertificate(certURL)
	if err != nil {
		return "", "", fmt.Errorf("certificate: %w", err)
	}

	// Save
	os.MkdirAll(dir, 0755)
	os.WriteFile(certPath, certPEM, 0644)
	os.WriteFile(keyPath, keyPEM, 0644)

	return certPath, keyPath, nil
}

func certsDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".tavpbox", "certs")
	os.MkdirAll(dir, 0755)
	return dir
}

func (c *Client) getDirectory() error {
	resp, err := c.http.Get(acmeDirectoryURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var d map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return err
	}

	c.dir = make(map[string]string)
	for k, v := range d {
		if s, ok := v.(string); ok {
			c.dir[k] = s
		}
	}
	return nil
}

func (c *Client) createAccount() error {
	payload := map[string]interface{}{
		"termsOfServiceAgreed": true,
		"contact":             []string{"mailto:admin@tavp.my.id"},
	}

	resp, err := c.post(c.dir["newAccount"], payload, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Debug: print response
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  ACME account response: %d %s\n", resp.StatusCode, string(body))

	c.accountURL = resp.Header.Get("Location")
	if c.accountURL == "" {
		// Try from response body
		var result struct {
			Account string `json:"account"`
		}
		if err := json.Unmarshal(body, &result); err == nil && result.Account != "" {
			c.accountURL = result.Account
		}
	}
	if c.accountURL == "" {
		return fmt.Errorf("no account URL (status: %d)", resp.StatusCode)
	}
	return nil
}

func (c *Client) createOrder() (orderURL, finalizeURL string, err error) {
	payload := map[string]interface{}{
		"identifiers": []map[string]string{
			{"type": "dns", "value": "*." + c.domain},
			{"type": "dns", "value": c.domain},
		},
	}

	resp, err := c.post(c.dir["newOrder"], payload, true)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var order struct {
		Finalize string `json:"finalize"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return "", "", err
	}

	return resp.Header.Get("Location"), order.Finalize, nil
}

func (c *Client) getChallenge(orderURL string) (challengeURL, keyAuth string, err error) {
	resp, err := c.post(orderURL, nil, true)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var order struct {
		Authorizations []string `json:"authorizations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return "", "", err
	}

	if len(order.Authorizations) == 0 {
		return "", "", fmt.Errorf("no authorizations")
	}

	// Get authorization
	authResp, err := c.post(order.Authorizations[0], nil, true)
	if err != nil {
		return "", "", err
	}
	defer authResp.Body.Close()

	var auth struct {
		Challenges []struct {
			Type  string `json:"type"`
			URL   string `json:"url"`
			Token string `json:"token"`
		} `json:"challenges"`
	}
	if err := json.NewDecoder(authResp.Body).Decode(&auth); err != nil {
		return "", "", err
	}

	for _, ch := range auth.Challenges {
		if ch.Type == "dns-01" {
			ka := ch.Token + "." + c.thumbprint()
			return ch.URL, ka, nil
		}
	}

	return "", "", fmt.Errorf("no dns-01 challenge")
}

func (c *Client) createDNSRecord(keyAuth string) (string, error) {
	hash := sha256.Sum256([]byte(keyAuth))
	value := base64.RawURLEncoding.EncodeToString(hash[:])

	payload := map[string]interface{}{
		"type":    "TXT",
		"name":    "_acme-challenge." + c.domain,
		"content": value,
		"ttl":     60,
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", c.cfZone)

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
		Result  struct {
			ID string `json:"id"`
		} `json:"result"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if !result.Success {
		msgs := ""
		for _, e := range result.Errors {
			msgs += e.Message + "; "
		}
		return "", fmt.Errorf("cloudflare: %s", msgs)
	}

	return result.Result.ID, nil
}

func (c *Client) deleteDNSRecord(recordID string) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", c.cfZone, recordID)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bearer "+c.cfToken)
	c.http.Do(req)
}

func (c *Client) respondChallenge(challengeURL string) error {
	resp, err := c.post(challengeURL, map[string]interface{}{}, true)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) finalizeOrder(finalizeURL string) (string, error) {
	// Generate domain key
	domainKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", err
	}

	// Create CSR
	tpl := &x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: "*." + c.domain},
		DNSNames: []string{"*." + c.domain, c.domain},
	}
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, tpl, domainKey)
	if err != nil {
		return "", err
	}

	// Save domain key
	keyBytes, _ := x509.MarshalECPrivateKey(domainKey)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	keyPath := filepath.Join(certsDir(), c.domain+"-key.pem")
	os.WriteFile(keyPath, keyPEM, 0644)

	// Finalize
	payload := map[string]interface{}{
		"csr": base64.RawURLEncoding.EncodeToString(csrDER),
	}

	resp, err := c.post(finalizeURL, payload, true)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var order struct {
		Certificate string `json:"certificate"`
		Status      string `json:"status"`
	}
	json.NewDecoder(resp.Body).Decode(&order)

	// Wait for cert to be ready
	for i := 0; i < 10 && order.Status != "valid"; i++ {
		time.Sleep(2 * time.Second)
		resp, err = c.post(resp.Header.Get("Location"), nil, true)
		if err != nil {
			break
		}
		json.NewDecoder(resp.Body).Decode(&order)
		resp.Body.Close()
	}

	if order.Certificate == "" {
		return "", fmt.Errorf("no certificate URL (status: %s)", order.Status)
	}

	return order.Certificate, nil
}

func (c *Client) getCertificate(certURL string) (certPEM, keyPEM []byte, err error) {
	resp, err := c.post(certURL, nil, true)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	certPEM, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	keyPath := filepath.Join(certsDir(), c.domain+"-key.pem")
	keyPEM, err = os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, err
	}

	return certPEM, keyPEM, nil
}

func (c *Client) thumbprint() string {
	pub := &c.accountKey.PublicKey
	pubBytes := elliptic.Marshal(pub.Curve, pub.X, pub.Y)
	hash := sha256.Sum256(pubBytes)
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func (c *Client) getNonce() error {
	resp, err := c.http.Head(c.dir["newNonce"])
	if err != nil {
		return err
	}
	c.nonce = resp.Header.Get("Replay-Nonce")
	return nil
}

func (c *Client) post(url string, payload interface{}, useAccount bool) (*http.Response, error) {
	if c.nonce == "" {
		if err := c.getNonce(); err != nil {
			return nil, err
		}
	}

	var payloadBytes []byte
	if payload != nil {
		payloadBytes, _ = json.Marshal(payload)
	}

	protected := map[string]interface{}{
		"alg":   "ES256",
		"nonce": c.nonce,
		"url":   url,
	}
	if useAccount {
		protected["kid"] = c.accountURL
	} else {
		protected["jwk"] = c.jwk()
	}

	protectedB64 := base64.RawURLEncoding.EncodeToString(must(protected))
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
	input := protectedB64 + "." + payloadB64

	hash := sha256.Sum256([]byte(input))
	sig, err := ecdsa.SignASN1(rand.Reader, c.accountKey, hash[:])
	if err != nil {
		return nil, err
	}

	jws := map[string]string{
		"protected": protectedB64,
		"payload":   payloadB64,
		"signature": base64.RawURLEncoding.EncodeToString(sig),
	}

	body, _ := json.Marshal(jws)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/jose+json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	if n := resp.Header.Get("Replay-Nonce"); n != "" {
		c.nonce = n
	}

	return resp, nil
}

func (c *Client) jwk() map[string]interface{} {
	pub := &c.accountKey.PublicKey
	return map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   base64.RawURLEncoding.EncodeToString(pub.X.Bytes()),
		"y":   base64.RawURLEncoding.EncodeToString(pub.Y.Bytes()),
	}
}

func must(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

// LoadOrGenerateCert loads existing cert or generates a new one
func LoadOrGenerateCert(cfToken, cfZone, domain string) (certPath, keyPath string, err error) {
	dir := certsDir()
	certPath = filepath.Join(dir, domain+".pem")
	keyPath = filepath.Join(dir, domain+"-key.pem")

	if data, err := os.ReadFile(certPath); err == nil {
		block, _ := pem.Decode(data)
		if block != nil {
			if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
				if time.Until(cert.NotAfter) > 30*24*time.Hour {
					return certPath, keyPath, nil
				}
			}
		}
	}

	return NewClient(cfToken, cfZone, domain).Generate()
}
