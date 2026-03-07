package command

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/Jabba-Team/jabba/cfg"
)

func Use(selector string) ([]string, error) {
	aliasValue := GetAlias(selector)
	if aliasValue != "" {
		selector = aliasValue
	}
	ver, err := LsBestMatch(selector)
	if err != nil {
		return nil, err
	}
	return usePath(filepath.Join(cfg.Dir(), "jdk", ver))
}

func usePath(path string) ([]string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	pth, _ := os.LookupEnv("PATH")
	pathSep := regexp.QuoteMeta(string(os.PathListSeparator))
	// Match jabba paths with optional trailing separator (for paths at end of PATH)
	rgxp := regexp.MustCompile(regexp.QuoteMeta(filepath.Join(cfg.Dir(), "jdk")) + "[^" + pathSep + "]+" + "[" + pathSep + "]?")
	// strip references to ~/.jabba/jdk/*, otherwise leave unchanged
	pth = rgxp.ReplaceAllString(pth, "")
	// Clean up any trailing path separator
	pth = regexp.MustCompile(pathSep+"$").ReplaceAllString(pth, "")
	if runtime.GOOS == "darwin" {
		path = filepath.Join(path, "Contents", "Home")
	}
	systemJavaHome, overrideWasSet := os.LookupEnv("JAVA_HOME_BEFORE_JABBA")
	if !overrideWasSet {
		systemJavaHome, _ = os.LookupEnv("JAVA_HOME")
	}
	newPath := filepath.Join(path, "bin") + string(os.PathListSeparator) + pth
	// Clean up any trailing path separator from final result
	newPath = regexp.MustCompile(pathSep+"$").ReplaceAllString(newPath, "")
	return []string{
		"export PATH=\"" + newPath + "\"",
		"export JAVA_HOME=\"" + path + "\"",
		"export JAVA_HOME_BEFORE_JABBA=\"" + systemJavaHome + "\"",
	}, nil
}
