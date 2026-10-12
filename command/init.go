package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jabba-Team/jabba/cfg"
)

func Init() error {
	shell := detectShell()
	integrationFile := getIntegrationFile(shell)
	rcFile := getRCFile(shell)

	// Create integration file if it doesn't exist
	if _, err := os.Stat(integrationFile); os.IsNotExist(err) {
		if err := createIntegrationFile(shell, integrationFile); err != nil {
			return fmt.Errorf("failed to create integration file: %w", err)
		}
		fmt.Printf("✓ Created %s\n", integrationFile)
	} else {
		fmt.Printf("✓ Integration file already exists: %s\n", integrationFile)
	}

	// Update RC file if needed
	if rcFile != "" {
		if !checkRCFileHasSource(rcFile, integrationFile) {
			if err := updateRCFile(shell, rcFile, integrationFile); err != nil {
				return fmt.Errorf("failed to update RC file: %w", err)
			}
			fmt.Printf("✓ Updated %s\n", rcFile)
		} else {
			fmt.Printf("✓ RC file already configured: %s\n", rcFile)
		}
	}

	fmt.Println("\nShell integration configured successfully!")
	fmt.Println("Restart your shell or run:")
	if shell == "powershell" {
		fmt.Printf("  . \"%s\"\n", integrationFile)
	} else {
		fmt.Printf("  source \"%s\"\n", integrationFile)
	}

	return nil
}

