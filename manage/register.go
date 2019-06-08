package manage

import "fmt"

type Cache interface {
	GetDatas() []Data
}

type Data interface {
	Key() string
	Size() int
	Data() interface{}
}

var cachesMap = make(map[string]map[string]Cache)

func Register(database, table string, cache Cache) error {
	tablesMap := cachesMap[database]
	if tablesMap == nil {
		tablesMap = make(map[string]Cache)
		cachesMap[database] = tablesMap
	}

	if _, ok := tablesMap[table]; ok {
		return fmt.Errorf("%s.%s aready exists", database, table)
	}
	tablesMap[table] = cache
	return nil
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
