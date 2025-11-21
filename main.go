package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Path struct {
	HCTL   string // 3:0:0
	Device string // sd*
	Major  string // 8:16
	Status string // active
	State  string // ready running
}

type LUN struct {
	Name   string // mpatha
	WWID   string // 3600..
	Device string // dm-0
	Vendor string // LIO-ORG,lun01
	Paths  []Path // список путей
}

func main() {
	var command string
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	if command == "man" {
		printMan()
		return
	}
	luns := parseMultipath()

	if command == "show" {
		printShow(luns)
		return
	}
}

func printMan() {
	fmt.Println("lunctl - Tools управления лунами ")
	fmt.Println("Доступные ключи")
	fmt.Println("man - Выводит мануал управления ")
	fmt.Println("show - Выводит подробный данные luns ")
}

func printShow(luns []LUN) {
	for _, lun := range luns {
		fmt.Printf("----------------------------------------------")
		fmt.Printf("\nLUN: %s\n", lun.Name)
		fmt.Printf("  WWID: %s\n", lun.WWID)
		fmt.Printf("  Device: %s\n", lun.Device)
		fmt.Printf("  Vendor: %s\n", lun.Vendor)
		fmt.Printf("  Total paths: %d\n", len(lun.Paths))

		for i, path := range lun.Paths {
			fmt.Printf("    Path %d: %s (%s)\n", i+1, path.Device, path.Status)

		}
	}
}

func parseMultipath() []LUN {
	var luns []LUN
	var currentLUN *LUN

	cmd := exec.Command("multipath", "-ll")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(output), "\n")
	pathCount := 0

	for _, line := range lines {
		if strings.HasPrefix(line, "mpath") {
			if currentLUN != nil {
				luns = append(luns, *currentLUN)
			}
			parts := strings.Fields(line)
			currentLUN = &LUN{
				Name:   parts[0],
				WWID:   strings.Trim(parts[1], "()"),
				Device: parts[2],
				Vendor: parts[3],
				Paths:  []Path{},
			}
			pathCount = 0
		}
		if strings.Contains(line, "|-") || strings.Contains(line, "`-") {
			parts := strings.Fields(line)

			for i, part := range parts {
				if strings.HasPrefix(part, "sd") || strings.HasPrefix(part, "vd") || strings.HasPrefix(part, "nd") || strings.HasPrefix(part, "nvme") {
					path := Path{
						Device: part,
					}
					if i+2 < len(parts) {
						path.Status = parts[i+2]
					}
					currentLUN.Paths = append(currentLUN.Paths, path)
					pathCount++
					break
				}
			}
		}
	}

	if currentLUN != nil {
		luns = append(luns, *currentLUN)
	}

	return luns
}
