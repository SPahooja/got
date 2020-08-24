package got

import (
	"fmt"
	"got/progress"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

func (d Download) downloadSection(i int, c [2]int, sum *int, bar *progress.Bar) error {
	r, err := d.getNewRequest("GET")
	if err != nil {
		return err
	}
	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", c[0], c[1]))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		return fmt.Errorf("Can't process, response is %v", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path.Join(d.DirPath, fmt.Sprintf("section-%v.tmp", i)), b, os.ModePerm)
	if err != nil {
		return err
	}
	fmt.Printf("section-%v.tmp", i)
	*sum = *sum + 1
	bar.Play(int64(*sum))

	return nil
}
