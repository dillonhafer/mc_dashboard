package main

import (
	"flag"
	"fmt"
	"io"
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

func (m *Minecraft) Run(cmd string) error {
	// screen -S Minecraft -p 0 -X stuff "`printf "/weather rain\r"`"
	return exec.Command("screen", "-S", m.Screen, "-p", "0", "-X", "stuff", fmt.Sprintf(`printf "/%s\r"`, cmd)).Run()
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

	urlCmds := map[string]string{
		"/weather/rain": "weather rain",
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cmd, found := urlCmds[r.URL.Path]
		if !found {
			w.WriteHeader(http.StatusNotFound)
			return
		}

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
