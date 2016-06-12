package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/mailru/easyjson"
	"github.com/oschwald/geoip2service/model"
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
	cache  map[uintptr]model.City
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

	cache := make(map[uintptr]model.City)

	// Add all data records to the cache so that we don't have to worry about
	// concurrency issues.
	// XXX - This would be much faster if we just iterated over the data
	// section and not the search tree. The verifier code does this. It relies
	// on implementation details of the MMDB::Writer code, but that is
	// probably fine.
	networks := db.Networks()
	for networks.Next() {
		_, offset, err := networks.NetworkOffset()
		if err != nil {
			return nil, err
		}
		if _, ok := cache[offset]; ok {
			continue
		}
		var record model.City
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
	var ip net.IP
	ipStr := c.Param("ip")
	if ipStr == "me" {
		ip = c.RemoteIP()
	} else {
		ip = net.ParseIP(ipStr)
	}
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
	easyjson.MarshalToWriter(record, c.Response.BodyWriter())

	return nil
}
