// Copyright (c) 2019 Daniel Oaks <daniel@danieloaks.net>
// Copyright (c) 2019 Shivaram Lingamneni
// released under the MIT license

package conn

import "crypto/tls"

// insecureTLSConfig is a TLS config that disables cert verification. We use it to run all
// tests, rather than each test having to define it separately.
var insecureTLSConfig *tls.Config

func init() {
	insecureTLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
}
