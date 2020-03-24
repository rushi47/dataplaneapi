// Copyright 2019 HAProxy Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package configuration

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	client_native "github.com/haproxytech/client-native"
	parser "github.com/haproxytech/config-parser/v2"
	"github.com/haproxytech/config-parser/v2/types"
)

//Node is structure required for connection to cluster
type Node struct {
	Address     string `json:"address"`
	APIBasePath string `json:"api_base_path"`
	APIPassword string `json:"api_password"`
	APIUser     string `json:"api_user"`
	Certificate string `json:"certificate,omitempty"`
	Description string `json:"description,omitempty"`
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Port        int64  `json:"port,omitempty"`
	Status      string `json:"status"`
	Type        string `json:"type"`
}

//ClusterSync fetches certificates for joining cluster
type ClusterSync struct {
	cfg       *Configuration
	certFetch chan struct{}
	cli       *client_native.HAProxyClient
}

func (c *ClusterSync) Monitor(cfg *Configuration, cli *client_native.HAProxyClient) {
	c.cfg = cfg
	c.cli = cli

	go c.monitorBootstrapKey()

	c.certFetch = make(chan struct{}, 2)
	go c.fetchCert()

	key := c.cfg.Cluster.BootstrapKey.Load()
	certFetched := cfg.Cluster.CertFetched.Load()

	if key != "" && !certFetched {
		c.cfg.BotstrapKeyReload()
	}
}

func (c *ClusterSync) monitorBootstrapKey() {
	for range c.cfg.GetBotstrapKeyChange() {
		key := c.cfg.Cluster.BootstrapKey.Load()
		c.cfg.Cluster.CertFetched.Store(false)
		if key == "" {
			//do we need to delete cert here maybe?
			c.cfg.Cluster.ActiveBootstrapKey.Store("")
			err := c.cfg.Save()
			if err != nil {
				log.Panic(err)
			}
			continue
		}
		if key == c.cfg.Cluster.ActiveBootstrapKey.Load() {
			fetched := c.cfg.Cluster.CertFetched.Load()
			if !fetched {
				c.certFetch <- struct{}{}
			}
			continue
		}
		raw, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			log.Println(err)
			continue
		}
		data := strings.Split(string(raw), ":")
		if len(data) != 5 {
			log.Println("bottstrap key in unrecognized format")
			continue
		}
		url := fmt.Sprintf("%s:%s", data[0], data[1])
		c.cfg.Cluster.URL.Store(url)
		c.cfg.Cluster.Port.Store(data[2])
		c.cfg.Cluster.APIBasePath.Store(data[3])
		c.cfg.Cluster.Mode.Store("cluster")
		err = c.cfg.Save()
		if err != nil {
			log.Panic(err)
		}
		csr, key, err := generateCSR()
		if err != nil {
			log.Println(err)
			continue
		}
		err = ioutil.WriteFile("tls.key", []byte(key), 0644)
		if err != nil {
			log.Println(err)
			continue
		}
		err = ioutil.WriteFile("csr.crt", []byte(csr), 0644)
		if err != nil {
			log.Println(err)
			continue
		}
		err = c.cfg.Save()
		if err != nil {
			log.Panic(err)
		}
		err = c.issueJoinRequest(url, data[2], data[3], csr, key)
		if err != nil {
			log.Println(err)
			continue
		}
		c.certFetch <- struct{}{}
	}
}

func (c *ClusterSync) issueJoinRequest(url, port, basePath string, csr, key string) error {
	url = fmt.Sprintf("%s:%s/%s", url, port, basePath)
	serverCfg := c.cfg.Server

	data, err := c.cli.Configuration.Parser.Get(parser.UserList, c.cfg.HAProxy.Userlist, "user")
	if err != nil {
		return fmt.Errorf("error reading userlist %v userlist in conf: %s", c.cfg.HAProxy.Userlist, err.Error())
	}
	users, ok := data.([]types.User)
	if !ok {
		return fmt.Errorf("error reading users from %v userlist in conf", c.cfg.HAProxy.Userlist)
	}
	if len(users) == 0 {
		return fmt.Errorf("no users configured in %v userlist in conf", c.cfg.HAProxy.Userlist)
	}
	var user *types.User
	for _, u := range users {
		if u.IsInsecure {
			user = &u
			break
		}
	}
	if user == nil {
		return fmt.Errorf("no available user for cluster comunication")
	}

	nodeData := Node{
		//ID:          "",
		Address:     serverCfg.Host,
		APIBasePath: serverCfg.APIBasePath,
		APIPassword: user.Name,
		APIUser:     user.Password,
		Certificate: csr,
		Description: "",
		Name:        c.cfg.Cluster.Name.Load(),
		Port:        int64(serverCfg.Port),
		Status:      "waiting_approval",
		Type:        "community",
	}
	bytesRepresentation, _ := json.Marshal(nodeData)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		return fmt.Errorf("error creating new POST request for cluster comunication")
	}
	req.Header.Add("X-Bootstrap-Key", c.cfg.Cluster.BootstrapKey.Load())
	req.Header.Add("Content-Type", "application/json")
	log.Printf("Joining cluster %s", url)
	httpClient := createHTTPClient()
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("status code not proper [%d] %s", resp.StatusCode, string(body))
	}
	var responseData Node
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return err
	}
	c.cfg.Cluster.ID.Store(responseData.ID)
	c.cfg.Cluster.Token.Store(resp.Header.Get("X-Node-Key"))
	c.cfg.Cluster.ActiveBootstrapKey.Store(c.cfg.Cluster.BootstrapKey.Load())
	c.cfg.Cluster.Status.Store(responseData.Status)
	log.Printf("Cluster joined, status: %s", responseData.Status)
	if responseData.Status == "active" {
		err = ioutil.WriteFile("tls.crt", []byte(responseData.Certificate), 0644)
		if err != nil {
			return err
		}
		c.cfg.Cluster.CertificatePath.Store("tls.crt")
		c.cfg.Cluster.CertFetched.Store(true)
		c.cfg.RestartServer()
	}
	err = c.cfg.Save()
	if err != nil {
		return err
	}
	return nil
}

