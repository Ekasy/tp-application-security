package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var lock = &sync.Mutex{}

const (
	PERMISSIONS               = 0644
	CERTIFICATE_TYPE          = "CERTIFICATE"
	KEY_TYPE                  = "ECDSA PRIVATE KEY"
	MAIN_CERTIFICATE_FILENAME = "certs/ca.crt"
	MAIN_KEY_FILENAME         = "certs/ca.pem"
	MAX_LEAF_LEAVE_TIME       = 24 * time.Hour
	MAX_MAIN_CERT_LEAVE_TIME  = 3650
	KEY_USAGE                 = x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageKeyAgreement | x509.KeyUsageDataEncipherment |
		x509.KeyUsageKeyAgreement | x509.KeyUsageContentCommitment | x509.KeyUsageDigitalSignature
)

type Certificate struct {
	tlsCert tls.Certificate
}

var singleCert *Certificate = nil

func GetSertificate() *Certificate {
	if singleCert == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleCert == nil {
			if err := CreateMainCertificate(); err != nil {
				return nil
			}
		}
	}
	singleCert.tlsCert.Leaf, _ = x509.ParseCertificate(singleCert.tlsCert.Certificate[0])
	return singleCert
}

func randInt(size uint) *big.Int {
	num, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), size))
	if err != nil {
		logrus.Errorln("[randInt]: cannot generate random int")
		return nil
	}
	return num
}

func CreateMainCertificate() error {
	key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		logrus.Error("[CreateMainCertificate]: cannot generate key")
		return err
	}

	serial := randInt(256)
	host, err := os.Hostname()
	if err != nil {
		logrus.Error("[CreateMainCertificate]: cannot get host")
		return err
	}

	cert := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: host},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(MAX_MAIN_CERT_LEAVE_TIME),
		KeyUsage:              KEY_USAGE,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
		SignatureAlgorithm:    x509.ECDSAWithSHA512,
	}

	certFile, err := x509.CreateCertificate(rand.Reader, cert, cert, key.Public(), key)
	if err != nil {
		logrus.Error("[CreateMainCertificate]: cannot create MAIN certificate", err.Error())
		return err
	}

	keyFile, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		logrus.Error("[CreateMainCertificate]: cannot create private key")
		return err
	}

	certPemFile := pem.EncodeToMemory(&pem.Block{
		Type:  CERTIFICATE_TYPE,
		Bytes: certFile,
	})

	keyPemFile := pem.EncodeToMemory(&pem.Block{
		Type:  KEY_TYPE,
		Bytes: keyFile,
	})

	cert_, err := tls.X509KeyPair(certPemFile, keyPemFile)
	if err != nil {
		return err
	}

	singleCert = &Certificate{
		tlsCert: cert_,
	}
	err = ioutil.WriteFile(MAIN_CERTIFICATE_FILENAME, certPemFile, PERMISSIONS)
	if err == nil {
		err = ioutil.WriteFile(MAIN_KEY_FILENAME, keyPemFile, PERMISSIONS)
	}
	return err
}

func CreateLeafCertificate(hosts ...string) (*tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		logrus.Errorln("[CreateLeafCertificate]: cannot generate key")
		return nil, err
	}

	serial := randInt(256)
	cert := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: hosts[0]},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(MAX_LEAF_LEAVE_TIME),
		KeyUsage:              KEY_USAGE,
		BasicConstraintsValid: true,
		SignatureAlgorithm:    x509.ECDSAWithSHA512,
	}

	for _, host := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			cert.IPAddresses = append(cert.IPAddresses, ip)
		} else {
			cert.DNSNames = append(cert.DNSNames, host)
		}
	}

	certFile, err := x509.CreateCertificate(rand.Reader, cert, singleCert.tlsCert.Leaf, key.Public(), singleCert.tlsCert.PrivateKey)
	if err != nil {
		logrus.Errorln("[CreateLeafCertificate]: cannot create cert file")
		return nil, err
	}

	certLeafFile, err := x509.ParseCertificate(certFile)
	if err != nil {
		logrus.Errorln("[CreateLeafCertificate]: cannot create cert LEAF file")
		return nil, err
	}

	tlsCert := &tls.Certificate{
		Certificate: [][]byte{certFile},
		PrivateKey:  key,
		Leaf:        certLeafFile,
	}
	return tlsCert, nil
}
