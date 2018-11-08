# localname

Update Route53 whenever your home IP address changes.

Application behaviour is configured through environment variables. These must
be set or the application will terminate.

* `LOCALNAME_DOMAIN` is the domain name you'd like to update
* `LOCALNAME_ZONE_ID` is the Hosed Zone ID for the domain you'd like to update
* `LOCALNAME_POLL_FREQ` is the frequency with which you'd like to poll for
  changes
* `AWS_ACCESS_KEY_ID` your AWS access key
* `AWS_SECRET_ACCESS_KEY` your AWS access secret

The polling frequency accepts any valid time duration according to Go's
[`time.ParseDuration()`](https://golang.org/pkg/time/#ParseDuration) function.

> A duration string is a possibly signed sequence of decimal numbers, each
> with optional fraction and a unit suffix, such as "300ms", "-1.5h" or
> "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

We would recomend only using positive intervals and picking something in the
minutes or hours range. Anything more frequent than that is unlikely to be
useful.

For more information on obtaining your AWS credentials, the Amazon AWS
documentation has an excelent guide: [Managing Access Keys for your AWS Root
User
Account](https://docs.aws.amazon.com/general/latest/gr/managing-aws-access-keys.html).

