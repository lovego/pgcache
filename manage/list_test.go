package manage

import "fmt"

type testCache1 struct {
	datas []Data
}
type testCache2 struct {
	testCache1
}
type testData struct {
	key  string
	size int
	data interface{}
}

func (t testCache1) GetDatas() []Data {
	return t.datas
}
func (t testCache2) Reload(noClear bool) error {
	return nil
}

func (t testData) Key() string {
	return t.key
}
func (t testData) Size() int {
	return t.size
}
func (t testData) Data(...string) (interface{}, error) {
	return t.data, nil
}

func ExampleListHtmlTable() {
	var table1 testCache1
	var table2, table3 testCache2

	table1.datas = []Data{
		testData{`key1.1`, 1, nil},
		testData{`key1.2`, 5, nil},
	}
	table2.datas = []Data{
		testData{`key2.1`, 4, nil},
		testData{`key2.2`, 9, nil},
		testData{`key2.3`, 0, nil},
	}
	table3.datas = []Data{
		testData{`key3.1`, 3, nil},
	}

	if err := Register(`db1`, `table1`, table1); err != nil {
		panic(err)
	}
	if err := Register(`db2`, `table2`, table2); err != nil {
		panic(err)
	}
	if err := Register(`db2`, `table3`, table3); err != nil {
		panic(err)
	}
	fmt.Println(listHtmlTable())
	// Output:
	// <table>
	// <tr> <th>Database</th> <th>Table</th> <th>Data</th> <th>Size</th> <th>Operation</th> </tr>
	//
	// <tr> <td rowspan="2">db1</td> <td rowspan="2">table1</td>
	// <td class="data"><a href="/caches/db1/table1/key1.1">key1.1</a></td> <td>1</td>
	// <td rowspan="2"></td>
	// </tr>
	// <tr> <td class="data"><a href="/caches/db1/table1/key1.2">key1.2</a></td> <td>5</td> </tr>
	//
	// <tr> <td rowspan="4">db2</td> <td rowspan="3">table2</td>
	// <td class="data"><a href="/caches/db2/table2/key2.1">key2.1</a></td> <td>4</td>
	// <td rowspan="3"><a href="/caches/db2/table2/reload">reload</a></td>
	// </tr>
	// <tr> <td class="data"><a href="/caches/db2/table2/key2.2">key2.2</a></td> <td>9</td> </tr>
	// <tr> <td class="data"><a href="/caches/db2/table2/key2.3">key2.3</a></td> <td>0</td> </tr>
	// <tr> <td>table3</td>
	// <td class="data"><a href="/caches/db2/table3/key3.1">key3.1</a></td> <td>3</td>
	// <td><a href="/caches/db2/table3/reload">reload</a></td>
	// </tr>
	//
	// </table>

}
