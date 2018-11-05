package main

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/pkg/errors"
)

var remote = "https://billglover-golang.appspot.com/ip"

func main() {

	curIP, err := getIP(remote)
	if err != nil {
		log.Println("unable to get external IP:", err)
	}
	log.Println("external IP address", curIP.String())

	err = updateDNS(curIP, "myip.billglover.me", "Z1EH888BF5XP0N")
	if err != nil {
		log.Println(err)
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			ip, err := getIP(remote)
			if err != nil {
				log.Println("unable to get external IP:", err)
			}

			if curIP.Equal(ip) == false {
				log.Println("external IP address changed", ip.String())

				err = updateDNS(curIP, "myip.billglover.me", "Z1EH888BF5XP0N")
				if err != nil {
					log.Println(err)
				}

				continue
			}
		}
	}
}

func getIP(service string) (net.IP, error) {
	req, err := http.NewRequest(http.MethodGet, remote, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not create request")
	}

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "could not make request")
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read response body")
	}

	ip := net.ParseIP(strings.TrimSpace(string(body)))
	if ip == nil {
		return nil, errors.New("no IP returned from external service")
	}

	return ip, nil
}

// UpdateDNS updates a DNS record to point to the provided IP address.
func updateDNS(ip net.IP, name string, zoneID string) error {

	svc := route53.New(session.New())
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(name),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(ip.String()),
							},
						},
						TTL:  aws.Int64(60),
						Type: aws.String("A"),
					},
				},
			},
			Comment: aws.String("updated by localname"),
		},
		HostedZoneId: aws.String(zoneID),
	}

	_, err := svc.ChangeResourceRecordSets(input)
	if err != nil {
		return errors.Wrap(err, "unable to update record set")
	}

	// TODO: result of svc.ChangeResourceRecordSets is a request
	// identifier. Need to decide if we want to watch for this
	// request to complete successfully, or just log the fact
	// that we have submitted the request.

	return nil
}
