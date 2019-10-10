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
		data, _ := ioutil.ReadFile(vaultFile)
		vaultPass = string(data)
	}

	os.Stderr.WriteString(fmt.Sprintf("Using vault file: %s\n", vaultPath))
	contents, _ := vault.DecryptFile(vaultPath, vaultPass)

	var data map[string]string

	if err := yaml.Unmarshal([]byte(contents), &data); err != nil {
		fmt.Println(err)
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

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

	if !cmd.ProcessState.Success() {
		fmt.Println("failed to run")
	}

}
