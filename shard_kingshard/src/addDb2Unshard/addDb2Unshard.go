package addDb2Unshard
// ./shardtable -config=ks_st.yaml -sqlfile=v.sql
import (
	"fmt"
	"os"
	"io"
	"bufio"
)

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

func addPreAndTail(line string, db string, newFilName string, index int) error {

	f, err := os.OpenFile(newFilName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	//打开文件后，习惯立即使用 defer 调用文件的 Close操作。
	defer f.Close()

	newtablnameline := "-\n" + "    db: " + line + "    node_groups: [local]\n"

	_, err = f.Write([]byte(newtablnameline))
	//fmt.Printf("newtablename wrote %d bytes\n", n4)
	if err != nil {
		return err
	}
	f.Sync()

	return nil
}

func addDb2Unshard(fileName string) (error) {

	newFileName := "/tmp/2"
	//fmt.Printf(",,,the new filname name is: %s\n", newFileName)
	_, err := os.OpenFile(newFileName, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	err = ReadLines("", fileName, newFileName, 0, addPreAndTail)
	if err != nil {
		return err
	}

	return nil
}

func main() {

	err := addDb2Unshard("/tmp/1")
	if err != nil {
		fmt.Printf("check Shard Dir error:%v\n", err.Error())
		os.Exit(1)
	}

}
