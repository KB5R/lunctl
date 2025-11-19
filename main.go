package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	fmt.Println("=== LUN CTL ===")

	cmd := exec.Command("multipathd", "show", "maps", "format", "'%n;%w;%N;%S;%s'") // 'name %n ;uuid %w ;paths %N ;size %S ;vend/prod/rev %s'

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка multipathd: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if strings.Contains(line, "invalid") {
			continue
		}
		if strings.Contains(line, "name") {
			continue
		}
		line = strings.Trim(line, "'")
		parts := strings.Split(line, ";")
		fmt.Println("Name:", parts[0])
		fmt.Println("UUID:", parts[1])
		fmt.Println("Paths:", parts[2])
		fmt.Println("Size:", parts[3])
		fmt.Println("Vendor:", parts[4])
	}
}
