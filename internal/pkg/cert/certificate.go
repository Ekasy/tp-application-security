package cert

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

type CertificateManager struct {
	rootDir    string
	genCertCmd string
}

var lock = &sync.Mutex{}
var singleCM *CertificateManager = nil

const (
	genCertDdir string = "gen_certificate"
	certsDir    string = "certs"
)

func GetCertificateManager() *CertificateManager {
	if singleCM == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleCM == nil {
			if err := initializeCertificateManager(); err != nil {
				fmt.Printf("[GetCertificateManager] cert manager not initialized: %s", err.Error())
				return nil
			}
		}
	}
	return singleCM
}

func initializeCertificateManager() error {
	rootDir, err := os.Getwd()
	if err != nil {
		return err
	}

	_, err = os.Stat(fmt.Sprintf("%s/%s", rootDir, genCertDdir)) // check genCert folder
	if os.IsNotExist(err) {
		return err
	}

	certs := fmt.Sprintf("%s/%s/%s", rootDir, genCertDdir, certsDir)
	_, err = os.Stat(certs) // check certs folder
	if os.IsNotExist(err) {
		err = os.Mkdir(certs, 0700)
		if os.IsNotExist(err) {
			return err
		}
	}

	singleCM = &CertificateManager{
		rootDir:    rootDir,
		genCertCmd: fmt.Sprintf("%s/%s/gen_cert.sh", rootDir, genCertDdir),
	}

	return nil
}

func (cm *CertificateManager) GenerateCertificate(host string) (*tls.Certificate, error) {
	certificatePath := fmt.Sprintf("%s/%s/%s/%s.crt", cm.rootDir, genCertDdir, certsDir, host)
	_, err := os.Stat(certificatePath)
	if os.IsNotExist(err) {
		command := exec.Command(cm.genCertCmd, host, strconv.Itoa(rand.Intn(1000000)))
		_, err := command.CombinedOutput()
		if err != nil {
			fmt.Printf("[GenerateCertificate] certificate not generated: %s", err.Error())
			return nil, err
		}
	}

	certificate, err := tls.LoadX509KeyPair(certificatePath, fmt.Sprintf("%s/%s/cert.key", cm.rootDir, genCertDdir))
	if err != nil {
		fmt.Printf("[GenerateCertificate] key pair not created: %s", err.Error())
		return nil, err
	}

	return &certificate, nil
}
