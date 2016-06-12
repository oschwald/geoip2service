package model

type City struct {
	City struct {
		GeoNameID uint              `json:"geoname_id" maxminddb:"geoname_id"`
		Names     map[string]string `json:"names" maxminddb:"names"`
	} `json:"city" maxminddb:"city"`
	Continent struct {
		Code      string            `json:"code" maxminddb:"code"`
		GeoNameID uint              `json:"geoname_id" maxminddb:"geoname_id"`
		Names     map[string]string `json:"names" maxminddb:"names"`
	} `json:"continent" maxminddb:"continent"`
	Country struct {
		GeoNameID uint              `json:"geoname_id" maxminddb:"geoname_id"`
		IsoCode   string            `json:"iso_code" maxminddb:"iso_code"`
		Names     map[string]string `json:"names" maxminddb:"names"`
	} `json:"country" maxminddb:"country"`
	Location struct {
		AccuracyRadius uint16  `json:"accuracy_radius" maxminddb:"accuracy_radius"`
		Latitude       float64 `json:"latitude" maxminddb:"latitude"`
		Longitude      float64 `json:"longitude" maxminddb:"longitude"`
		MetroCode      uint    `json:"metro_code" maxminddb:"metro_code"`
		TimeZone       string  `json:"time_zone" maxminddb:"time_zone"`
	} `json:"location" maxminddb:"location"`
	Postal struct {
		Code string `json:"code" maxminddb:"code"`
	} `json:"postal" maxminddb:"postal"`
	RegisteredCountry struct {
		GeoNameID uint              `json:"geoname_id" maxminddb:"geoname_id"`
		IsoCode   string            `json:"iso_code" maxminddb:"iso_code"`
		Names     map[string]string `json:"names" maxminddb:"names"`
	} `json:"registered_country" maxminddb:"registered_country"`
	RepresentedCountry struct {
		GeoNameID uint              `json:"geoname_id" maxminddb:"geoname_id"`
		IsoCode   string            `json:"iso_code" maxminddb:"iso_code"`
		Names     map[string]string `json:"names" maxminddb:"names"`
		Type      string            `json:"type" maxminddb:"type"`
	} `json:"represented_country" maxminddb:"represented_country"`
	Subdivisions []struct {
		GeoNameID uint              `json:"geoname_id" maxminddb:"geoname_id"`
		IsoCode   string            `json:"iso_code" maxminddb:"iso_code"`
		Names     map[string]string `json:"names" maxminddb:"names"`
	} `json:"subdivisions" maxminddb:"subdivisions"`
	Traits struct {
		IsAnonymousProxy    bool `json:"is_anonymous_proxy" maxminddb:"is_anonymous_proxy"`
		IsSatelliteProvider bool `json:"is_satellite_provider" maxminddb:"is_satellite_provider"`
	} `json:"traits" maxminddb:"traits"`
}
