package main

import (
	"github.com/oschwald/maxminddb-golang"
	"net"
	"reflect"
	"testing"
)

func Test_getUserIpInformation(t *testing.T) {
	cityDBReader, err := maxminddb.Open("./GeoLite2-City.mmdb")
	if err != nil {
		t.Errorf("Error opening GeoLite2-City.mmdb: %v", err)
	}

	asnDBReader, err := maxminddb.Open("./GeoLite2-ASN.mmdb")
	if err != nil {
		t.Errorf("Error opening GeoLite2-ASN.mmdb: %v", err)
	}

	type args struct {
		clientIp  net.IP
		userAgent string
	}
	tests := []struct {
		name string
		args args
		want FingerprintResponse
	}{
		{name: "127.0.0.1", args: struct {
			clientIp  net.IP
			userAgent string
		}{clientIp: net.ParseIP("127.0.0.1"), userAgent: "curl/8.6.0"}, want: FingerprintResponse{IP: "127.0.0.1", UserAgent: "curl/8.6.0", ASNNetwork: "127.0.0.0/8"}},
		{name: "::1", args: struct {
			clientIp  net.IP
			userAgent string
		}{clientIp: net.ParseIP("::1"), userAgent: "curl/8.6.0"}, want: FingerprintResponse{IP: "::1", UserAgent: "curl/8.6.0", ASNNetwork: "::/104"}},
		{name: "70.53.250.221", args: struct {
			clientIp  net.IP
			userAgent string
		}{clientIp: net.ParseIP("70.53.250.221"), userAgent: "curl/8.6.0"}, want: FingerprintResponse{
			IP:            "70.53.250.221",
			UserAgent:     "curl/8.6.0",
			ASNName:       "BACOM",
			ASNNetwork:    "70.53.250.0/24",
			ContinentCode: "NA",
			ContinentName: "North America",
			CountryCode:   "CA",
			CountryName:   "Canada",
			CityName:      "Québec",
			Timezone:      "America/Toronto",
			Latitude:      46.8801,
			Longitude:     -71.1927,
		}},
		{name: "2001:4860:4860::8888", args: struct {
			clientIp  net.IP
			userAgent string
		}{clientIp: net.ParseIP("2001:4860:4860::8888"), userAgent: "curl/8.6.0"}, want: FingerprintResponse{
			IP:            "2001:4860:4860::8888",
			UserAgent:     "curl/8.6.0",
			ASNName:       "GOOGLE",
			ASNNetwork:    "2001:4860:4800::/37",
			ContinentCode: "NA",
			ContinentName: "North America",
			CountryCode:   "US",
			CountryName:   "United States",
			CityName:      "",
			Timezone:      "America/Chicago",
			Latitude:      37.751,
			Longitude:     -97.822,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getUserIpInformation(cityDBReader, asnDBReader, tt.args.clientIp, tt.args.userAgent); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getUserIpInformation() = %v, want %v", got, tt.want)
			}
		})
	}
}
