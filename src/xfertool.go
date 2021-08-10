package main

import (
	"io"
	"net"
	"fmt"
	"time"
	"bufio"
	"bytes"
	"strings"
	"strconv"

	"io/ioutil"
	"crypto/tls"
	"crypto/rand"

        "gopkg.in/yaml.v2"
)

type DomConfig struct {
	FullChainCertPath string `yaml:"FullChainCertPath"`
	PrivKeyCertPath string `yaml:"PrivKeyCertPath"`
	DefaultPort string `yaml:"DefaultPort"`
	Domains map[string]string `yaml:"Domains"`
}

var dconfig DomConfig

func forward(conn net.Conn) {
	header := ""
	host := ""
	portnum := ""
	contentl := 0
	bufReader := bufio.NewReader(conn)

	for {
		bytes, err := bufReader.ReadBytes('\n')
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		
		strbytes := string(bytes[:])
		if strings.HasPrefix(strbytes, "Host") {
			parts := strings.Split(strbytes, ":")
			host = strings.TrimSpace(parts[1])
		}

		if strings.HasPrefix(strings.ToLower(strbytes), "content-length") {
			parts := strings.Split(strbytes, ":")
			contentls := strings.TrimSpace(parts[1])
			contentl, err = strconv.Atoi(contentls)
			if err != nil {
				fmt.Println("Unable To Get Content-Length: " + err.Error())
				return
			}
		}

		if strings.TrimSpace(strbytes) == "" {
			break
		}

		header += strbytes
	}	

	bufx := make([]byte, contentl)

	_, _ = bufReader.Read(bufx)

	buf := new(bytes.Buffer)
	buf.WriteString(header)
	buf.WriteString("X-Forwarded-For: " + conn.RemoteAddr().String())
	buf.WriteString("\n")
	buf.WriteString("\r\n")
	buf.Write(bufx)

	for domain, port := range dconfig.Domains {
		if strings.ToLower(host) == strings.ToLower(domain) {
			portnum = port
		}
	}

	if portnum == "" {
		portnum = dconfig.DefaultPort
	}

	if ! strings.HasPrefix(portnum, ":") {
		portnum = ":" + portnum
	}

	fmt.Println("dialing: " + portnum)
	client, err := net.Dial("tcp", portnum)
	if err != nil {
		fmt.Println("error dialing: " + err.Error())
		return 
	}

	go func() {
		defer client.Close()
		defer conn.Close()
		io.CopyN(client, buf, int64(buf.Len()))
		io.Copy(client, conn)
	}()

	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(conn, client)
	}()
}

func main() {
	etcflag := false
	b, err := ioutil.ReadFile("/etc/xfertool/xfertool.yml")
	if err != nil {
		etcflag = true
	}

	if etcflag == true {
		b, err = ioutil.ReadFile("xfertool.yml")
		if err != nil {
			fmt.Println("Error Reading Config File From Two Locations (/etc/xfertool/xfertool.yml) And (Current Directory): " + err.Error())
			return
		}
	}

	if len(b) == 0 {
		fmt.Println("Failed To Load A Config!")
		return
	}

	yml := string(b)

        err = yaml.Unmarshal([]byte(yml), &dconfig)

	cert, err := tls.LoadX509KeyPair(dconfig.FullChainCertPath, dconfig.PrivKeyCertPath) 
	if err != nil {
		fmt.Println("Failed to load keypair: " + err.Error())
		return
	}

	now := time.Now()
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth: tls.RequestClientCert,
	}

	config.Time = func() time.Time { return now }
	config.Rand = rand.Reader

	listener, err := tls.Listen("tcp", ":443", &config)
	if err != nil {
		fmt.Println("Failed to listen: " + err.Error())
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed To Accept Connection: " + err.Error())
			return
		}

		go forward(conn)

	}
}
