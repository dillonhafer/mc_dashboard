package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

type MinecraftRunner interface {
	Run(cmd string) error
}

type DummyMinecraft struct {
	Logger io.Writer
}

func (dm *DummyMinecraft) Run(cmd string) error {
	_, err := fmt.Fprintf(dm.Logger, "DummyMinecraft: %v", cmd)
	return err
}

type Minecraft struct {
	Screen string
}

type MinecraftCmd struct {
	Command string
}

type MinecraftCmdCollection struct {
	Pool map[string]MinecraftCmd
}

func (mc *MinecraftCmdCollection) FromJson(jsonStr []byte) error {
	var data = &mc.Pool
	return json.Unmarshal(jsonStr, data)
}

func (m *Minecraft) Run(cmd string) error {
	return exec.Command("screen", "-S", m.Screen, "-p", "0", "-X", "stuff", fmt.Sprintf(`/%s\r`, cmd)).Run()
}

func FindCmd(cmdName string) string {
	mc := new(MinecraftCmdCollection)
	content, err := ioutil.ReadFile("commands.json")
	if err != nil {
		fmt.Print("Error:", err)
	}

	mc.FromJson(content)
	return mc.Pool[cmdName].Command
}

func main() {
	var screen string
	flag.StringVar(&screen, "screen", "", "runs commands in specific screen")
	flag.Parse()

	var mc MinecraftRunner

	if screen == "" {
		mc = &DummyMinecraft{Logger: os.Stderr}
	} else {
		mc = &Minecraft{Screen: screen}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Command Must be a POST\n")
			return
		}

		cmd := FindCmd(r.URL.Path)

		err := mc.Run(cmd)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
	})

	err := http.ListenAndServe(":8000", nil)
	fmt.Fprintln(os.Stderr, err)
}
