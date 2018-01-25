package main

import (
	"fmt"
	"os/exec"
	"strings"
	"bytes"
	"regexp"
	"log"
)

type Vbox struct {
	tool    string
	verbose bool
}

func NewVbox(tool string, verbose bool) *Vbox {
	return &Vbox{tool: tool, verbose: verbose}
}

func (vbox *Vbox) Output(params []string) []byte {
	return executeNormal(vbox.tool, params, vbox.verbose )
}

func (vbox *Vbox) OutputString(params []string) string {
	return string(vbox.Output(params))
}

func (vbox *Vbox) Run(params []string) {
	log := vbox.Output(params)
	fmt.Printf("%s\n", log)
}

func (vbox *Vbox) Command(params []string) {
	vbox.Run(params)
}

func (vbox *Vbox) CommandForce(params []string) {
	log := executeForce(vbox.tool, params, vbox.verbose)
	fmt.Printf("%s\n", log)
}

func (vbox *Vbox) StartVm(vmName string) {
	vbox.Run([]string{"startvm", vmName, "--type", "headless"})
}

func (vbox *Vbox) StartVmGui(vmName string) {
	vbox.Run([]string{"startvm", vmName, "--type", "gui"})
}

func (vbox *Vbox) StopVm(vmName string) {
	vbox.Run([]string{"controlvm", vmName, "poweroff"})
}
func (vbox *Vbox) SendKeyToVm(vmName string,KeyCode []string) {
	cmd := []string{"controlvm", vmName, "keyboardputscancode"}
	cmd = append(cmd,KeyCode...)

	vbox.Run(cmd)
}

func (vbox *Vbox) AllVms() map[string]string {
	return vbox.getVmList([]string{"vms"})
}

func (vbox *Vbox) RunningVms() map[string]string {
	return vbox.getVmList([]string{"runningvms"})
}

// Unfortunately VBoxManage.exe returns != 0 when help command executed
func (vbox *Vbox) Help(params []string) {
	passed := []string{"help"}
	vbox.CommandForce(append(passed, params...))
}

// "NX" {b53046d9-9f2c-41ef-945b-806a8bc6a032} みたいなのが出る
func (vbox *Vbox) getVmList(params []string) map[string]string {
	ret := map[string]string{}
	log := vbox.OutputString(append([]string{"list"}, params...))
	for _, entry := range strings.Split(log, "\n") {
		if len(entry) == 0 {
			continue
		}
		name, hash := parseVmEntryLog(entry)
		ret[name] = hash
	}
	return ret
}

func parseVmEntryLog(text string) (string, string) {
	body := strings.Replace(text, "\"", "", -1)
	params := strings.Split(body, " ")
	return params[0], params[1]
}

func execute(tool string, params []string, debug bool, handleError bool) []byte {

	var stdout, stderr bytes.Buffer

		fmt.Printf(" >> %v %v\n", tool, params)


	cmd := exec.Command(tool, params...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("VBoxManage error: %s", stderrString)
	}

	if err == nil {
		// Sometimes VBoxManage gives us an error with a zero exit code,
		// so we also regexp match an error string.
		m, _ := regexp.MatchString("VBoxManage([.a-z]+?): error:", stderrString)
		if m {
			err = fmt.Errorf("VBoxManage error: %s", stderrString)
		}
	}

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	return nil

}

func executeForce(tool string, params []string, debug bool) []byte {
	return execute(tool, params, debug, false)
}

func executeNormal(tool string, params []string, debug bool) []byte {
	return execute(tool, params, debug, true)
}