package got

import (
	"errors"
	"fmt"
	"got/metadata"
	"got/progress"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/muesli/termenv"
)

var CProfile = termenv.ColorProfile()

type Download struct {
	URL           string
	TargetPath    string
	TotalSections int
	DirPath       string
}

func Got(url string, path string) error {
	if path == "" {
		path, err := os.Getwd()
		if err != nil {
			log.Println(err)
			return err
		}
		err = Letsgot(url, path)
		if err != nil {
			return err
		}

	} else {
		err := Letsgot(url, path)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewGot() error {
	url, err := getlink()
	if err != nil {
		return err
	}
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return err
	}
	err = Letsgot(url, path)
	if err != nil {
		return err
	}
	return nil
}

func Letsgot(url string, path string) error {

	//startTime := time.Now()
	var bar progress.Bar
	bar.NewOption(0, 10)

	d := Download{
		URL:           url,
		TargetPath:    "",
		TotalSections: 10,
		DirPath:       path,
	}

	if err := d.do(&bar); err != nil {
		fmt.Printf("\n%s\n", termenv.String("An error occured while downloading the file").Foreground(CProfile.Color("0")).Background(CProfile.Color("#E88388")))
		// fmt.Printf("\n%s", err)
		log.Fatal(err)
		return err
	}

	//fmt.Printf("\n\n ✅ Download completed in %v seconds\n", time.Now().Sub(startTime).Seconds())

	return nil
}

func getlink() (string, error) {
	var url string
	fmt.Printf("%s\n", termenv.String("Enter URL").Foreground(CProfile.Color("#FDD835")))
	fmt.Scanf("%s\n", &url)

	if url == "" {
		log.Fatal("URL can't be empty")
		return "", errors.New("URL can't be empty")
	}
	return url, nil
}

func (d Download) do(bar *progress.Bar) error {

	fmt.Printf("\n%s\n", termenv.String("🔍 CHECKING URL ").Bold().Foreground(CProfile.Color("#FDD835")))

	r, err := d.getNewRequest("HEAD")
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(r)

	if err != nil {
		return err
	}

	if resp.StatusCode > 299 {
		return fmt.Errorf("Can't process, got %v response", resp.StatusCode)
	}

	fmt.Printf("\n%s\n", termenv.String(" URL OK ").Bold().Foreground(CProfile.Color("#8BC34A")))
	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}

	a := metadata.GetFileName(resp)
	d.TargetPath = a

	var yesOrNo string
	fmt.Printf("\n%s %f %s\n", termenv.String(" => SIZE IS ").Bold().Foreground(CProfile.Color("#FFB300")), (float64(size)*float64(0.001))/1000, "MB")
	fmt.Printf("\n%s\n", termenv.String("🌑 PROCEED DOWNLOADING ? ([Y}es/[N]o)").Bold().Foreground(CProfile.Color("#FFB300")))
	fmt.Scanf("%s", &yesOrNo)

	if yesOrNo == "n" || yesOrNo == "no" || yesOrNo == "N" || yesOrNo == "NO" || yesOrNo == "nO" || yesOrNo == "No" {
		fmt.Printf("%s\n", termenv.String("🚫EXITING").Bold().Foreground(CProfile.Color("#FF7043")))
		os.Exit(2)
		return nil
	}

	fmt.Printf("\n%s\n", termenv.String("🚀 Downloading... ").Bold().Foreground(CProfile.Color("#8BC34A")))

	var sections = make([][2]int, d.TotalSections)

	eachSize := size / d.TotalSections

	now := time.Now()
	for i := range sections {
		if i == 0 {
			sections[i][0] = 0
		} else {
			sections[i][0] = sections[i-1][1] + 1
		}

		if i < d.TotalSections-1 {
			sections[i][1] = sections[i][0] + eachSize
		} else {
			sections[i][1] = size - 1
		}
	}

	var wg sync.WaitGroup
	sum2 := 0

	for i, s := range sections {
		wg.Add(1)

		go func(i int, s [2]int, bar *progress.Bar) {
			defer wg.Done()
			err = d.downloadSection(i, s, &sum2, bar)
			if err != nil {
				panic(err)
			}
		}(i, s, bar)

	}
	wg.Wait()
	fmt.Printf("\n%s\n", termenv.String("🚀 Merging... ").Bold().Foreground(CProfile.Color("#8BC34A")))

	err = d.mergeFiles(sections)
	fmt.Printf("\n\n ✅ Download completed in %v seconds\n", time.Since(now))

	return err
}

// Get a new http request
func (d Download) getNewRequest(method string) (*http.Request, error) {
	r, err := http.NewRequest(
		method,
		d.URL,
		nil,
	)
	if err != nil {
		return nil, err
	}
	r.Header.Set("User-Agent", "TDM")
	return r, nil
}
