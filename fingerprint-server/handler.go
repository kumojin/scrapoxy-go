package main

import (
	"encoding/json"
	"fmt"
	"github.com/oschwald/maxminddb-golang"
	"net"
	"net/http"
)

type City struct {
	City struct {
		Names     map[string]string `maxminddb:"names"`
		GeoNameID uint              `maxminddb:"geoname_id"`
	} `maxminddb:"city"`
	Postal struct {
		Code string `maxminddb:"code"`
	} `maxminddb:"postal"`
	Continent struct {
		Names     map[string]string `maxminddb:"names"`
		Code      string            `maxminddb:"code"`
		GeoNameID uint              `maxminddb:"geoname_id"`
	} `maxminddb:"continent"`
	Subdivisions []struct {
		Names     map[string]string `maxminddb:"names"`
		IsoCode   string            `maxminddb:"iso_code"`
		GeoNameID uint              `maxminddb:"geoname_id"`
	} `maxminddb:"subdivisions"`
	RepresentedCountry struct {
		Names             map[string]string `maxminddb:"names"`
		IsoCode           string            `maxminddb:"iso_code"`
		Type              string            `maxminddb:"type"`
		GeoNameID         uint              `maxminddb:"geoname_id"`
		IsInEuropeanUnion bool              `maxminddb:"is_in_european_union"`
	} `maxminddb:"represented_country"`
	Country struct {
		Names             map[string]string `maxminddb:"names"`
		IsoCode           string            `maxminddb:"iso_code"`
		GeoNameID         uint              `maxminddb:"geoname_id"`
		IsInEuropeanUnion bool              `maxminddb:"is_in_european_union"`
	} `maxminddb:"country"`
	RegisteredCountry struct {
		Names             map[string]string `maxminddb:"names"`
		IsoCode           string            `maxminddb:"iso_code"`
		GeoNameID         uint              `maxminddb:"geoname_id"`
		IsInEuropeanUnion bool              `maxminddb:"is_in_european_union"`
	} `maxminddb:"registered_country"`
	Location struct {
		TimeZone       string  `maxminddb:"time_zone"`
		Latitude       float64 `maxminddb:"latitude"`
		Longitude      float64 `maxminddb:"longitude"`
		MetroCode      uint    `maxminddb:"metro_code"`
		AccuracyRadius uint16  `maxminddb:"accuracy_radius"`
	} `maxminddb:"location"`
	Traits struct {
		IsAnonymousProxy    bool `maxminddb:"is_anonymous_proxy"`
		IsAnycast           bool `maxminddb:"is_anycast"`
		IsSatelliteProvider bool `maxminddb:"is_satellite_provider"`
	} `maxminddb:"traits"`
}

type FingerprintResponse struct {
	IP        string `json:"ip""`
	UserAgent string `json:"useragent"`

	ASNName       string  `json:"asnName"`
	ASNNetwork    string  `json:"asnNetwork"`
	ContinentCode string  `json:"continentCode"`
	ContinentName string  `json:"continentName"`
	CountryCode   string  `json:"countryCode"`
	CountryName   string  `json:"countryName"`
	CityName      string  `json:"cityName"`
	Timezone      string  `json:"timezone"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
}

type FingerprintRecord struct {
	Network   string `maxminddb:"network"`
	Continent struct {
		Code  string            `maxminddb:"code"`
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"continent"`
	Country struct {
		Code  string            `maxminddb:"iso_code"`
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Location struct {
		Timezone  string  `maxminddb:"time_zone"`
		Latitude  float64 `maxminddb:"latitude"`
		Longitude float64 `maxminddb:"longitude"`
	} `maxminddb:"location"`
	Traits struct {
		Network string `maxminddb:"network"`
	} `maxminddb:"traits"`
}

// The ASNRecord struct corresponds to the data in the GeoLite2 ASN database.
type ASNRecord struct {
	AutonomousSystemOrganization string `maxminddb:"autonomous_system_organization"`
	AutonomousSystemNumber       uint   `maxminddb:"autonomous_system_number"`
}

type Handler struct {
	cityDBReader *maxminddb.Reader
	asnDBReader  *maxminddb.Reader
}

// NewHandler creates a new instance of the Handler struct with the given database reader.
func NewHandler(cityDB, asnDB *maxminddb.Reader) *Handler {
	return &Handler{cityDBReader: cityDB, asnDBReader: asnDB}
}

func (h Handler) HandleJsonAPIRequest(w http.ResponseWriter, r *http.Request) {
	requestCounter.Inc()

	clientIp := getUserIP(r)
	response := getUserIpInformation(h.cityDBReader, h.asnDBReader, clientIp, r.UserAgent())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Handle404Request handles a 404 request by returning an HTTP Not Found status code.
func (h Handler) Handle404Request(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

// getUserIP returns the IP address of the user accessing the server. It first checks if
// the X-Forwarded-For header is present in the HTTP request headers. If it is, it parses
// and returns the IP address from this header. If not, it checks the X-Real-IP header
// and returns the IP address from there if it is present. If neither header is present,
// it retrieves the IP address from the RemoteAddr field in the http.Request struct.
func getUserIP(req *http.Request) net.IP {
	if len(req.Header.Get("X-Forwarded-For")) > 1 {
		userIP := req.Header.Get("X-Forwarded-For")
		return net.ParseIP(userIP)
	}

	if len(req.Header.Get("X-Real-IP")) > 1 {
		userIP := req.Header.Get("X-Real-IP")
		return net.ParseIP(userIP)
	}

	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil {
		return net.ParseIP(ip)
	}

	return net.ParseIP(req.RemoteAddr)
}

func getUserIpInformation(cityDBReader, asnDBReader *maxminddb.Reader, clientIp net.IP, userAgent string) FingerprintResponse {
	var toto any
	cityDBReader.Lookup(clientIp, &toto)

	var record FingerprintRecord
	network, _, err := cityDBReader.LookupNetwork(clientIp, &record)
	if err != nil {
		fmt.Println(err)
	}

	response := FingerprintResponse{
		IP:            clientIp.String(),
		UserAgent:     userAgent,
		ASNName:       "",
		ASNNetwork:    network.String(),
		ContinentCode: record.Continent.Code,
		ContinentName: record.Continent.Names["en"],
		CountryCode:   record.Country.Code,
		CountryName:   record.Country.Names["en"],
		CityName:      record.City.Names["en"],
		Timezone:      record.Location.Timezone,
		Latitude:      record.Location.Latitude,
		Longitude:     record.Location.Longitude,
	}

	var asnRecord ASNRecord
	err = asnDBReader.Lookup(clientIp, &asnRecord)
	if err == nil {
		response.ASNName = asnRecord.AutonomousSystemOrganization
	}
	return response
}
