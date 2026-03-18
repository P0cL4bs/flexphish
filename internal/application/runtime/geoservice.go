package runtime

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

type GeoService struct {
	db *geoip2.Reader
}

func NewGeoService(dbPath string) (*GeoService, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, err
	}

	return &GeoService{db: db}, nil
}

func (s *GeoService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *GeoService) Lookup(ipStr string) (country string, city string) {
	if s == nil || s.db == nil {
		return "", ""
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", ""
	}

	record, err := s.db.City(ip)
	if err != nil {
		return "", ""
	}

	if record.Country.Names != nil {
		country = record.Country.Names["en"]
	}

	if record.City.Names != nil {
		city = record.City.Names["en"]
	}

	return
}
