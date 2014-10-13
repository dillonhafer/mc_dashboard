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
	_, err := fmt.Fprintf(dm.Logger, "DummyMinecraft: %v\n", cmd)
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

func commandApi(mc MinecraftRunner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Command Must be a POST\n")
			return
		}

		cmd := FindCmd(r.URL.Path[4:])
		err := mc.Run(cmd)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
	}
}

func main() {
	var screen string
	var mc MinecraftRunner
	staticFiles := http.FileServer(http.Dir("public"))

	flag.StringVar(&screen, "screen", "", "runs commands in specific screen")
	flag.Parse()

	if screen == "" {
		mc = &DummyMinecraft{Logger: os.Stderr}
	} else {
		mc = &Minecraft{Screen: screen}
	}

	http.Handle("/", staticFiles)
	http.HandleFunc("/api/", commandApi(mc))

	fmt.Println("Server running and listening on port 8000")
	fmt.Println("Run `mc_dashboard -h` for more startup options")
	fmt.Println("Ctrl-C to shutdown server")
	err := http.ListenAndServe(":8000", nil)
	fmt.Fprintln(os.Stderr, err)
}
