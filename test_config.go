package main

import (
	"fmt"
	"github.com/TheShellMaster/zivpn-desktop/internal/config"
)

func main() {
	c := config.Connection{Server: "1.1.1.1", Port: 5667, Password: "test"}
	path, _ := c.WriteEngineConfig(".")
	fmt.Println(path)
}