func (c *ClusterSync) activateFetchCert(err error) {
	go func(err error) {
		log.Println(err)
		time.Sleep(1 * time.Minute)
		c.certFetch <- struct{}{}
	}(err)
}

func (c *ClusterSync) fetchCert() {
	for range c.certFetch {
		key := c.cfg.Cluster.BootstrapKey.Load()
		if key == "" || c.cfg.Cluster.Token.Load() == "" {
			continue
		}
		//if not, sleep and start all over again
		certFetched := c.cfg.Cluster.CertFetched.Load()
		if !certFetched {
			url := c.cfg.Cluster.URL.Load()
			port := c.cfg.Cluster.Port.Load()
			aPIBasePath := c.cfg.Cluster.APIBasePath.Load()
			id := c.cfg.Cluster.ID.Load()
			url = fmt.Sprintf("%s:%s/%s/%s", url, port, aPIBasePath, id)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				c.activateFetchCert(err)
				continue
			}
			req.Header.Add("X-Node-Key", c.cfg.Cluster.Token.Load())
			req.Header.Add("Content-Type", "application/json")
			httpClient := createHTTPClient()
			resp, err := httpClient.Do(req)
			if err != nil {
				c.activateFetchCert(err)
				continue
			}
			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				c.activateFetchCert(err)
				continue
			}
			if resp.StatusCode != 200 {
				c.activateFetchCert(fmt.Errorf("status code not proper [%d] %s", resp.StatusCode, string(body)))
				continue
			}
			var responseData Node
			err = json.Unmarshal(body, &responseData)
			if err != nil {
				c.activateFetchCert(err)
				continue
			}
			c.cfg.Cluster.Status.Store(responseData.Status)
			log.Printf("Fetching certificate, status: %s", responseData.Status)

			if responseData.Status == "active" {
				err = ioutil.WriteFile("tls.crt", []byte(responseData.Certificate), 0644)
				if err != nil {
					log.Println(err.Error())
					continue
				}
				c.cfg.Cluster.CertificatePath.Store("tls.crt")
				c.cfg.Cluster.CertFetched.Store(true)
				c.cfg.RestartServer()
			}
			err = c.cfg.Save()
			if err != nil {
				log.Println(err)
			}
		}
		if !certFetched {
			go func() {
				time.Sleep(1 * time.Minute)
				c.certFetch <- struct{}{}
			}()
		}
	}
}

func generateCSR() (string, string, error) {
	keyBytes, _ := rsa.GenerateKey(rand.Reader, 2048)

	oidEmailAddress := asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}
	emailAddress := "test@example.com"
	subj := pkix.Name{
		CommonName:         "haproxy.com",
		Country:            []string{"US"},
		Province:           []string{""},
		Locality:           []string{"Waltham"},
		Organization:       []string{"HAProxy Technologies LLC"},
		OrganizationalUnit: []string{"IT"},
		ExtraNames: []pkix.AttributeTypeAndValue{
			{
				Type: oidEmailAddress,
				Value: asn1.RawValue{
					Tag:   asn1.TagIA5String,
					Bytes: []byte(emailAddress),
				},
			},
		},
	}

	template := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}
	csrBytes, _ := x509.CreateCertificateRequest(rand.Reader, &template, keyBytes)
	var buf bytes.Buffer
	err := pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
	if err != nil {
		return "", "", err
	}

	caPrivKeyPEMBuff := new(bytes.Buffer)
	err = pem.Encode(caPrivKeyPEMBuff, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(keyBytes),
	})
	if err != nil {
		return "", "", err
	}
	return buf.String(), caPrivKeyPEMBuff.String(), err
}

func createHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 20,
			TLSClientConfig: &tls.Config{
				//nolint
				InsecureSkipVerify: true, //this is deliberate, might only have self signed certificate
			},
		},
		Timeout: time.Duration(10) * time.Second,
	}
	return client
}
