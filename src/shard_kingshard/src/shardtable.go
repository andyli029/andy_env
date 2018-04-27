// ./shardtable -config=ks_st.yaml -sqlfile=v.sql
// ./shard  -isHash=true -key="78592163702298085376845956" -slot=1024
// ./shard -c=./ddl_2.3_test/sharding.yaml -s=./ddl_2.3_test/V8__a.sql -isfly=true -target=8 -cmd=migrate
// ./shard -c=./ddl_2.3_test/sharding.yaml -s=./ddl_2.3_test/V8__a.sql -isfly=true -target=8 -cmd=migrate -isLocal=true
//
//  if want to use flyway:
//	./shardtable -c=/home/lizha/kingshard/src/shard_kingshard/sqlfileshard/ddl_2.3_test/sharding.yaml
// 	-s=/home/lizha/kingshard/src/shard_kingshard/sqlfileshard/ddl_2.3_test/V6__a.sql -isfly=true -target=6 -cmd=migrate
//
//  the cmd will create dir: $shard0 $shard1 in /home/lizha/kingshard/src/shard_kingshard/sqlfileshard/ddl_2.3_test/
//
//
//

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"errors"
	"bytes"

	"bufio"
	"io"
	"strings"
	"unsafe"
	"strconv"
	"hash/crc32"
	"reflect"
)

const (
	TABLE_HANDLE_FALSE = 0
	TABLE_HANDLE_TRUE  = 1
	TABLE_HANDLE_OVER  = 2
)

const (
	// SH_TK_ID_CREATE ...
	SH_TK_ID_CREATE = 1
	// SH_TK_ID_ALTER ...
	SH_TK_ID_ALTER = 2
)

var (
	PARSE_SH_TOKEN_MAP = map[string]int{
		"create":      SH_TK_ID_CREATE,
		"alter":       SH_TK_ID_ALTER,
	}
)

const NEW_SHARD_DIR = "/tmp/.shard/"
const CREATE_TABLE = "create table"
const ALTER_TABLE = "alter table"
const TABLE = "table"


func getDBTable(token string) (string, string) {
	//ks大小写不敏感: db table lowercase
	token = strings.ToLower(token)

	if len(token) == 0 {
		return "", ""
	}

	vec := strings.SplitN(token, ".", 2)
	if len(vec) == 2 {
		return strings.Trim(vec[0], "`"), strings.Trim(vec[1], "`")
	}
	return "", strings.Trim(vec[0], "`")
}

func IsSqlSep(r rune) bool {
	return r == ' ' || r == ',' ||
		r == '\t' || r == '/' ||
		r == '\n' || r == '\r'
}

func isDDL(line string) bool {
	line = strings.ToLower(line)

	uncleanWord := strings.FieldsFunc(line, IsSqlSep)

	cmd := uncleanWord[0]

	tokenID, ok := PARSE_SH_TOKEN_MAP[cmd]
	if ok == true {
		switch tokenID {
			case SH_TK_ID_CREATE:
				return true

			case SH_TK_ID_ALTER:
				return true

			default:
				return false
		}
	}

	return false
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t'
}

func filterSpace(sqlOld string) string {

	sql := []byte(sqlOld)

	var hasSpace = false
	for i := 0; i < len(sql); i++ {
		if isSpace(sql[i]) {
			sql[i] = ' '
			if hasSpace != true {
				hasSpace = true
				continue
			}
		}

		if hasSpace == true {
			for j := i; j < len(sql); {
				if isSpace(sql[j]) {
					sql = append(sql[:j], sql[j+1:]...)
				} else {
					hasSpace = false
					break
				}
			}
		}
	}

	return *(*string)(unsafe.Pointer(&sql))
	//return string(sql)
}

type sqlfile struct {
	CreateTable []string
}

//var configFileName_ string

var tables []string
//var shardTables []string

var flyLocation = make(map[string]string)

var (

	isLocal = flag.Bool("isLocal", false, "just create file locally, not exec command to remote")

	isHash = flag.Bool("isHash", false, "compute the hash, need the flag.")
	key =  flag.String("key", "", "compute the hash, need the key, only support string now.")
	slot =  flag.Uint64("slot", 0, "compute the hash, need the slot.")

	configFile = flag.String("c", "ks_st.yaml", "shardtable config file")
	sqlFile = flag.String("s", "v.sql", "separate table sql file")

	isfly  = flag.Bool("isfly", false, "is used by flyway or not, default is not")
	target = flag.String("target", "", "flyway target without V")
	cmd    = flag.String("cmd", "", "flyway cmd ---just support migrate now")
)

