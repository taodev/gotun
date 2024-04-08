package gotun

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

type Listable[T any] []T

func (l Listable[T]) MarshalJSON() ([]byte, error) {
	arrayList := []T(l)
	if len(arrayList) == 1 {
		return yaml.Marshal(arrayList[0])
	}
	return yaml.Marshal(arrayList)
}

func (l *Listable[T]) UnmarshalJSON(content []byte) error {
	err := yaml.Unmarshal(content, (*[]T)(l))
	if err == nil {
		return nil
	}
	var singleItem T
	newError := yaml.Unmarshal(content, &singleItem)
	if newError != nil {
		return newError
	}
	*l = []T{singleItem}
	return nil
}

type SSHOutboundOptions struct {
	ServerAddr           string           `yaml:"server_addr"`
	User                 string           `yaml:"user,omitempty"`
	Password             string           `yaml:"password,omitempty"`
	PrivateKey           string           `yaml:"private_key,omitempty"`
	PrivateKeyPath       string           `yaml:"private_key_path,omitempty"`
	PrivateKeyPassphrase string           `yaml:"private_key_passphrase,omitempty"`
	HostKey              Listable[string] `yaml:"host_key,omitempty"`
	HostKeyAlgorithms    Listable[string] `yaml:"host_key_algorithms,omitempty"`
	ClientVersion        string           `yaml:"client_version,omitempty"`
}

type SSH struct {
	ctx          context.Context
	client       *ssh.Client
	clientAccess sync.Mutex

	serverAddr        string
	user              string
	hostKey           []ssh.PublicKey
	hostKeyAlgorithms []string
	clientVersion     string
	authMethod        []ssh.AuthMethod
}

func NewSSH(ctx context.Context, options SSHOutboundOptions) (*SSH, error) {
	outbound := &SSH{
		ctx:               ctx,
		serverAddr:        options.ServerAddr,
		user:              options.User,
		hostKeyAlgorithms: options.HostKeyAlgorithms,
		clientVersion:     options.ClientVersion,
	}

	if outbound.user == "" {
		outbound.user = "root"
	}

	if outbound.clientVersion == "" {
		outbound.clientVersion = "SSH-2.0-OpenSSH_"
		if rand.Intn(2) == 0 {
			outbound.clientVersion += "7." + strconv.Itoa(rand.Intn(10))
		} else {
			outbound.clientVersion += "8." + strconv.Itoa(rand.Intn(9))
		}
	}

	if options.Password != "" {
		outbound.authMethod = append(outbound.authMethod, ssh.Password(options.Password))
	}

	if len(options.PrivateKey) > 0 || options.PrivateKeyPath != "" {
		var privateKey []byte
		if len(options.PrivateKey) > 0 {
			privateKey = []byte(options.PrivateKey)
		} else {
			var err error
			privateKey, err = os.ReadFile(os.ExpandEnv(options.PrivateKeyPath))
			if err != nil {
				return nil, errors.New("read private key")
			}
		}
		var signer ssh.Signer
		var err error
		if options.PrivateKeyPassphrase == "" {
			signer, err = ssh.ParsePrivateKey(privateKey)
		} else {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(options.PrivateKeyPassphrase))
		}
		if err != nil {
			return nil, errors.New("parse private key")
		}
		outbound.authMethod = append(outbound.authMethod, ssh.PublicKeys(signer))
	}

	if len(options.HostKey) > 0 {
		for _, hostKey := range options.HostKey {
			key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(hostKey))
			if err != nil {
				return nil, errors.New(fmt.Sprint("parse host key ", key))
			}
			outbound.hostKey = append(outbound.hostKey, key)
		}
	}
	return outbound, nil
}

func (s *SSH) connect() (*ssh.Client, error) {
	if s.client != nil {
		return s.client, nil
	}

	s.clientAccess.Lock()
	defer s.clientAccess.Unlock()

	if s.client != nil {
		return s.client, nil
	}

	conf := &ssh.ClientConfig{
		User:              s.user,
		Auth:              s.authMethod,
		ClientVersion:     s.clientVersion,
		HostKeyAlgorithms: s.hostKeyAlgorithms,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			if len(s.hostKey) == 0 {
				return nil
			}
			serverKey := key.Marshal()
			for _, hostKey := range s.hostKey {
				if bytes.Equal(serverKey, hostKey.Marshal()) {
					return nil
				}
			}
			return errors.New("host key mismatch, server send " + key.Type() + " " + base64.StdEncoding.EncodeToString(serverKey))
		},
	}

	var err error
	if s.client, err = ssh.Dial("tcp", s.serverAddr, conf); err != nil {
		return s.client, err
	}

	go func() {
		s.client.Wait()
		s.client.Close()

		s.clientAccess.Lock()
		s.client = nil
		s.clientAccess.Unlock()
	}()

	return s.client, nil
}

func (s *SSH) Close() error {
	if s.client != nil {
		return s.client.Close()
	}

	return nil
}

func (s *SSH) Dial(ctx context.Context, network string, addr string) (net.Conn, error) {
	client, err := s.connect()
	if err != nil {
		return nil, err
	}

	return client.Dial(network, addr)
}
