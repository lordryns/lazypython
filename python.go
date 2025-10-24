package main

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
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

	pkgs.scripts = getPythonScriptsFromDisk()

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

func getPythonScriptsFromDisk() []pythonScript {
	var scripts []pythonScript
	var cwd, cerr = os.Getwd()
	if cerr != nil {
		return scripts
	}
	var files, err = os.ReadDir(cwd)
	if err != nil {
		return scripts
	}

	var filteredFiles []os.DirEntry

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".py") {
			filteredFiles = append(filteredFiles, file)
		}
	}

	for _, file := range filteredFiles {
		var flen int
		var funcCount int
		var classCount int
		var nfile, lerr = os.Open(filepath.Join(cwd, file.Name()))
		if lerr == nil {
			var scanner = bufio.NewScanner(nfile)
			for scanner.Scan() {
				flen += 1
				if strings.HasPrefix(scanner.Text(), "def") {
					funcCount += 1
				}

				if strings.HasPrefix(scanner.Text(), "class") {
					classCount += 1
				}
			}

			if err := scanner.Err(); err != nil {
				continue
			}
		}
		var s = pythonScript{path: file.Name(), lines: flen, functions: funcCount, classes: classCount}
		scripts = append(scripts, s)

	}

	return scripts
}
