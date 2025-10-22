package main

import (
	"bufio"
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
