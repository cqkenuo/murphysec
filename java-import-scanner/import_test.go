package java_import_scanner

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"murphysec-cli-simple/logger"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestName(t *testing.T) {
	dir := "C:\\Users\\iseki\\Desktop\\新建文件夹\\HundredBai_camellia_master"
	filepathCh := make(chan string, 50)
	// 文件读取
	go func() {
		// 去重，防环
		visited := map[string]struct{}{}
		// 广度优先遍历目录树
		q := list.New()
		q.PushBack(dir)

		for q.Len() > 0 {
			curr := q.Front().Value.(string)
			q.Remove(q.Front())
			if _, ok := visited[curr]; ok {
				continue
			}
			visited[curr] = struct{}{}
			flist, e := ioutil.ReadDir(curr)
			if e != nil {
				logger.Err.Println("ReadDir failed,", curr, ".", e.Error())
				continue
			}
			for _, it := range flist {
				if it.IsDir() {
					q.PushBack(filepath.Join(curr, it.Name()))
					continue
				}
				if strings.HasSuffix(it.Name(), ".java") {
					filepathCh <- filepath.Join(curr, it.Name())
				}
			}
		}
		close(filepathCh)
	}()
	// parsing
	parserWg := sync.WaitGroup{}
	type td struct {
		imports  []string
		filename string
	}
	c := make(chan td)
	for i := 0; i < 4; i++ {
		parserWg.Add(1)
		go func() {
			for {
				filePath := <-filepathCh
				if filePath == "" {
					break
				}
				c <- td{imports: parseJavaFileImport(filePath), filename: filePath}
			}
			parserWg.Done()
		}()
	}
	go func() {
		for {
			t := <-c
			if t.filename == "" {
				break
			}
			fmt.Println("filename:", t.filename)
			for _, it := range t.imports {
				fmt.Println("  import", it)
			}
		}
	}()
	parserWg.Wait()
	close(c)
	return
}
