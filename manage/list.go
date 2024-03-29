package manage

import (
	"bytes"
	"fmt"
	"os"
	"sort"
)

func List() []byte {
	hostName, _ := os.Hostname()
	return []byte(listHtml(hostName, listHtmlTable()))
}

func listHtmlTable() string {
	var dbs = make([]string, 0, len(cachesMap))
	for db := range cachesMap {
		dbs = append(dbs, db)
	}
	sort.Strings(dbs)

	buf := bytes.NewBufferString(`
<table>
<tr> <th>Database</th> <th>Table</th> <th>Data</th> <th>Size</th> <th>Operation</th> </tr>

`)

	for _, db := range dbs {
		rows, count := listDbTables(db, cachesMap[db])
		buf.WriteString(fmt.Sprintf(
			"<tr> <td%s>%s</td> %s\n", rowspanAttr(count), db, rows,
		))
	}
	buf.WriteString("</table>")
	return buf.String()
}

func listDbTables(db string, tablesMap map[string]Cache) (string, int) {
	var tables = make([]string, 0, len(tablesMap))
	for table := range tablesMap {
		tables = append(tables, table)
	}
	sort.Strings(tables)

	var buf = bytes.NewBuffer(nil)
	var totalCount int
	for i, table := range tables {
		if i > 0 {
			buf.WriteString("<tr> ")
		}
		totalCount += listDbTable(buf, db, table, tablesMap[table])
	}
	return buf.String(), totalCount
}

func listDbTable(buf *bytes.Buffer, db, table string, cache Cache) int {
	var datas = cache.GetDatas()
	var data0 Data
	if len(datas) > 0 {
		data0 = datas[0]
	}

	var reload string
	if _, ok := cache.(interface {
		Reload(noClear bool) error
	}); ok {
		reload = fmt.Sprintf(`<a href="./caches/%s/%s/reload">reload</a>`, db, table)
	}

	buf.WriteString(fmt.Sprintf(`<td%s>%s</td>
%s
<td%s>%s</td>
</tr>
`, rowspanAttr(len(datas)), table, listData(db, table, data0), rowspanAttr(len(datas)), reload,
	))

	for i := 1; i < len(datas); i++ {
		buf.WriteString(fmt.Sprintf("<tr> %s </tr>\n", listData(db, table, datas[i])))
	}
	if len(datas) == 0 {
		return 1
	}
	return len(datas)
}

func listData(db, table string, data Data) string {
	if data == nil {
		return `<td></td> <td></td>`
	}
	return fmt.Sprintf(
		`<td class="data"><a href="./caches/%s/%s/%s">%s</a></td> <td>%d</td>`,
		db, table, data.Key(), data.Key(), data.Size(),
	)
}

func rowspanAttr(count int) string {
	if count <= 1 {
		return ""
	}
	return fmt.Sprintf(` rowspan="%d"`, count)
}
