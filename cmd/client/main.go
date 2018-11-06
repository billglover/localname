package main

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var remote = "https://billglover-golang.appspot.com/ip"

func main() {

	// get environment configuration
	domain := mustGetenv("LOCALNAME_DOMAIN")
	zoneID := mustGetenv("LOCALNAME_ZONE_ID")
	awsAccessKey := mustGetenv("AWS_ACCESS_KEY_ID")
	awsAccessSecret := mustGetenv("AWS_SECRET_ACCESS_KEY")
	pollFreq := mustGetenv("LOCALNAME_POLL_FREQ")

	dur, err := time.ParseDuration(pollFreq)
	if err != nil {
		log.Fatal("unable to parse LOCALNAME_POLL_FREQ:", pollFreq)
	}

	// start monitoring for changes in IP
	cancel, err := start(domain, zoneID, awsAccessKey, awsAccessSecret, dur)
	defer cancel()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Start monitors the external IP address and updates the DNS record provided
// if the IP address changes. It makes an initial attempt to determine the
// IP address and update the domain name. If it encounters an error in this
// initial attempt it returns an error. If the initial attempt is successful
// it continues to monitor for changes in the external IP address. Monitoring
// can be cancelled by calling the cancel function. All errors encountered
// after the initial attempt are deemed retryable and logged. They are not
// returned to the caller.
func start(domain, zoneID, awsKey, awsSecret string, d time.Duration) (context.CancelFunc, error) {

	ctx, cancel := context.WithCancel(context.Background())

	// Get the current IP address so that we have something to compare against.
	curIP, err := getIP(remote)
	if err != nil {
		return cancel, errors.Wrap(err, "unable to get external IP")
	}

	// TODO: need to pass in AWS credentials
	err = updateDNS(curIP, domain, zoneID)
	if err != nil {
		return cancel, errors.Wrap(err, "unable to update DNS record")
	}

	go func(ctx context.Context, curIP net.IP) {

		ticker := time.NewTicker(d)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:

				ip, err := getIP(remote)
				if err != nil {
					log.Println("unable to get external IP:", err)
					continue
				}

				if curIP.Equal(ip) == false {
					log.Println("external IP address changed", ip.String())
					err = updateDNS(ip, domain, zoneID)
					if err != nil {
						log.Println("unable to update DNS:", err)
						continue
					}
					curIP = ip
				}
				log.Println("external IP address unchanged", ip.String())
			}
		}
	}(ctx, curIP)
	return cancel, nil
}

func getIP(service string) (net.IP, error) {
	req, err := http.NewRequest(http.MethodGet, remote, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not create request")
	}

	client := http.Client{
		Timeout: 5 * time.Second,
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
	log.Println("updating DNS:", name, zoneID, ip.String())

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

	return nil
}

func mustGetenv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatal(key, " is not set")
	}
	return val
}
