package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

var re = regexp.MustCompile(`^([ ]*)([0-9]+)([ ]+)\*![ ]*(ENDIF|IF|LOOP|UNTIL|WHILE)(?: <(..)>)?$`)

type stackEntry struct {
	command  string
	label    string
	lineNum  int
	endLabel string
}

var labelNum int

func makeLabel() string {
	labelNum++
	return fmt.Sprintf("LABEL%d", labelNum)
}

func printLine(space1, line, space2, label, text string) (int, error) {
	return printLineN("5", space1, line, space2, label, text)
}

func printLine2(space1, line, space2, label, text string) (int, error) {
	return printLineN("7", space1, line, space2, label, text)
}

func printLineN(units, space1, line, space2, label, text string) (int, error) {
	out := fmt.Sprintf("%s%s%s%s", space1, line[:len(line)-1], units, space2)
	if label != "" {
		out = out + label
	}
	if text != "" {
		out = out + " " + text
	}
	return fmt.Printf("%s\n", out)
}

func branch(test string) string {
	switch test {
	case "CC":
		return "BCC"
	case "CS":
		return "BCS"
	case "EQ":
		return "BEQ"
	case "HS":
		return "BCS"
	case "LO":
		return "BCC"
	case "MI":
		return "BMI"
	case "NE":
		return "BNE"
	case "PL":
		return "BPL"
	}
	return "B??"
}

func branchOpposite(test string) string {
	switch test {
	case "CC":
		return "BCS"
	case "CS":
		return "BCC"
	case "EQ":
		return "BNE"
	case "HS":
		return "BCC"
	case "LO":
		return "BCS"
	case "MI":
		return "BPL"
	case "NE":
		return "BEQ"
	case "PL":
		return "BMI"
	}
	return "B??"
}

func process(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	var stack []stackEntry
	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		fmt.Println(line)
		groups := re.FindStringSubmatch(line)
		lastLabel := ""
		if groups == nil {
			lastLabel = ""
			continue
		}
		space1, line, space2, command, test := groups[1], groups[2], groups[3], groups[4], groups[5]
		_ = test
		last := &stackEntry{command: "(none)"}
		if len(stack) > 0 {
			last = &stack[len(stack)-1]
		}
		switch command {
		case "LOOP":
			if lastLabel == "" {
				lastLabel = makeLabel()
			}
			printLine(space1, line, space2, lastLabel, "")
			stack = append(stack, stackEntry{command: command, label: lastLabel, lineNum: lineNum})
		case "UNTIL":
			if last.command != "LOOP" {
				return fmt.Errorf("%d: %s found with no corresponding LOOP. Last command: %s/%d", lineNum, command, last.command, last.lineNum)
			}
			br := branchOpposite(test)
			cmd := fmt.Sprintf("%s %s", br, last.label)
			printLine(space1, line, space2, "", cmd)

			// If we had a WHILE, we need an endlabel printed
			if last.endLabel != "" {
				printLine2(space1, line, space2, last.endLabel, "")
				lastLabel = last.endLabel
			}
			stack = stack[:len(stack)-1]
		case "WHILE":
			if last.command != "LOOP" {
				return fmt.Errorf("%d: %s found with no corresponding LOOP. Last command: %s/%d", lineNum, command, last.command, last.lineNum)
			}
			if last.endLabel == "" {
				last.endLabel = makeLabel()
			}
			br := branchOpposite(test)
			cmd := fmt.Sprintf("%s %s", br, last.endLabel)
			printLine(space1, line, space2, "", cmd)
		case "IF":
			lastLabel = ""
			label := makeLabel()
			stack = append(stack, stackEntry{command: command, label: label, lineNum: lineNum})
			cmd := fmt.Sprintf("%s %s", branchOpposite(test), label)
			printLine(space1, line, space2, "", cmd)

		case "ENDIF":
			if last.command != "IF" {
				return fmt.Errorf("%d: %s found with no corresponding IF. Last command: %s/%d", lineNum, command, last.command, last.lineNum)
			}
			printLine(space1, line, space2, last.label, "")
			lastLabel = last.label
			stack = stack[:len(stack)-1]
		default:
			return fmt.Errorf("%d: unknown command: %s", lineNum, command)
		}

	}

	if err = scanner.Err(); err != nil {
		return err
	}

	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: starbang <input file>\n")
		os.Exit(1)
	}
	if err := process(os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
