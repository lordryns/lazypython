package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/k3a/html2text"
)

var pythonPackages []string

func fetchPackagesFromIndex() {
	var response, err = http.Get("https://pypi.org/simple/")
	if err != nil {
		return
	}

	defer response.Body.Close()
	var scanner = bufio.NewScanner(response.Body)

	for scanner.Scan() {
		var plain = html2text.HTML2Text(scanner.Text())
		var split = strings.Split(plain, "/")
		if len(split) > 2 {
			pythonPackages = append(pythonPackages, split[2])
		}

	}

}

// had ai grab the important args to make this struct
type PackageInfo struct {
	Info struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Summary     string `json:"summary"`
		AuthorEmail string `json:"author_email"`
	} `json:"info"`

	Releases map[string][]struct {
		Size int `json:"size"`
	} `json:"releases"`

	Downloads struct {
		LastDay   int `json:"last_day"`
		LastWeek  int `json:"last_week"`
		LastMonth int `json:"last_month"`
	} `json:"downloads"`
}

func getPackageInfo(name string) PackageInfo {
	var pkg PackageInfo

	var resp, err = http.Get(fmt.Sprintf("https://pypi.org/pypi/%v/json", name))
	if err != nil {
		return pkg
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&pkg)

	var resp2, err2 = http.Get(fmt.Sprintf("https://pypistats.org/api/packages/%v/recent", name))
	defer resp.Body.Close()

	var stats struct {
		Data struct {
			LastDay   int `json:"last_day"`
			LastWeek  int `json:"last_week"`
			LastMonth int `json:"last_month"`
		} `json:"data"`
	}
	json.NewDecoder(resp2.Body).Decode(&stats)

	if err2 == nil {
		pkg.Downloads = stats.Data
	}
	return pkg
}