const banner string = "------------------SQL FILE SHARD----------------------------------------"

func execCommand(commandName string, cmdstr string) error {
	//显示运行的命令

	if (*isLocal) {
		fmt.Printf("just create file locally, not exec command to remote.\n")
		return nil
	}

	fmt.Printf("the cmd name is: %s.\n", cmdstr)

	w := bytes.NewBuffer(nil)
	cmdnew := exec.Command("sh", "-c", cmdstr)
	cmdnew.Stderr = w

	err := cmdnew.Run()

	if err != nil {
		fmt.Printf("unexpected execution failure error is: %v.\n", err)
		fmt.Printf("Stderr: %s\n", string(w.Bytes()))
		return err
	}
	return nil
}

func execCommand2(commandName string, cmds string) bool {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	cmd := exec.Command("sh", "-c", cmds)

	//显示运行的命令
	fmt.Println(cmd.Args)
	//StdoutPipe方法返回一个在命令Start后与命令标准输出关联的管道。Wait方法获知命令结束后会关闭这个管道，一般不需要显式的关闭该管道。
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return false
	}

	cmd.Start()
	//创建一个流来读取管道内内容，这里逻辑是通过一行一行的读取的
	reader := bufio.NewReader(stdout)

	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		fmt.Println(line)
	}

	//阻塞直到该命令执行完成，该命令必须是被Start方法开始执行的
	cmd.Wait()
	return true
}

func ReadLines(db string , fileName string, newFileName string, index int, handler func(string, string, string, int) error) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}

	fmt.Printf("the new filname name is: %s.\n", newFileName)
	_, err = os.OpenFile(newFileName, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		//line = line + "\n"
		//line = strings.TrimSpace(line)

		handler(line, db, newFileName, index)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return nil
}

func iteratorOneSQLInFile(db string , oldFile *os.File,  newFile *os.File, index int, handler func(string, string, int, *os.File) error) error {

	buf := bufio.NewReader(oldFile)
	for {
		oneSQL, err := buf.ReadString(';')
		//line = line + "\n"
		//line = strings.TrimSpace(line)

		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		err = handler(oneSQL, db, index, newFile)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return nil
}

func getShardSQLInfile(orgSQL string, newFile *os.File, table string, db string, handleState *int) error {

	fmt.Printf("the one sql is: %s \n", orgSQL)

	err, tableName := getTableNameINSQL(orgSQL)
	if err != nil {
		return err
	}

	if strings.ToLower(table) != tableName {
		return errors.New("the tableName in the sql file is not equal in the schema of the config")
	}

	n3, err := newFile.Write([]byte(orgSQL))
	fmt.Printf("got the table wrote %d bytes\n", n3)
	if err != nil {
		return err
	}
	newFile.Sync()

	return nil
}

func GetOneSQL(fileName string, newFileName string, table string, db string,
	handler func(string, *os.File, string, string, *int) error) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}

	fmt.Printf("the new filname name is: %s.\n", newFileName)
	newFile, err := os.OpenFile(newFileName, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	var handleState int
	handleState = TABLE_HANDLE_FALSE
	buf := bufio.NewReader(f)
	for {

		oneSQL, err := buf.ReadString(';')
		//line = line + "\n"
		//line = strings.TrimSpace(line)

		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		err = handler(oneSQL, newFile, table, db, &handleState)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}

	f.Close()
	newFile.Sync()
	newFile.Close()

	return nil
}


func filterShardSQL(fileName string, newFile string, table string, db string) (error) {

	return GetOneSQL(fileName, newFile, table, db, getShardSQLInfile)
}


func getTableNameINSQL(orgSQL string) (err error, tableName string) {

	oneSQL := strings.ToLower(orgSQL)
	oneSQL = filterSpace(oneSQL)

	var table string
	for k, _ := range PARSE_SH_TOKEN_MAP {

		if strings.Contains(oneSQL, k) {
			pos := strings.Index(oneSQL, TABLE)
			{
				uncleanWord := strings.FieldsFunc(oneSQL[pos:], IsSqlSep)
				_, table = getDBTable(uncleanWord[1])
			}
		}
	}

	if len(table) == 0 {
		return fmt.Errorf("table is null"), ""
	}

	return nil, table
}

