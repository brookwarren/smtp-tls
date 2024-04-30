package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: smtp-tls hostname port [--debug]")
		os.Exit(1)
	}

	hostname := os.Args[1]
	port := os.Args[2]
	debug := len(os.Args) > 3 && os.Args[3] == "--debug"

	conn, err := net.Dial("tcp", hostname+":"+port)
	if err != nil {
		fmt.Printf("Failed to connect to %s:%s: %v\n", hostname, port, err)
		os.Exit(1)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, hostname)
	if err != nil {
		fmt.Printf("Failed to create SMTP client: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	if err := client.Hello("localhost"); err != nil {
		fmt.Printf("Failed to send EHLO: %v\n", err)
		os.Exit(1)
	}

	if ok, _ := client.Extension("STARTTLS"); !ok {
		fmt.Println("STARTTLS is not supported")
		os.Exit(1)
	}

	if err := client.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		fmt.Printf("Failed to start TLS: %v\n", err)
		os.Exit(1)
	}

	state, ok := client.TLSConnectionState()
	if !ok {
		fmt.Println("Failed to get TLS connection state")
		os.Exit(1)
	}

	cert := state.PeerCertificates[0]

	if debug {
		fmt.Printf("CN: %s\n", cert.Subject.CommonName)
		fmt.Printf("SANs: %v\n", cert.DNSNames)
		fmt.Printf("Expiration: %s\n", cert.NotAfter.Format(time.RFC3339))
	} else {
		remainingDays := int(time.Until(cert.NotAfter).Hours() / 24)
		fmt.Printf("%d\n", remainingDays)
	}
}
