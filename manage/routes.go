package manage

import (
	"fmt"
	"strings"

	"github.com/lovego/goa"
)

func Routes(router *goa.RouterGroup) {
	router.Get(`/caches`, func(c *goa.Context) {
		c.Write(List())
	})

	router.Get(`/caches/([^/]+)/([^/]+)/([^/]+)`, func(c *goa.Context) {
		if data, err := Detail(
			c.Param(0), c.Param(1), c.Param(2), c.URL.Query().Get("keys"),
		); err == nil {
			c.Json(data)
		} else {
			c.Write([]byte(err.Error()))
		}
	})

	router.Get(`/caches/([^/]+)/([^/]+)/reload`, func(c *goa.Context) {
		if err := Reload(c.Param(0), c.Param(1)); err == nil {
			c.Ok("reload success.")
		} else {
			c.Write([]byte(err.Error()))
		}
	})
}

func Detail(database, table, key, dataKeysStr string) (interface{}, error) {
	data := getData(database, table, key)
	if data == nil {
		return nil, fmt.Errorf("data %s.%s %s does not exists.", database, table, key)
	}
	dataKeysStr = strings.TrimSpace(dataKeysStr)
	var dataKeys []string
	if dataKeysStr != "" {
		dataKeys = strings.Split(dataKeysStr, ",")
		for i := range dataKeys {
			dataKeys[i] = strings.TrimSpace(dataKeys[i])
		}
	}
	return data.Data(dataKeys...)
}

func Reload(database, table string) error {
	cache := getCache(database, table)
	if cache == nil {
		return fmt.Errorf("table %s.%s does not exists.", database, table)
	}
	reload, ok := cache.(interface {
		Reload() error
	})
	if !ok {
		return fmt.Errorf("table %s.%s is not reloadable.", database, table)
	}
	return reload.Reload()
}

func getData(database, table, key string) Data {
	cache := getCache(database, table)
	if cache == nil {
		return nil
	}
	for _, data := range cache.GetDatas() {
		if data.Key() == key {
			return data
		}
	}
	return nil
}

func getCache(database, table string) Cache {
	tablesMap := cachesMap[database]
	if tablesMap == nil {
		return nil
	}
	return tablesMap[table]
}
