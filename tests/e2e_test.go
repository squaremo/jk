package tests

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func find(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		switch {
		case err != nil:
			return err
		case info.IsDir():
			return nil
		case strings.HasSuffix(path, "~"):
			return nil
		}
		files = append(files, strings.TrimPrefix(path, dir))
		return nil
	})

	return files, err
}

func basename(testFile string) string {
	ext := filepath.Ext(testFile)
	return testFile[:len(testFile)-len(ext)]
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func shouldErrorOut(testFile string) bool {
	return exists(testFile + ".error")
}

func shouldSkip(testFile string) bool {
	return exists(testFile + ".skip")
}

func invocation(testFile, outputDir string) []string {
	data, err := ioutil.ReadFile(testFile + ".cmd")
	if err != nil {
		return nil
	}
	content := strings.TrimSuffix(string(data), "\n")

	parts := strings.Split(content, " ")
	// Strip jk from the cmd line, it's added later
	if len(parts) > 0 && parts[0] == "jk" {
		parts = parts[1:]
	}
	if len(parts) == 0 {
		return nil
	}

	replacer := strings.NewReplacer(
		"%d", outputDir,
		"%t", testFile[:len(testFile)-3],
	)
	// Replace special strings
	for i := range parts {
		parts[i] = replacer.Replace(parts[i])
	}

	return parts
}

func runTest(t *testing.T, file string) {
	base := basename(file)
	expectedDir := base + ".expected"
	gotDir := base + ".got"

	if shouldSkip(file) {
		return
	}

	cmdline := invocation(file, gotDir)
	if cmdline == nil {
		cmdline = []string{"run", "-o", gotDir, file}
	}
	cmd := exec.Command("jk", cmdline...)
	output, err := cmd.CombinedOutput()

	// 0. Check process exit code.
	if shouldErrorOut(file) {
		_, ok := err.(*exec.ExitError)
		assert.True(t, ok)
	} else {
		assert.NoError(t, err)
	}

	// 1. Compare stdout/err.
	expected, _ := ioutil.ReadFile(file + ".expected")
	assert.Equal(t, string(expected), string(output))

	// 2. Compare produced files.
	expectedFiles, _ := find(expectedDir)
	gotFiles, _ := find(gotDir)

	// 2. a) Compare the list of files.
	if !assert.Equal(t, expectedFiles, gotFiles) {
		assert.FailNow(t, "generated files not equivalent; bail")
	}

	// 2. b) Compare file content.
	for i := range expectedFiles {
		expected, err := ioutil.ReadFile(expectedDir + expectedFiles[i])
		assert.NoError(t, err)
		got, err := ioutil.ReadFile(gotDir + gotFiles[i])
		assert.NoError(t, err)

		assert.Equal(t, string(expected), string(got))
	}
}

func listTestFiles(t *testing.T) []string {
	// Some tests aren't actually in this directory, but a .cmd file is used to
	// tune how jk is run. We need to account for those, making sure tests with
	// both a test-*.js file and a .cmd file aren't run twice.
	cmds, err := filepath.Glob("test-*.js.cmd")
	assert.NoError(t, err)

	files, err := filepath.Glob("test-*.js")
	assert.NoError(t, err)

	for _, cmd := range cmds {
		// Remove .cmd extension
		files = append(files, cmd[:len(cmd)-4])
	}

	// Deduplicate test files
	unique := make(map[string]struct{})
	for _, key := range files {
		unique[key] = struct{}{}
	}

	files = make([]string, 0, len(unique))
	for key := range unique {
		files = append(files, key)
	}

	sort.Strings(files)
	return files
}

func TestEndToEnd(t *testing.T) {
	files := listTestFiles(t)

	for _, file := range files {
		t.Run(file[:len(file)-3], func(t *testing.T) {
			runTest(t, file)
		})
	}
}