//the table name
//  "mbk_cars"
//changed to
//	"mbk_shard_cars_$indexCloned"
//
func getOneShardingInFile(orgSQL string, db string, index int, newf *os.File) error {

	if len(orgSQL) == 0 {
		return fmt.Errorf("the one SQL is null.\n")
	}

	fmt.Printf("the newtablenameline: %s\n", orgSQL)

	err, tableName := getTableNameINSQL(orgSQL)
	if err != nil {
		return err
	}

	str_ind := fmt.Sprintf("%04d", index)
	shardTable := tableName + "_" + str_ind
	shardSQL := strings.Replace(orgSQL, tableName, shardTable, 1)

	_, err = newf.Write([]byte(shardSQL))
	//fmt.Printf("newtablename wrote %d bytes\n", n4)
	if err != nil {
		return err
	}
	newf.Sync()

	return nil
}

//sri: ShardRuleIndex
//cn: cloneNum
func ParsesSqlFile2(cfg *Config, db string, fileName string, sri int, cloneNum int, newFiles *[]string) (error) {

	fmt.Printf("the file name is: %s.\n", fileName)
	dirs := strings.Split(fileName, "/",)
	dlen := len(dirs)
	theLastDir := dirs[dlen - 1]
	fmt.Printf("the last dir name is: %s\n", theLastDir)

	theAbPath := strings.Split(fileName, ".sql")

	err := os.Mkdir(theAbPath[0], 0777)
	//fmt.Printf(",,,,,the Ab path name is: %s\n", theAbPath[0])
	if err != nil {
		return err
	}

	newFileNameprefix := theAbPath[0] + "/" + theLastDir
	//fmt.Printf("the prefix path name is: %s\n", newFileNameprefix)

	for i := 0; i < cloneNum; i++ {
		str_ind := fmt.Sprintf("%04d", i)

		newFileName := newFileNameprefix + "." + str_ind
		//fmt.Printf(",,,the new filname name is: %s\n", newFileName)

		oldf, err := os.OpenFile(fileName, os.O_RDWR, 0666)
		if err != nil {
			return err
		}

		newf, err := os.OpenFile(newFileName, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			return err
		}

		err = iteratorOneSQLInFile(db, oldf, newf, i, getOneShardingInFile)
		if err != nil {
			return err
		}

		oldf.Close()
		newf.Close()

		*newFiles = append(*newFiles, newFileName)
	}

	return nil
}

//ShardSQLFilesUnit
type ShardSQLFilesUnit struct {
	SQLFile	[]string				// the just only one table in one sqlfile
	cloneShardSQLFiles *[]string	// the prefix and the index is added in the clone tables which is in table@db dir
}