func createIntegrationFile(shell, integrationFile string) error {
	// Ensure directory exists
	dir := filepath.Dir(integrationFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var content string
	jabbaHome := cfg.Dir()

	switch shell {
	case "powershell":
		content = generatePowerShellIntegration(jabbaHome)
	case "fish":
		content = generateFishIntegration(jabbaHome)
	case "nushell":
		content = generateNushellIntegration(jabbaHome)
	default:
		content = generateBashIntegration(jabbaHome)
	}

	return os.WriteFile(integrationFile, []byte(content), 0644)
}

func generateBashIntegration(jabbaHome string) string {
	return fmt.Sprintf(`# https://github.com/Jabba-Team/jabba
# This file is intended to be "sourced" (i.e. ". ~/.jabba/jabba.sh")

export JABBA_HOME="%s"

jabba() {
    local fd3=$(mktemp /tmp/jabba-fd3.XXXXXX)
    
    # Detect if we're on Windows (Git Bash/MSYS/Cygwin) or Unix/Linux/macOS
    case "$OSTYPE" in
        msys*|cygwin*|win32)
            # Windows: use .exe extension and --fd3 flag (fd redirection doesn't work)
            JABBA_SHELL_INTEGRATION=ON "$JABBA_HOME/bin/jabba.exe" "$@" --fd3 "$fd3"
            local exit_code=$?
            if [ -s "$fd3" ]; then
                # Convert Windows paths to Unix-style for Git Bash
                # 1. Replace backslashes with forward slashes
                # 2. Convert C:/ to /c/ (drive letters)
                # 3. Replace semicolons with colons (PATH separator)
                eval $(cat "$fd3" | sed 's#\\#/#g' | sed 's#\([A-Za-z]\):/#/\L\1/#g' | sed 's#;#:#g')
            fi
            ;;
        *)
            # Unix/Linux/macOS: use fd redirection (standard approach)
            (JABBA_SHELL_INTEGRATION=ON $JABBA_HOME/bin/jabba "$@" 3>| ${fd3})
            local exit_code=$?
            eval $(cat ${fd3})
            ;;
    esac
    
    rm -f ${fd3}
    return ${exit_code}
}

if [ ! -z "$(jabba alias default)" ]; then
    jabba use default
fi
`, jabbaHome)
}

func generatePowerShellIntegration(jabbaHome string) string {
	return fmt.Sprintf(`# https://github.com/Jabba-Team/jabba
# This file is intended to be "sourced" (i.e. ". ~/.jabba/jabba.ps1")

$env:JABBA_HOME = "%s"

function jabba {
    $fd3 = [IO.Path]::GetTempFileName()
    # Do NOT pipe to Out-Null — that would suppress stdout output from
    # query commands like 'jabba current', 'jabba ls', etc.
    & $env:JABBA_HOME\bin\jabba.exe $args --fd3 $fd3
    $exitCode = $LASTEXITCODE
    if (Test-Path $fd3) {
        $cmd = Get-Content $fd3 | Out-String
        if ($cmd) {
            # jabba.exe writes bash-style shell commands to fd3:
            #   export PATH="..."
            #   export JAVA_HOME="..."
            #   unset JAVA_HOME_BEFORE_JABBA
            # Convert them to PowerShell equivalents before executing.
            #   export VAR="value"  ->  $env:VAR = "value"
            #   unset VAR           ->  Remove-Item Env:VAR
            $cmd = [regex]::Replace($cmd, 'export\s+(\w+)="([^"]*)"', '$$env:$1 = "$2"')
            $cmd = [regex]::Replace($cmd, 'unset\s+(\w+)', 'Remove-Item Env:$1 -ErrorAction SilentlyContinue')
            Invoke-Expression $cmd
        }
        Remove-Item $fd3
    }
    return $exitCode
}

if (jabba alias default) {
    jabba use default
}
`, jabbaHome)
}

func generateFishIntegration(jabbaHome string) string {
	return fmt.Sprintf(`# https://github.com/Jabba-Team/jabba
# This file is intended to be "sourced" (i.e. ". ~/.jabba/jabba.fish")

set -xg JABBA_HOME "%s"

function jabba
    set fd3 (mktemp /tmp/jabba-fd3.XXXXXX)
    env JABBA_SHELL_INTEGRATION=ON $JABBA_HOME/bin/jabba $argv 3> $fd3
    set exit_code $status
    eval (cat $fd3 | sed "s/^export/set -xg/g" | sed "s/=/ /g" | sed "s/:/\" \"/g" | sed "s/\"/\\\\\"/g" | sed "s/'\\\\''/'/g")
    rm -f $fd3
    return $exit_code
end

if test ! -z (jabba alias default)
    jabba use default
end
`, jabbaHome)
}

func generateNushellIntegration(jabbaHome string) string {
	return fmt.Sprintf(`# https://github.com/Jabba-Team/jabba
# This file is intended to be "sourced" (i.e. "source ~/.jabba/jabba.nu")

def --env jabba [...params:string] {
    $env.JABBA_HOME = '%s'
    let fd3 = mktemp -t jabba-fd3.XXXXXX.env
    nu -c $"$env.JABBA_SHELL_INTEGRATION = 'ON'
      ($env.JABBA_HOME)/bin/jabba ...($params) --fd3 ($fd3)"
    let exit_code = $env.LAST_EXIT_CODE
    if ( ls $fd3 | where size > 0B | is-not-empty ) {
       (
            open $fd3
            | str trim
            | lines
            | parse 'export {name}="{value}"'
            | transpose --header-row --as-record)| load-env
    }
    if $exit_code != 0 {
        return $exit_code
    }
}
`, jabbaHome)
}

func updateRCFile(shell, rcFile, integrationFile string) error {
	// Ensure directory exists
	dir := filepath.Dir(rcFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create file if it doesn't exist
	if _, err := os.Stat(rcFile); os.IsNotExist(err) {
		if err := os.WriteFile(rcFile, []byte(""), 0644); err != nil {
			return err
		}
	}

	// Read existing content
	content, err := os.ReadFile(rcFile)
	if err != nil {
		return err
	}

	var sourceLine string
	switch shell {
	case "powershell":
		sourceLine = fmt.Sprintf("\nif (Test-Path \"%s\") { . \"%s\" }\n", integrationFile, integrationFile)
	case "fish":
		sourceLine = fmt.Sprintf("\n[ -s \"%s\" ]; and source \"%s\"\n", integrationFile, integrationFile)
	case "nushell":
		sourceLine = fmt.Sprintf("\nsource '%s'\n", integrationFile)
	default:
		sourceLine = fmt.Sprintf("\n[ -s \"%s\" ] && source \"%s\"\n", integrationFile, integrationFile)
	}

	// Append source line
	newContent := string(content)
	if !strings.HasSuffix(newContent, "\n") && len(newContent) > 0 {
		newContent += "\n"
	}
	newContent += sourceLine

	return os.WriteFile(rcFile, []byte(newContent), 0644)
}
