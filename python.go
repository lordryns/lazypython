package main

import (
	"os/exec"
	"strings"
)

func generatePackageDetails() (pythonManager, error) {
	var pkgs pythonManager
	var cmd = exec.Command("pip", "freeze")
	var fres, err = cmd.Output()
	if err != nil {
		return pkgs, err
	}

	var sres = strings.Split(string(fres), "\n")

	for _, s := range sres {
		var split = strings.Split(s, "==")
		if len(split) > 1 {
			pkgs.packages = append(pkgs.packages, pythonPackage{path: split[0], version: split[1]})

		}
	}

	return pkgs, nil
}

func getPythonVersion() string {
	var cmd = exec.Command("python", "--version")

	var output, err = cmd.Output()
	if err != nil {
		return "NONE"
	}

	return string(output)
}
