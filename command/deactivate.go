package command

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/Jabba-Team/jabba/cfg"
)

func Deactivate() ([]string, error) {
	pth, _ := os.LookupEnv("PATH")
	pathSep := regexp.QuoteMeta(string(os.PathListSeparator))
	// Match jabba paths with optional trailing separator (for paths at end of PATH)
	rgxp := regexp.MustCompile(regexp.QuoteMeta(filepath.Join(cfg.Dir(), "jdk")) + "[^" + pathSep + "]+" + "[" + pathSep + "]?")
	// strip references to ~/.jabba/jdk/*, otherwise leave unchanged
	pth = rgxp.ReplaceAllString(pth, "")
	// Clean up any trailing path separator
	pth = regexp.MustCompile(pathSep+"$").ReplaceAllString(pth, "")
	javaHome, overrideWasSet := os.LookupEnv("JAVA_HOME_BEFORE_JABBA")
	if !overrideWasSet {
		javaHome, _ = os.LookupEnv("JAVA_HOME")
	}
	return []string{
		"export PATH=\"" + pth + "\"",
		"export JAVA_HOME=\"" + javaHome + "\"",
		"unset JAVA_HOME_BEFORE_JABBA",
	}, nil
}
