package package_parser

import (
	"fmt"
	"github.com/tadnir/goop/utils"
	"log"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type GoPackage struct {
	packageName  string
	packageFiles map[string]*GoFile
}

func isFileGenerated(path string) (bool, error) {
	myFile, err := os.Open(path)
	if err != nil {
		log.Fatalf("impossible to open file: %s", err)
	}

	defer myFile.Close()
	generatedString := "// Code generated"
	buf := make([]byte, len(generatedString))
	n, err := myFile.Read(buf)
	if err != nil {
		return false, fmt.Errorf("impossible to read file: %s", err)
	}
	if n != len(generatedString) {
		return false, nil
	}

	return string(buf) == generatedString, nil
}

func ParsePackage(packageName string, packagePath string, ignoreGenerated bool) (*GoPackage, error) {
	pack := &GoPackage{packageName: packageName, packageFiles: map[string]*GoFile{}}
	entries, err := os.ReadDir(packagePath)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".go") || e.IsDir() {
			continue
		}

		if ignoreGenerated {
			isGen, err := isFileGenerated(filepath.Join(packagePath, e.Name()))
			if err != nil {
				return nil, err
			}

			if isGen {
				fmt.Printf("Skipping generated file %s in package \"%s\".\n", e.Name(), packageName)
				continue
			}
		}

		packFile, err := ParsePackageFile(packagePath, e.Name())
		if err != nil {
			return nil, err
		}

		if packFile.packageName != pack.packageName {
			return nil, fmt.Errorf("package %s contains multiple packages definitions", packagePath)
		}

		pack.packageFiles[e.Name()] = packFile
	}

	return pack, nil
}

func (pack *GoPackage) GetName() string {
	return pack.packageName
}

func (pack *GoPackage) GetFiles() []*GoFile {
	return slices.Collect(maps.Values(pack.packageFiles))
}

func (pack *GoPackage) GetFile(fileName string) (*GoFile, error) {
	f, ok := pack.packageFiles[fileName]
	if !ok {
		return nil, fmt.Errorf("file %s not found in package %s", fileName, pack.packageName)
	}

	return f, nil
}

func (pack *GoPackage) GetStructs() []*Struct {
	return slices.Concat(utils.Map(maps.Values(pack.packageFiles), (*GoFile).GetStructs)...)
}

func (pack *GoPackage) String() string {
	return fmt.Sprintf("Package: %v\n", pack.packageName) +
		strings.Join(utils.Map(slices.Values(pack.GetFiles()), (*GoFile).String), "\n")
}
