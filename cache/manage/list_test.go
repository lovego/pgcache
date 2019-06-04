package manage

import "fmt"

type testCache1 struct {
	datas []Data
}
type testCache2 struct {
	testCache1
}
type testData struct {
	key, desc string
	data      interface{}
}

func (t testCache1) Datas() []Data {
	return t.datas
}
func (t testCache2) Reload() error {
	return nil
}

func (t testData) Key() string {
	return t.key
}
func (t testData) Desc() string {
	return t.desc
}
func (t testData) Data() interface{} {
	return t.data
}

func ExampleList() {
	var table1 testCache1
	var table2, table3 testCache2

	table1.datas = []Data{
		testData{`key1.1`, `desc1.1`, nil},
		testData{`key1.2`, `desc1.2`, nil},
	}
	table2.datas = []Data{
		testData{`key2.1`, `desc2.1`, nil},
		testData{`key2.2`, `desc2.2`, nil},
		testData{`key2.3`, `desc2.3`, nil},
	}
	table3.datas = []Data{
		testData{`key3.1`, `desc3.1`, nil},
	}

	TryRegister(`db1`, `table1`, table1)
	TryRegister(`db2`, `table2`, table2)
	TryRegister(`db2`, `table3`, table3)
	fmt.Println(string(List()))
	// Output:
	// <table style="width: 70%; margin: 5% 10%; border-collapse: collapse;">
	// <style>td {padding: 5px 10px; border: 1px dashed gray; }</style>
	// <tr> <th>Database</th> <th>Table</th> <th>Data</th> <th>Operations</th> </tr>
	//
	// <tr> <td rowspan="2">db1</td> <td rowspan="2">table1</td>
	// <td><a href="/caches/db1/table1/key1.1">key1.1(desc1.1)</a></td>
	// <td rowspan="2"></td>
	// </tr>
	// <tr> <td><a href="/caches/db1/table1/key1.2">key1.2(desc1.2)</a></td> </tr>
	//
	// <tr> <td rowspan="4">db2</td> <td rowspan="3">table2</td>
	// <td><a href="/caches/db2/table2/key2.1">key2.1(desc2.1)</a></td>
	// <td rowspan="3"><a href="/caches/db2/table2/reload">reload</a></td>
	// </tr>
	// <tr> <td><a href="/caches/db2/table2/key2.2">key2.2(desc2.2)</a></td> </tr>
	// <tr> <td><a href="/caches/db2/table2/key2.3">key2.3(desc2.3)</a></td> </tr>
	// <tr> <td>table3</td>
	// <td><a href="/caches/db2/table3/key3.1">key3.1(desc3.1)</a></td>
	// <td><a href="/caches/db2/table3/reload">reload</a></td>
	// </tr>
	//
	// </table>
}
