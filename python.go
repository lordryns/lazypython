package main

import (
	"bufio"
	"github.com/pelletier/go-toml/v2"
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

	var uvConfig = readTomlFile()
	var uvPackageStrings []pythonPackage

	for _, pkgS := range uvConfig.Project.Dependencies {
		var sS = strings.Split(pkgS, ">=")
		if len(sS) > 1 {
			uvPackageStrings = append(uvPackageStrings,
				pythonPackage{path: strings.TrimSpace(sS[0]), version: sS[1]},
			)
		}
	}

	for _, s := range sres {
		var split = strings.Split(s, "==")
		if len(split) > 1 {
			pkgs.packages = append(pkgs.packages, pythonPackage{path: split[0], version: split[1]})
		}
	}

	var checkIfPackageExists = func(pkg string) bool {
		for _, p := range pkgs.packages {
			if p.path == pkg {
				return true
			}
		}

		return false
	}
	for _, pkg := range uvPackageStrings {
		if checkIfPackageExists(pkg.path) {
			continue
		}

		pkgs.packages = append(pkgs.packages, pkg)
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

type Config struct {
	Project struct {
		Name           string
		Version        string
		Description    string
		Readme         string
		RequiresPython string `toml:"requires-python"`
		Dependencies   []string
	}
	DependencyGroups map[string][]string `toml:"dependency-groups"`
	Tool             struct {
		Uv struct {
			IndexURL   string `toml:"index-url"`
			PythonPath string `toml:"python-path"`
			CacheDir   string `toml:"cache-dir"`
		}
	}
}

func readTomlFile() Config {
	// i'm not as familiar with uv so i'll just put this check here in case
	var tomlName string
	var _, err = os.Stat("uv.toml")
	if !os.IsNotExist(err) {
		tomlName = "uv.toml"
	} else {
		tomlName = "pyproject.toml"
	}
	data, err := os.ReadFile(tomlName)

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {

	}
	return cfg
}
