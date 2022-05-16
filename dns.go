package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/meshify-app/meshify/model"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

var (
	ServerTable map[string]string
	ServerLock  sync.Mutex
	DnsTable    map[string][]string
	DnsLock     sync.Mutex
)

func StartDNS() error {
	ServerTable = make(map[string]string)
	DnsTable = make(map[string][]string)

	dns.HandleFunc(".", handleQueries)

	var conf []byte
	for exists := false; !exists; {
		file, err := os.Open(GetDataPath() + "meshify.conf")
		if err != nil {
			time.Sleep(time.Second)
		} else {
			conf, err = ioutil.ReadAll(file)
			file.Close()
			if err != nil {
				log.Errorf("Error reading meshify config file: %v", err)
				time.Sleep(time.Second)
			} else {
				exists = true
			}
		}
	}

	var msg model.Message
	err := json.Unmarshal(conf, &msg)
	if err != nil {
		log.Errorf("Error reading message from config file")
		return err
	}

	ServerLock.Lock()
	defer ServerLock.Unlock()

	DnsLock.Lock()
	defer DnsLock.Unlock()

	for i := 0; i < len(msg.Config); i++ {
		index := -1
		for j := 0; j < len(msg.Config[i].Hosts); j++ {
			if msg.Config[i].Hosts[j].HostGroup == config.HostID {
				index = j
				break
			}
		}
		if index == -1 {
			log.Errorf("Error reading message %v", msg)
		} else {
			host := msg.Config[i].Hosts[index]
			name := strings.ToLower(host.Name)
			DnsTable[name] = append(DnsTable[name], host.Current.Address...)
			msg.Config[i].Hosts = append(msg.Config[i].Hosts[:index], msg.Config[i].Hosts[index+1:]...)
			for j := 0; j < len(msg.Config[i].Hosts); j++ {
				n := strings.ToLower(msg.Config[i].Hosts[j].Name)
				DnsTable[n] = append(DnsTable[n], msg.Config[i].Hosts[j].Current.Address...)
				if msg.Config[i].Hosts[j].Current.Endpoint != "" {
					ip_port := msg.Config[i].Hosts[j].Current.Endpoint
					parts := strings.Split(ip_port, ":")
					ip := parts[0]
					ServerTable[ip] = ip
				}
			}

			if len(host.Current.Address[0]) > 3 {
				address := host.Current.Address[0][:len(host.Current.Address[0])-3] + ":53"
				server := &dns.Server{Addr: address, Net: "udp", TsigSecret: nil, ReusePort: true}
				log.Infof("Starting DNS Server on %s", address)
				go func() {
					if err := server.ListenAndServe(); err != nil {
						log.Errorf("Failed to setup the DNS server on %s: %s\n", address, err.Error())
					}
				}()
			}
		}
	}

	return nil
}

func UpdateDNS(msg model.Message) error {

	serverTable := make(map[string]string)
	dnsTable := make(map[string][]string)

	for i := 0; i < len(msg.Config); i++ {
		index := -1
		for j := 0; j < len(msg.Config[i].Hosts); j++ {
			if msg.Config[i].Hosts[j].HostGroup == config.HostID {
				index = j
				break
			}
		}
		if index == -1 {
			log.Errorf("Error reading message for DNS update: %v", msg)
			return errors.New("Error reading message")
		} else {
			host := msg.Config[i].Hosts[index]
			name := strings.ToLower(host.Name)
			dnsTable[name] = append(dnsTable[name], host.Current.Address...)
			msg.Config[i].Hosts = append(msg.Config[i].Hosts[:index], msg.Config[i].Hosts[index+1:]...)
			for j := 0; j < len(msg.Config[i].Hosts); j++ {
				n := strings.ToLower(msg.Config[i].Hosts[j].Name)
				dnsTable[n] = append(dnsTable[n], msg.Config[i].Hosts[j].Current.Address...)
				if msg.Config[i].Hosts[j].Current.Endpoint != "" {
					ip_port := msg.Config[i].Hosts[j].Current.Endpoint
					parts := strings.Split(ip_port, ":")
					ip := parts[0]
					serverTable[ip] = ip
				}
			}
		}
	}
	DnsLock.Lock()
	DnsTable = dnsTable
	DnsLock.Unlock()

	ServerLock.Lock()
	ServerTable = serverTable
	ServerLock.Unlock()

	return nil
}

func handleQueries(w dns.ResponseWriter, r *dns.Msg) {
	var rr dns.RR
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = true

	q := strings.ToLower(r.Question[0].Name)
	q = strings.Trim(q, ".")

	log.Infof("DNS Query: %s", q)
	addrs := DnsTable[q]
	if addrs == nil {
		m.Rcode = dns.RcodeServerFailure
	} else {
		if r.Question[0].Qtype == dns.TypeA {
			offset := rand.Intn(len(addrs))
			for i := 0; i < len(addrs); i++ {
				x := (offset + i) % len(addrs)
				if !strings.Contains(addrs[x], ":") {
					ip, _, _ := net.ParseCIDR(addrs[x])
					rr = &dns.A{Hdr: dns.RR_Header{Name: r.Question[0].Name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    300},
						A: ip.To4(),
					}
					m.Answer = append(m.Answer, rr)
					m.Rcode = dns.RcodeSuccess
				}
			}
		}
		if r.Question[0].Qtype == dns.TypeAAAA {
			offset := rand.Intn(len(addrs))
			for i := 0; i < len(addrs); i++ {
				x := (offset + i) % len(addrs)
				if strings.Contains(addrs[x], ":") {
					ip, _, _ := net.ParseCIDR(addrs[x])
					rr = &dns.A{Hdr: dns.RR_Header{Name: r.Question[0].Name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    300},
						A: ip.To4(),
					}
					m.Answer = append(m.Answer, rr)
					m.Rcode = dns.RcodeSuccess
				}
			}
		}
	}

	w.WriteMsg(m)
	go LogMessage(q)

}

// This sends a multicast message with the DNS query to anyone listening
func LogMessage(query string) {

	raddr, err := net.ResolveUDPAddr("udp", "224.1.1.1:53281")
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return
	}
	defer conn.Close()

	conn.WriteMsgUDP([]byte(query), nil, raddr)

	fmt.Fprint(conn, query)

}
