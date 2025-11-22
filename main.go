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
	Size   string // SIZE.. Luns
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
	if command == "list" {
		printList(luns)
	}
}

func printMan() {
	fmt.Printf("%-8s | %-40s\n", "Key", "Value")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("%-8s | %-40s\n", "lunctl", "Tools управления лунами")
	fmt.Println("Доступные ключи")
	fmt.Println("man - Выводит мануал управления")
	fmt.Println("show - Выводит подробный данные luns")
	fmt.Println("list")
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
		// Парсим основную строку LUN
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
				Size:   "", // ← ИСПРАВЛЕНО: пустая строка, заполним потом
				Paths:  []Path{},
			}
			pathCount = 0
		} else if currentLUN != nil && strings.Contains(line, "size=") {
			// Парсим строку с размером
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "size=") {
					currentLUN.Size = strings.TrimPrefix(part, "size=")
					break
				}
			}
		} else if currentLUN != nil && (strings.Contains(line, "|-") || strings.Contains(line, "`-")) {
			// ← ИСПРАВЛЕНО: добавил else if, теперь это внутри цикла!
			parts := strings.Fields(line)
			for i, part := range parts {
				if strings.HasPrefix(part, "sd") || strings.HasPrefix(part, "vd") ||
					strings.HasPrefix(part, "nd") || strings.HasPrefix(part, "nvme") {
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
	} // ← Закрывающая скобка цикла for

	// Добавляем последний LUN
	if currentLUN != nil {
		luns = append(luns, *currentLUN)
	}

	return luns
}

func printShow(luns []LUN) {
	for _, lun := range luns {
		fmt.Printf("----------------------------------------------")
		fmt.Printf("\nLUN: %s\n", lun.Name)
		fmt.Printf("  WWID: %s\n", lun.WWID)
		fmt.Printf("  Size: %s\n", lun.Size)
		fmt.Printf("  Device: %s\n", lun.Device)
		fmt.Printf("  Vendor: %s\n", lun.Vendor)
		fmt.Printf("  Total paths: %d\n", len(lun.Paths))

		for i, path := range lun.Paths {
			fmt.Printf("    Path %d: %s (%s)\n", i+1, path.Device, path.Status)

		}
	}
}

func printList(luns []LUN) {
	fmt.Printf("%-8s | %-40s | %-5s | %-6s\n", "Name", "WWID", "Paths", "Size")
	fmt.Println(strings.Repeat("-", 70))

	for _, lun := range luns {
		fmt.Printf("%-8s | %-40s | %-5d | %-6s\n",
			lun.Name, lun.WWID, len(lun.Paths), lun.Size)
	}
}
