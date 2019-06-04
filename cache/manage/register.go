package manage

import (
	"log"
)

type Cache interface {
	Datas() []Data
}

type Data interface {
	Key() string
	Desc() string
	Data() interface{}
}

var cachesMap = make(map[string]map[string]Cache)

func TryRegister(database, table string, ifc interface{}) {
	cache, ok := ifc.(Cache)
	if !ok {
		return
	}
	tablesMap := cachesMap[database]
	if tablesMap == nil {
		tablesMap = make(map[string]Cache)
		cachesMap[database] = tablesMap
	}

	if _, ok := tablesMap[table]; ok {
		log.Panicf("%s.%s aready exists", database, table)
	}
	tablesMap[table] = cache
}

func Unregister(database, table string) {
	tablesMap := cachesMap[database]
	if tablesMap == nil {
		return
	}
	delete(tablesMap, table)
}

func UnregisterDB(database string) {
	delete(cachesMap, database)
}
