package got

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

func (d Download) mergeFiles(sections [][2]int) error {
	f, err := os.OpenFile(path.Join(d.DirPath, d.TargetPath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	for i := range sections {
		tmpFileName := fmt.Sprintf("section-%v.tmp", i)
		b, err := ioutil.ReadFile(path.Join(d.DirPath, tmpFileName))
		if err != nil {
			return err
		}
		_, err = f.Write(b)
		if err != nil {
			return err
		}
		err = os.Remove(tmpFileName)
		if err != nil {
			return err
		}

	}

	return nil

}
