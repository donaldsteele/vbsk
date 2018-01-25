package main
import (
	"fmt"
	"github.com/urfave/cli"
	"os"
	"path/filepath"

	"strings"

	"time"


	"log"

)

var DEFAULT_TOOL_PATH = fmt.Sprintf("C:%cProgram Files%cOracle%cVirtualBox%cVBoxManage.exe", filepath.Separator, filepath.Separator, filepath.Separator, filepath.Separator)
var VM_NAME = "MyVM"
var KEYBOARD_DELAY = 50

type kbBuffer struct {
	buffer    []string

}

const keyboardDelayOption = "delay"

const vmNameOption = "vm"
const toolPathOption = "tool"

var OptionFlags = []cli.Flag{
	cli.StringFlag{
		Name:  fmt.Sprintf("%s, t", toolPathOption),
		Value: DEFAULT_TOOL_PATH,
		Usage: "specify path to VBoxManage.exe",
	},
	cli.StringFlag{
		Name:  fmt.Sprintf("%s, m", vmNameOption),
		Value: vmNameOption,
		Usage: "the name of the vm you wish to control",
	},
	cli.BoolFlag{
		Name:  "verbose, V",
		Usage: "verbose mode",
	},
	cli.IntFlag{
		Name:  fmt.Sprintf("%s, d", keyboardDelayOption),
		Value: KEYBOARD_DELAY,
		Usage: "the delay to send keyboard input in miliseconds",
	},
}

var CommandList = []cli.Command{
	{
		Name:    "now",
		Aliases: []string{},
		Usage:   "list current VM status",
		Action:  cmdNow,
	},
	{
		Name:    "cmd",
		Aliases: []string{},
		Usage:   "send command on console to vm",
		Action:  cmdExecute,
	},
}


func NewkbBuffer() *kbBuffer {
	return &kbBuffer{}
}

func getGlobalContext(c *cli.Context) *cli.Context {
	parent := c.Parent()
	if parent != nil {
		return parent
	} else {
		return c
	}
}
func loadVbox(c *cli.Context) *Vbox {
	ctx := getGlobalContext(c)
	return NewVbox(ctx.String(toolPathOption), ctx.Bool("verbose"))
}

func loadKeyboard(c *cli.Context) *keyboard {
	return Newkeyboard()

}

func cmdNow(c *cli.Context) error {
	vbox := loadVbox(c)
	all := vbox.AllVms()
	running := vbox.RunningVms()
	space := 0
	for k, _ := range all {
		if space < len(k) {
			space = len(k)
		}
	}
	fmt.Println("\nVM status:")
	for k, _ := range all {
		if _, exists := running[k]; exists {
			fmt.Printf(fmt.Sprintf("%%%ds: Run\n", space+1), k)
		} else {
			fmt.Printf(fmt.Sprintf("%%%ds: stop\n", space+1), k)
		}
	}
	return nil
}

func cmdExecute(c *cli.Context) error {

	vbox := loadVbox(c)
	kbd := loadKeyboard(c)
	scancodes := kbd.scancodes(strings.Join(c.Args()," "))
	ctx := getGlobalContext(c)
	vmName := ctx.String(vmNameOption)
	/* fmt.Printf("%v", scancodes) */

	kb := NewkbBuffer()

		for _, code := range scancodes {
			if code == "wait" {
				sendbuffer(kb,vbox,vmName)
				time.Sleep(1 * time.Second)
				continue
			}

			if code == "wait5" {
				sendbuffer(kb,vbox,vmName)
				time.Sleep(5 * time.Second)
				continue
			}

			if code == "wait10" {
				sendbuffer(kb,vbox,vmName)
				time.Sleep(10 * time.Second)
				continue
			}

			kb.buffer = append(kb.buffer,code)
			//if the length of the buffer is larger than 8 characters send what we already have.
            if len(kb.buffer) > 14 {
				fmt.Printf("hit the 7 char max, sending buffer now")
				sendbuffer(kb,vbox,vmName)
				time.Sleep(time.Millisecond * time.Duration(ctx.Int(keyboardDelayOption))) //sleep 50 miliseconds to simulate typing
			}



			fmt.Printf("code is => %v\n", code)
			//vbox.SendKeyToVm(ctx.String(vmNameOption),code)
//            time.Sleep(time.Millisecond * time.Duration(ctx.Int(keyboardDelayOption))) //sleep 50 miliseconds to simulate typing

		/*	if err := driver.VBoxManage("controlvm", vmName, "keyboardputscancode", code); err != nil {
				err := fmt.Errorf("Error sending boot command: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return nil
			}
		*/
		}

	sendbuffer(kb,vbox,vmName)

	return nil
}

func sendbuffer(k *kbBuffer,vbox *Vbox, vmName string) {

	if  len(k.buffer) > 0 {
		log.Print("Sending Buffer")
		vbox.SendKeyToVm(strings.Trim(vmName, " "),k.buffer)
		k.buffer = k.buffer[:0]
	}
}


func main() {
	app := cli.NewApp()
	app.Name = "vbox"
	app.Usage = "Virtual Box operation Tool"
	app.Version = "1.0.0"
	app.Commands = CommandList
	app.Action = nil
	app.Flags = OptionFlags

	app.Run(os.Args)
}
