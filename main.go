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
		GeoNameID uint              `maxminddb:"geoname_id"`
		Names     map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Continent struct {
		Code      string            `maxminddb:"code"`
		GeoNameID uint              `maxminddb:"geoname_id"`
		Names     map[string]string `maxminddb:"names"`
	} `maxminddb:"continent"`
	Country struct {
		GeoNameID uint              `maxminddb:"geoname_id"`
		IsoCode   string            `maxminddb:"iso_code"`
		Names     map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
	Location struct {
		AccuracyRadius uint16  `maxminddb:"accuracy_radius"`
		Latitude       float64 `maxminddb:"latitude"`
		Longitude      float64 `maxminddb:"longitude"`
		MetroCode      uint    `maxminddb:"metro_code"`
		TimeZone       string  `maxminddb:"time_zone"`
	} `maxminddb:"location"`
	Postal struct {
		Code string `maxminddb:"code"`
	} `maxminddb:"postal"`
	RegisteredCountry struct {
		GeoNameID uint              `maxminddb:"geoname_id"`
		IsoCode   string            `maxminddb:"iso_code"`
		Names     map[string]string `maxminddb:"names"`
	} `maxminddb:"registered_country"`
	RepresentedCountry struct {
		GeoNameID uint              `maxminddb:"geoname_id"`
		IsoCode   string            `maxminddb:"iso_code"`
		Names     map[string]string `maxminddb:"names"`
		Type      string            `maxminddb:"type"`
	} `maxminddb:"represented_country"`
	Subdivisions []struct {
		GeoNameID uint              `maxminddb:"geoname_id"`
		IsoCode   string            `maxminddb:"iso_code"`
		Names     map[string]string `maxminddb:"names"`
	} `maxminddb:"subdivisions"`
	Traits struct {
		IsAnonymousProxy    bool `maxminddb:"is_anonymous_proxy"`
		IsSatelliteProvider bool `maxminddb:"is_satellite_provider"`
	} `maxminddb:"traits"`
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
