package utils

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

func DisplayStatus(label string, value string, err error) {
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("[%v]\n%v\n\n", label, value)
}

func DisplayTable(labels []string, values []string, err error) {
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	labelsWidth := 0
	valuesWidth := 0

	for i, _ := range labels {
		if curLabel := utf8.RuneCountInString(labels[i]); curLabel > labelsWidth {
			labelsWidth = curLabel
		}

		if curValue := utf8.RuneCountInString(values[i]); curValue > valuesWidth {
			valuesWidth = curValue
		}
	}

	lineLabels := cAppend("-", labelsWidth+2)
	lineValues := cAppend("-", valuesWidth+2)

	fmt.Printf("+%v+%v+\n", lineLabels, lineValues)
	displayLineOfTable("Label", "Value", labelsWidth, valuesWidth)
	fmt.Printf("+%v+%v+\n", lineLabels, lineValues)
	for i, _ := range labels {
		displayLineOfTable(labels[i], values[i], labelsWidth, valuesWidth)
	}
	fmt.Printf("+%v+%v+\n\n", lineLabels, lineValues)
}

func DisplayHeaderTable(labels []string, values [][]string, err error) {
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	columnsWidth := make([]int, len(labels))

	for i, _ := range labels {
		if curLabel := utf8.RuneCountInString(labels[i]); curLabel > columnsWidth[i] {
			columnsWidth[i] = curLabel
		}

		for _, v := range values {
			if len(v) == len(columnsWidth) {
				if curValue := utf8.RuneCountInString(v[i]); curValue > columnsWidth[i] {
					columnsWidth[i] = curValue
				}
			}
		}
	}

	total := displaySeparator(columnsWidth)

	for i, _ := range columnsWidth {
		fmt.Printf("| %v%v ", labels[i], cAppend(" ", columnsWidth[i]-utf8.RuneCountInString(labels[i])))

	}
	fmt.Printf("|\n")

	displaySeparator(columnsWidth)

	if len(values) == 0 {
		fmt.Printf("| %v |\n", cAppend(" ", total))
	} else {
		for _, v := range values {
			for i, _ := range columnsWidth {
				if i < len(v) {
					if len(v) == len(columnsWidth) {
						fmt.Printf("| %v%v ", v[i], cAppend(" ", columnsWidth[i]-utf8.RuneCountInString(v[i])))
					} else {
						fmt.Printf("| %v%v ", v[i], cAppend(" ", total-utf8.RuneCountInString(v[i])))
					}
				}
			}
			fmt.Printf("|\n")
		}
	}

	displaySeparator(columnsWidth)

	fmt.Println()
}

func displayLineOfTable(label string, value string, labelLen int, valueLen int) {
	fmt.Printf("| %v%v ", label, cAppend(" ", labelLen-utf8.RuneCountInString(label)))
	fmt.Printf("| %v%v ", value, cAppend(" ", valueLen-utf8.RuneCountInString(value)))
	fmt.Printf("|\n")
}

func displaySeparator(columnsWidth []int) int {
	res := -3
	for i, _ := range columnsWidth {
		res += columnsWidth[i] + 3

		fmt.Printf("+%v", cAppend("-", columnsWidth[i]+2))
	}
	fmt.Printf("+\n")

	return res
}

func cAppend(s string, n int) string {
	var buffer bytes.Buffer

	for i := 0; i < n; i++ {
		buffer.WriteString(s)
	}

	return buffer.String()
}

func removePunctuation(text string) string {
	pct := `[!//#$%&()*+,-./:;<=>?@[]^_{|}~]`
	regexString := "\\s*(\\.{3}|\\s+\\-\\s+|\\s+\"(?:\\s+)?|\\s|[" + pct + "])"

	re := regexp.MustCompile(regexString)

	return strings.TrimSpace(re.ReplaceAllString(text, " "))
}
