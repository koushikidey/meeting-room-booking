package cache

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/koushikidey/go-meetingroombook/pkg/models"
	"github.com/patrickmn/go-cache"
)

type allCache struct {
	employees *cache.Cache
}

const (
	defaultExpirationTime = 5 * time.Minute
	purgeTime             = 10 * time.Minute
)

var C *allCache

func InitCache() {
	C = &allCache{
		employees: cache.New(defaultExpirationTime, purgeTime),
	}
}

func (c *allCache) Read(id uint) ([]byte, bool) {
	employee, ok := c.employees.Get(strconv.FormatUint(uint64(id), 10))
	if ok {
		log.Println("data fetched from cache")
		res, err := json.Marshal(employee.(models.Employee))
		if err != nil {
			log.Println("error in cache marshal:", err)
			return nil, false
		}
		return res, true
	}
	return nil, false
}

func (c *allCache) Update(id uint, employee models.Employee) {
	c.employees.Set(strconv.FormatUint(uint64(id), 10), employee, cache.DefaultExpiration)
}