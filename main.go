package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/ghodss/yaml"
	vault "github.com/sosedoff/ansible-vault-go"
)

var vaultPath string
var vaultPass string

var (
	version = "dev"
)

func main() {

	flag.StringVar(&vaultPath, "vault-path", "", "Path to ansible vault")
	flag.StringVar(&vaultPass, "vault-pass", "", "Vault password")
	flag.Parse()

	if os.Args[1] == "version" {
		fmt.Printf("version: %s", version)
		return
	}

	vaultFile := os.Getenv("ANSIBLE_VAULT_PASSWORD_FILE")
	if vaultPass == "" && vaultFile != "" {
		os.Stderr.WriteString(fmt.Sprintf("Using vault password file: %s\n", vaultFile))
		data, err := ioutil.ReadFile(vaultFile)
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("Unable to read vault file: %v\n", err))
			os.Exit(-1)
		}
		vaultPass = string(data)
	}

	os.Stderr.WriteString(fmt.Sprintf("Using vault file: %s\n", vaultPath))
	contents, err := vault.DecryptFile(vaultPath, vaultPass)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Unable to decrypt: %v\n", err))
		os.Exit(-1)
	}

	var data map[string]string

	if err := yaml.Unmarshal([]byte(contents), &data); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Unable to unmarshal: %v -> %s\n", err, contents))
		os.Exit(-1)

	}

	index := 0

	for i, item := range os.Args {
		if item == "--" {
			index = i
		}
	}

	args := os.Args[index+1:]

	environ := os.Environ()

	for k, v := range data {
		os.Stderr.WriteString(fmt.Sprintf("Setting key: %s\n", k))
		environ = append(environ, fmt.Sprintf("%s=%s", k, v))
	}

	cmd := exec.Command("bash", "-c", strings.Join(args, " "))

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = environ

	if err := cmd.Run(); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Error running: %v \n", err))
		os.Exit(-1)
	}

	if !cmd.ProcessState.Success() {
		os.Stderr.WriteString(fmt.Sprintf("Error running, success not returned: %v \n", err))
		os.Exit(-1)
	}

}
