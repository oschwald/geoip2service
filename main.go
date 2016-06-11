package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/oschwald/maxminddb-golang"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

var (
	addr     = flag.String("addr", ":8080", "TCP address to listen to")
	compress = flag.Bool("compress", false, "Whether to enable transparent response compression")
	dbFile   = flag.String("dbFile", "/usr/local/share/GeoIP/GeoLite2-City.mmdb", "Path to the GeoLite2 or GeoIP2 City database")
)

type cachedDb struct {
	reader *maxminddb.Reader
	cache  map[uintptr]City
}

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

func main() {
	flag.Parse()

	cdb, err := createCachedDb()
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	router := routing.New()
	router.Get("/geoip/v2.1/city/<ip>", func(c *routing.Context) error {
		return cityRequestHandler(c, cdb)
	})

	handler := router.HandleRequest
	if *compress {
		handler = fasthttp.CompressHandler(handler)
	}
	if err := fasthttp.ListenAndServe(*addr, handler); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func createCachedDb() (*cachedDb, error) {
	db, err := maxminddb.Open(*dbFile)
	if err != nil {
		return nil, err
	}

	cache := make(map[uintptr]City)

	// Add all data records to the cache so that we don't have to worry about
	// concurrency issues.
	networks := db.Networks()
	for networks.Next() {
		_, offset, err := networks.NetworkOffset()
		if err != nil {
			return nil, err
		}
		if _, ok := cache[offset]; ok {
			continue
		}
		var record City
		err = db.Decode(offset, &record)
		if err != nil {
			return nil, err
		}
		cache[offset] = record

	}
	if networks.Err() != nil {
		return nil, networks.Err()
	}

	return &cachedDb{
		cache:  cache,
		reader: db,
	}, nil
}

func cityRequestHandler(c *routing.Context, cdb *cachedDb) error {
	ip := net.ParseIP(c.Param("ip"))
	if ip == nil {
		c.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return nil
	}

	offset, err := cdb.reader.LookupOffset(ip)
	if err != nil {
		return err
	}
	if offset == maxminddb.NotFound {
		c.Response.SetStatusCode(fasthttp.StatusNotFound)
		return nil
	}

	record, ok := cdb.cache[offset]
	if !ok {
		return fmt.Errorf("offset without a record: %d", offset)
	}
	// not caching JSON as I plan to add things like IP address, etc., to
	// match official web service
	enc := json.NewEncoder(c.Response.BodyWriter())
	enc.Encode(record)

	return nil
}