func shardSQLProcess(cfg *Config, fileName string) error {

	ShardRuleNum := len(cfg.Schema.ShardRule)
	ShardSQLFiles := make(map[string]map[string]*ShardSQLFilesUnit)

	for i := 0; i < ShardRuleNum; i++ {
		db := cfg.Schema.ShardRule[i].DB
		table := cfg.Schema.ShardRule[i].Table

		ssf := new(ShardSQLFilesUnit)
		cssf := make([]string, 0)
		ssf.cloneShardSQLFiles = &cssf

		//check: if the db table existed in rules
		if _, ok := ShardSQLFiles[db]; ok {
			if _, ok := ShardSQLFiles[db][table]; ok {
				return fmt.Errorf("table %s rule in %s duplicate", table, db)
			} else {
				ShardSQLFiles[db][table] = ssf
			}
		} else {
			m := make(map[string]*ShardSQLFilesUnit)
			ShardSQLFiles[db] = m				// think like:NewDefaultRule
			ShardSQLFiles[db][table] = ssf		// more think
		}
		// the prefix need be changed.
		newFileName := NEW_SHARD_DIR + table + "@" + db + ".sql"

		err := filterShardSQL(fileName, newFileName, table, db)
		if err != nil {
			return fmt.Errorf("getShardTableDbSQLfile error:%v\n", err.Error())
		}

		ShardSQLFiles[db][table].SQLFile = append(ShardSQLFiles[db][table].SQLFile, newFileName)

		var cloneNum int
		locations := cfg.Schema.ShardRule[i].Locations
		for i := 0; i < len(locations); i++ {
			cloneNum += locations[i]
		}
		fmt.Printf("get the locations, the cloneNum is: %d.\n", cloneNum)

		err = ParsesSqlFile2(cfg, db, ShardSQLFiles[db][table].SQLFile[0], i, cloneNum, ShardSQLFiles[db][table].cloneShardSQLFiles)
		if err != nil {
			return fmt.Errorf("Parses SqlFile failed error:%v\n", err.Error())
		}

		if *isfly {
			if len(*target) == 0 || len(*cmd) == 0 {
				return errors.New("use flyway, but the cmd is unknown\n")

				if *cmd != "migrate" {
					return fmt.Errorf("the cmd is unknown:%s.", *cmd)
				}
			}

			err = procSqlFilesByShardRuleByFLYWAY(cfg, i, db, table, *ShardSQLFiles[db][table].cloneShardSQLFiles, target, cmd)
			if err != nil {
				return err
			}
		} else {
			err = procSqlFilesByShardRule2(cfg, i, db, table, *ShardSQLFiles[db][table].cloneShardSQLFiles)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func procSqlFiletoNode(nodeCfg NodeConfig, sqlFile string) error {

	master := nodeCfg.Master
	strs   :=strings.Split(master, ":")
	addr   := strs[0]
	port   := strs[1]
	user   := nodeCfg.User
	passwd := nodeCfg.Password

	//fmt.Printf("the addr: %s, port: %s, user: %s, passwd: %s, sqlfile: %s.\n", addr, port, user, passwd, sqlFile)
	cmdstr := "mysql -h " + addr + " -P " + port + " -u " + user + " -p" + passwd + " < " + sqlFile

	// the execCommand2 from the www, the err is not return, so changed to execCommand1
	err := execCommand("sh", cmdstr)
	if err != nil {
		return fmt.Errorf("the sql cmd error.%s\n", cmdstr)
	}

	return nil
}

func procSqlFiletoNodeWithDB(nodeCfg NodeConfig, sqlFile string, db string) error {

	master := nodeCfg.Master
	strs   :=strings.Split(master, ":")
	addr   := strs[0]
	port   := strs[1]
	user   := nodeCfg.User
	passwd := nodeCfg.Password

	//fmt.Printf("the addr: %s, port: %s, user: %s, passwd: %s, sqlfile: %s.\n", addr, port, user, passwd, sqlFile)
	cmdstr := "mysql -h " + addr + " -P " + port + " -u " + user + " -p" + passwd + " -D" + db + " < " + sqlFile

	// the execCommand2 from the www, the err is not return, so changed to execCommand1
	err := execCommand("sh", cmdstr)
	if err != nil {
		return fmt.Errorf("the sql cmd error.%s\n", cmdstr)
	}

	return nil
}

// flyway -X -user=mobike001 -password=bHqOExjRU3OfX -url=jdbc:mysql://10.0.2.9:3306
// -schemas=mbk_pay -locations=filesystem:./mbk_pay -target=1 migrate
//

func procSqlFiletoNodeByFlYWAY(nodeCfg NodeConfig, db string, path string, target *string, cmd *string) error {

	master := nodeCfg.Master
	user   := nodeCfg.User
	passwd := nodeCfg.Password

	// Default table: flyway_schema_history
	// modou table: schema_version

	//fmt.Printf("the addr: %s, port: %s, user: %s, passwd: %s, sqlfile: %s.\n", addr, port, user, passwd, sqlFile)
	cmdstr := "flyway -X" + " -user=" + user + " -password=" + passwd + " -url=jdbc:mysql://" + master +
		" -table=schema_version" + " -schemas=" + db + " -locations=filesystem:" + flyLocation[nodeCfg.Name] +
			" -target=" + *target + " " + *cmd

	// the execCommand2 from the www, the err is not return, so changed to execCommand1
	err := execCommand("sh", cmdstr)
	if err != nil {
		return fmt.Errorf("the exec cmd:%s, error.%s", cmdstr, err.Error())
	}

	return nil
}

func togetherOneNodeSQL(sqlFilesSort map[string][]string, sqlFilesSortToOneNode map[string][]string) error {

	for k, v := range sqlFilesSort {

		node := k
		oneNodeFiles := v

		fmt.Printf("the node:%s, the files:%v \n", node, oneNodeFiles)

		// TODO FLYWAY ARG
		var nodepath string
		if *isfly == false {
			nodepath = "/tmp/.shard/" + node
		} else {
			if len(*sqlFile) == 0 || len(*target) == 0 {
				return errors.New("the sqlFile is null")
			}

			absPath := *sqlFile
			dirs := strings.Split(absPath, "/",)
			dlen := len(dirs)
			fileName := dirs[dlen - 1]

			theDir := strings.TrimRight(absPath, fileName)

			nodepath = theDir + node + "/" + fileName

			versionIndex := strings.Index(fileName, "_")
			versionNumStr := fileName[1:versionIndex]

			versionNum, err := strconv.Atoi(versionNumStr)
			if err != nil {
				return err
			}

			targetNum, err := strconv.Atoi(*target)
			if err != nil {
				return err
			}

			if versionNum != targetNum {
				return errors.New("the target is not equal the version.\n")
			}

			flyLocation[node] = theDir + "/" + node
		}

		sqlFilesSortToOneNode[node] = append(sqlFilesSortToOneNode[node], nodepath)

		nodeFile, err := os.OpenFile(nodepath, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			return err
		}

		nodeBuf := bufio.NewWriter(nodeFile)

		for _, v := range oneNodeFiles {

			f, err := os.Open(v)
			if err != nil {
				return err
			}

			buf := bufio.NewReader(f)

			n, err := buf.WriteTo(nodeBuf)
			fmt.Printf("got the table wrote %d bytes\n", n)
			if err != nil {
				return err
			}

			num, err := nodeBuf.WriteString("\n")
			fmt.Printf("got the table wrote %d bytes\n", num)
			if err != nil {
				return err
			}

			nodeBuf.Flush()
			f.Close()
		}

		nodeFile.Close()

	}
	return nil
}

//SchemaShardRuleIndex ssri
func procSqlFilesByShardRule2(cfg *Config, ssri int, db string, table string, sqlFiles []string) error{

	sqlFilesSort := make(map[string][]string)
	sqlFilesSortToOneNode := make(map[string][]string)

	shardNodes := cfg.Schema.ShardRule[ssri].Nodes

	sqlFilesIndex := 0
	for i := 0; i < len(shardNodes); i++ {
		node := shardNodes[i]
		locations := cfg.Schema.ShardRule[ssri].Locations[i]

		if _, ok := sqlFilesSort[node]; ok {
			return fmt.Errorf("the node:%s duplicate", node)
		}

		for j := locations; j > 0; j-- {
			sqlFilesSort[node] = append(sqlFilesSort[node], sqlFiles[sqlFilesIndex])
			sqlFilesIndex ++
		}
	}

	if (sqlFilesIndex != len(sqlFiles)) {
		return fmt.Errorf("the Locations num is not qual the sql file num:<%d, %d>.\n", sqlFilesIndex, sqlFiles)
	}

	err := togetherOneNodeSQL(sqlFilesSort, sqlFilesSortToOneNode)
	if err != nil {
		fmt.Printf("together one sharding file to node SQL file:%v\n", err.Error())
		return err
	}

	var NodeConf NodeConfig
	for m := 0; m < len(shardNodes); m++ {

		node := shardNodes[m]
		// TODO NodeConf clean
		for k := 0; k < len(cfg.Nodes); k++ {
			if cfg.Nodes[k].Name == shardNodes[m] {
				NodeConf = cfg.Nodes[k]
				break
			}
		}

		// change the sql files together to onenode sql file
		for l := 0; l < len(sqlFilesSort[node]); l++ {
			//err := procSqlFiletoNode(NodeConf, sqlFilesSort[node][l])
			//err := procSqlFiletoNodeWithDB(NodeConf, sqlFilesSort[node][l], db)
			//err := procSqlFiletoNodeWithDB(NodeConf, sqlFilesSortToOneNode[node][0], db)
			if err != nil {
				return err
			}
		}

		err := procSqlFiletoNodeWithDB(NodeConf, sqlFilesSortToOneNode[node][0], db)
		if err != nil {
			return err
		}
	}
	return nil
}

// SchemaShardRuleIndex ssri
func procSqlFilesByShardRuleByFLYWAY(cfg *Config, ssri int, db string, table string, sqlFiles []string, target *string, cmd *string) error{

	sqlFilesSort := make(map[string][]string)
	sqlFilesSortToOneNode := make(map[string][]string)
	shardNodes := cfg.Schema.ShardRule[ssri].Nodes

	sqlFilesIndex := 0
	for i := 0; i < len(shardNodes); i++ {
		node := shardNodes[i]
		locations := cfg.Schema.ShardRule[ssri].Locations[i]

		if _, ok := sqlFilesSort[node]; ok {
			return fmt.Errorf("the node:%s duplicate", node)
		}

		for j := locations; j > 0; j-- {
			sqlFilesSort[node] = append(sqlFilesSort[node], sqlFiles[sqlFilesIndex])
			sqlFilesIndex ++
		}
	}

	if (sqlFilesIndex != len(sqlFiles)) {
		return fmt.Errorf("the Locations num is not qual the sql file num:<%d, %d>.\n", sqlFilesIndex, sqlFiles)
	}

	err := togetherOneNodeSQL(sqlFilesSort, sqlFilesSortToOneNode)
	if err != nil {
		return err
	}

	var NodeConf NodeConfig
	for m := 0; m < len(shardNodes); m++ {

		node := shardNodes[m]
		// TODO NodeConf clean
		for k := 0; k < len(cfg.Nodes); k++ {
			if cfg.Nodes[k].Name == shardNodes[m] {
				NodeConf = cfg.Nodes[k]
				break
			}
		}

		err := procSqlFiletoNodeByFlYWAY(NodeConf, db, sqlFilesSortToOneNode[node][0], target, cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func checkShardDir() error {
	shardDir := NEW_SHARD_DIR

	isEx, err := PathExists(shardDir)
	if err != nil {
		return  err
	}

	if isEx == false {
		err := os.Mkdir(shardDir, 0777)
		if err != nil {
			return  err
		}
	}

	return nil
}

// the string optimize
func slice(s string) (b []byte) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pbytes.Data = pstring.Data
	pbytes.Len = pstring.Len
	pbytes.Cap = pstring.Len
	return
}

func HashValue(value interface{}) uint64 {
	switch val := value.(type) {
	case int:
		return uint64(val)
	case uint64:
		return uint64(val)
	case int64:
		return uint64(val)
	case string:
		if v, err := strconv.ParseUint(val, 10, 64); err != nil {
			return uint64(crc32.ChecksumIEEE(slice(val)))
		} else {
			return uint64(v)
		}
	case []byte:
		return uint64(crc32.ChecksumIEEE(val))
	}
	panic("ERROR:Unexpected key variable type.")
}

func main_2() {
	flag.Parse()
	fmt.Println(banner)

	if *isHash {
		if len(*key) == 0 || *slot == 0 {
			fmt.Println("ERROR:must input one key and one slot")
			os.Exit(1)
		}

		idx := HashValue(*key) % (*slot)
		fmt.Printf("the hash index is %04d.\n", idx)

		os.Exit(0)
	}

	if len(*configFile) == 0 {
		fmt.Println("ERROR:must use a config file")
		os.Exit(1)
	}

	if len(*sqlFile) == 0 {
		fmt.Println("ERROR:must use a sql file")
		os.Exit(1)
	}

	fmt.Printf("the config file is: %s.\n", *configFile)
	fmt.Printf("the sql file is: %s.\n", *sqlFile)
	fmt.Printf("the sql file is: %t.\n", *isfly)

	err :=os.RemoveAll(NEW_SHARD_DIR)
	if err != nil {
		fmt.Printf("parse config file error:%v\n", err.Error())
		os.Exit(1)
	}

	cfg, err := ParseConfigFile(*configFile)
	if err != nil {
		fmt.Printf("ERROR:parse config file error:%v\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("the nodes num is: %d.\n", len(cfg.Nodes))

	err = checkShardDir()
	if err != nil {
		fmt.Printf("ERROR:check Shard Dir error:%v\n", err.Error())
		os.Exit(1)
	}

	//var hashTables []string
	err = shardSQLProcess(cfg, *sqlFile)
	if err != nil {
		fmt.Printf("ERROR:shardSQLProcess error:%v\n", err.Error())
		os.Exit(1)
	}

	for k := 0; k < len(tables); k++ {
		fmt.Printf("!!!!!!!!!!the tables name is: %s.\n", tables[k])
	}
}

func main() {
	main_2()
}

/*
func getShardTableSQLFile(cfg *Config, fileName string) error {

	ShardRuleNum := len(cfg.Schema.ShardRule)
	ShardSQLFiles := make(map[string]map[string]*ShardSQLFilesUnit)

	for i := 0; i < ShardRuleNum; i++ {
		db := cfg.Schema.ShardRule[i].DB
		table := cfg.Schema.ShardRule[i].Table

		ssf := new(ShardSQLFilesUnit)
		cssf := make([]string, 0)
		ssf.cloneShardSQLFiles = &cssf

		//check: if the database exist in rules
		if _, ok := ShardSQLFiles[db]; ok {
			if _, ok := ShardSQLFiles[db][table]; ok {
				return fmt.Errorf("table %s rule in %s duplicate", table, db)
			} else {
				ShardSQLFiles[db][table] = ssf
			}
		} else {
			m := make(map[string]*ShardSQLFilesUnit)
			ShardSQLFiles[db] = m				// think like:NewDefaultRule
			ShardSQLFiles[db][table] = ssf		// more think
		}
		// the prefix need be changed.
		newFileName :=  NEW_SHARD_DIR + table + "@" + db + ".sql"
		err := filterShardTableDbSQLfile(fileName, newFileName, table, db)
		if err != nil {
			fmt.Printf("getShardTableDbSQLfile error:%v\n", err.Error())
			return err
		}

		ShardSQLFiles[db][table].SQLFile = append(ShardSQLFiles[db][table].SQLFile, newFileName)

		var cloneNum int
		locations := cfg.Schema.ShardRule[i].Locations
		for i := 0; i < len(locations); i++ {
			cloneNum += locations[i]
		}
		fmt.Printf("get the locations, the cloneNum is: %d.\n", cloneNum)

		err = ParsesSqlFile(cfg, db, ShardSQLFiles[db][table].SQLFile[0], i, cloneNum, ShardSQLFiles[db][table].cloneShardSQLFiles)
		if err != nil {
			fmt.Printf("parse config file error:%v\n", err.Error())
			return fmt.Errorf("ParsesSqlFile failed.")
		}

		// lizhiang 到每个节点执行相应的语句，注释掉不执行，只生成相应的sql
		err = procSqlFilesByShardRule2(cfg, i, db, table, *ShardSQLFiles[db][table].cloneShardSQLFiles)
		if err != nil {
			fmt.Printf("process config file error:%v\n", err.Error())
			os.Exit(1)
		}
	}

	return nil
}

func getShardTableDbSQLfile(line string, newFilName string, table string, db string, handleState *int) error {
	//fmt.Printf("the create table is: %S.\n", line)

	if *handleState == TABLE_HANDLE_OVER {
		return nil
	}

	//newFilName := table + "@" + db + ".sql"
	//fmt.Printf("11111the new filname name is: %s.\n", newFilName)
	f, err := os.OpenFile(newFilName, os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	//打开文件后，习惯立即使用 defer 调用文件的 Close操作。
	defer f.Close()


	if strings.Contains(line, CREATE_TABLE) {
		//if isDDL(line) {

		// TODO the end judge need be repaired
		// judge the end from "CREATE TABLE"
		if *handleState == TABLE_HANDLE_TRUE {
			*handleState = TABLE_HANDLE_OVER
			fmt.Printf("got the table name and the content, GAME OVER!\n")
			return nil
		}

		if strings.Contains(line, table) {
			fmt.Printf("In the head, get the table name: %s.\n", table)
			*handleState = TABLE_HANDLE_TRUE
		}
	}

	if *handleState == TABLE_HANDLE_TRUE {

		if len(line) == 0 {
			fmt.Printf("the line is null.\n")
			return nil
		}

		n3, err := f.Write([]byte(line))
		fmt.Printf("got the table wrote %d bytes\n", n3)
		if err != nil {
			return err
		}
		f.Sync()
	}

	return nil
}

func filterShardTableDbSQLfile(fileName string, newFileName string, table string, db string) (error) {

	return ReadLines_infilter(fileName, newFileName, table, db, getShardTableDbSQLfile)
}

func ReadLines_infilter(fileName string, newFileName string, table string, db string,
	handler func(string, string, string, string, *int) error) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}

	fmt.Printf("the new filname name is: %s.\n", newFileName)
	_, err = os.OpenFile(newFileName, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	var handleState int
	handleState = TABLE_HANDLE_FALSE
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		//line = line + "\n"
		//line = strings.TrimSpace(line)

		handler(line, newFileName, table, db, &handleState)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return nil
}


//the table name
//  "mbk_cars"
//changed to
//	"mbk_shard_cars_$indexCloned"
//
//TODO FIX:the function is impled by bytes, could be replaced by using strings
func getCreateTable(line string, db string, newFilName string, index int) error {
	//fmt.Printf("the create table is: %S.\n", line)

	//newfilname := filename + "." + strconv.Itoa(index)
	//fmt.Printf("the new filname name is: %s.\n", newFilName)
	f, err := os.OpenFile(newFilName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	//打开文件后，习惯立即使用 defer 调用文件的 Close操作。
	defer f.Close()

	if len(line) == 0 {
		fmt.Printf("the line is null.\n")
		return nil
	}

	head := CREATE_TABLE
	i := 0
	for i = 0; i < len(head); i++ {
		if line[i] != head[i] {
			//return nil
			n2, err := f.Write([]byte(line))
			fmt.Printf("wrote %d bytes\n", n2)
			if err != nil {
				return err
			}
			f.Sync()

			return nil
		}
	}

	//fmt.Printf("the create table is: %s.\n", line)
	left := false
	var left_index int
	//var right_index int

	//fmt.Printf("the table index is: %d.\n", i)
	var newtablename string
	var post string
	j := i
	for j = i; j < len(line); j++ {
		if line[j] == '`' {
			if left == false {
				left = true
				left_index = j
			} else {
				//tablename := make([]byte, (j - left_index + 1))
				post = line[j:] //post == "` (\n"
				//fmt.Printf("&&&&the post name is:%s\n", post)
				tablename := line[left_index+1:j]
				//fmt.Printf("the create table name is:%s.\n", tablename)

				for k := 0; k < len(tablename); k++ {
					if tablename[k] == '_' {
						pre := tablename[:k]
						//newpre := "shard_" + pre
						newpre := pre
						//str_ind := strconv.Itoa(index)
						//%04d
						str_ind := fmt.Sprintf("%04d", index)
						newtablename = newpre + tablename[k:] + "_" + str_ind
						fmt.Printf("@@@@@the new table name is: %s.\n", newtablename)

						// used for debug
						tables = append(tables, newtablename)
						break
					}
				}

				break
			}

		} else {
			continue
		}
	}

	newtablnameline := CREATE_TABLE + " `" + newtablename + post//+ "\n"
	strings := strings.Split(newtablnameline, "\n")
	fmt.Printf("the newtablenameline: [%s].\n", strings[0])

	// TODO create database
	// add the use db in the head
	dbline := "use " + db + ";\n"
	_, err = f.WriteAt([]byte(dbline), 0)
	//fmt.Printf("newtablename wrote %d bytes\n", n3)
	if err != nil {
		return err
	}
	f.Sync()

	_, err = f.Write([]byte(newtablnameline))
	//fmt.Printf("newtablename wrote %d bytes\n", n4)
	if err != nil {
		return err
	}
	f.Sync()

	return nil
}

//SchemaShardRuleIndex ssri
func procSqlFilesByShardRule(cfg *Config, ssri int, db string, table string, sqlFiles []string) error{

	sqlFilesSort := make(map[string][]string)

	shardNodes := cfg.Schema.ShardRule[ssri].Nodes

	sqlFilesIndex := 0
	for i := 0; i < len(shardNodes); i++ {
		node := shardNodes[i]
		locations := cfg.Schema.ShardRule[ssri].Locations[i]

		if _, ok := sqlFilesSort[node]; ok {
			return fmt.Errorf("the node:%s duplicate", node)
		}

		for j := locations; j > 0; j-- {
			sqlFilesSort[node] = append(sqlFilesSort[node], sqlFiles[sqlFilesIndex])
			sqlFilesIndex ++
		}
	}

	if (sqlFilesIndex != len(sqlFiles)) {
		fmt.Printf("the Locations num is not qual the sql file num error.\n")
		return errors.New("NOT EQUAL.")
	}

	var NodeConf NodeConfig
	for m := 0; m < len(shardNodes); m++ {

		node := shardNodes[m]
		// TODO NodeConf clean
		for k := 0; k < len(cfg.Nodes); k++ {
			if cfg.Nodes[k].Name == shardNodes[m] {
				NodeConf = cfg.Nodes[k]
				break
			}
		}

		for l := 0; l < len(sqlFilesSort[node]); l++ {
			//err := procSqlFiletoNode(NodeConf, sqlFilesSort[node][l])
			err := procSqlFiletoNodeWithDB(NodeConf, sqlFilesSort[node][l], db)
			if err != nil {
				fmt.Printf("process config file error:%v\n", err.Error())
				return err
			}
		}
	}
	return nil
}

func main_1() {
	flag.Parse()
	fmt.Println(banner)

	if len(*configFile) == 0 {
		fmt.Println("must use a config file")
		os.Exit(1)
	}

	if len(*sqlFile) == 0 {
		fmt.Println("must use a sql file")
		os.Exit(1)
	}

	fmt.Printf("the config file is: %s.\n", *configFile)
	fmt.Printf("the sql file is: %s.\n", *sqlFile)

	cfg, err := parseConfigFile(*configFile)
	if err != nil {
		fmt.Printf("parse config file error:%v\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("the nodes num is: %d.\n", len(cfg.Nodes))

	err = checkShardDir()
	if err != nil {
		fmt.Printf("check Shard Dir error:%v\n", err.Error())
		os.Exit(1)
	}

	//var hashTables []string
	err = getShardTableSQLFile(cfg, *sqlFile)
	if err != nil {
		fmt.Printf("get ShardTableSQLFile error:%v\n", err.Error())
		os.Exit(1)
	}

	for k := 0; k < len(tables); k++ {
		fmt.Printf("!!!!!!!!!!the tables name is: %s.\n", tables[k])
	}
}

*/