package main

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
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

	pkgs.scripts = getPythonScriptsFromDisk(".")

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

func getPythonScriptsFromDisk(path string) []pythonScript {
	var scripts []pythonScript

	files, err := os.ReadDir(path)
	if err != nil {
		return scripts
	}

	for _, file := range files {
		fullPath := filepath.Join(path, file.Name())

		if file.IsDir() && file.Name() != ".venv" {
			scripts = append(scripts, getPythonScriptsFromDisk(fullPath)...)
			continue
		}

		if !strings.HasSuffix(file.Name(), ".py") {
			continue
		}

		nfile, err := os.Open(fullPath)
		if err != nil {
			continue
		}
		defer nfile.Close()

		scanner := bufio.NewScanner(nfile)
		var flen, funcCount, classCount int
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			flen++
			if strings.HasPrefix(line, "def ") {
				funcCount++
			}
			if strings.HasPrefix(line, "class ") {
				classCount++
			}
		}

		if err := scanner.Err(); err == nil {
			scripts = append(scripts, pythonScript{
				path:      fullPath,
				lines:     flen,
				functions: funcCount,
				classes:   classCount,
			})
		}
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

type InstallResponseObject struct {
	content string
	isErr   bool
}

func runInstallCommandAndRespond(command string, pkg string) InstallResponseObject {
	var obj InstallResponseObject

	var cmd = exec.Command(command, pkg)

	if command == "uv" {
		cmd = exec.Command(command, "add", pkg)
	}
	var stdout, _ = cmd.StdoutPipe()
	var stderr, _ = cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		obj.content = err.Error()
		obj.isErr = true
		return obj
	}

	outBytes, _ := io.ReadAll(stdout)
	errBytes, _ := io.ReadAll(stderr)

	if err := cmd.Wait(); err != nil {
		obj.content = err.Error()
		obj.isErr = true
		return obj

	}

	if len(errBytes) > 0 {
		obj.content = string(errBytes)
		obj.isErr = true
	} else {
		obj.content = string(outBytes)

	}
	return obj
}
