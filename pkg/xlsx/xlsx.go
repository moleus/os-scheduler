package xlsx

import (
	"errors"
	"fmt"
	m "github.com/Moleus/os-solver/pkg/machine"
	"github.com/xuri/excelize/v2"
	"os"
	"strconv"
)

const countOfHardcodedColors = 10

func printRow(f *excelize.File, sheet string, offset int, row int, values []string) {
	for pos, val := range values {
		err := f.SetCellValue(sheet, fmt.Sprintf("%s%d", fmt.Sprintf("%c", 'A'+offset+pos), row), val)
		if err != nil {
			return
		}
	}
}
func GetF(fileName string, sheet string) *excelize.File {
	var f *excelize.File
	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
		f = excelize.NewFile()

	} else {
		f, err = excelize.OpenFile(fileName)
		if err != nil {
			panic(err)
		}
		err = f.Close()
		if err != nil {
			panic(err)
		}
	}
	index, err := f.NewSheet(sheet)
	if err != nil {
		panic(err)
	}
	f.SetActiveSheet(index)
	err = f.DeleteSheet("Sheet1")
	if err != nil {
		return nil
	}
	return f
}
func setStyle(f *excelize.File, spreed string, cell string, val string, colors [countOfHardcodedColors]int) {
	if val == "-" {
		return
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		panic(err)
	}
	if i > countOfHardcodedColors {
		return
	}
	err = f.SetCellStyle(spreed, cell, cell, colors[i])
	if err != nil {
		panic(err)
	}
}
func GenerateStyles(f *excelize.File) [countOfHardcodedColors]int {
	colors := [countOfHardcodedColors]string{"E0EBF5", "#93e476", "#efb2b9", "#6a74eb", "#f0b1e4", "#c1b1f0", "#ead669", "#ebaa6a", "#eb836a", "#6aeb71"}
	var styles [countOfHardcodedColors]int
	for i := 0; i < countOfHardcodedColors; i++ {
		style, err := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Color: []string{colors[i]}, Pattern: 1},
		})
		if err != nil {
			fmt.Println(err)
		}
		styles[i] = style
	}
	return styles
}
func SnapshotStateXlsx(f *excelize.File, sheet string, tick string, cpusStateString []string, io1State string, io2State string, colors [countOfHardcodedColors]int, cpuCount int) {
	err := f.SetCellValue(sheet, fmt.Sprintf("A%s", tick), tick)
	if err != nil {
		return
	}
	for pos, val := range cpusStateString {
		err := f.SetCellValue(sheet, fmt.Sprintf("%s%s", fmt.Sprintf("%c", 'A'+pos+1), tick), val)
		if err != nil {
			return
		}
		setStyle(f, sheet, fmt.Sprintf("%s%s", fmt.Sprintf("%c", 'A'+pos+1), tick), val, colors)
	}
	err = f.SetCellValue(sheet, fmt.Sprintf("%s%s", fmt.Sprintf("%c", 'A'+cpuCount+1), tick), io1State)
	if err != nil {
		return
	}
	setStyle(f, sheet, fmt.Sprintf("%s%s", fmt.Sprintf("%c", 'A'+cpuCount+1), tick), io1State, colors)
	err = f.SetCellValue(sheet, fmt.Sprintf("%s%s", fmt.Sprintf("%c", 'A'+cpuCount+2), tick), io2State)
	if err != nil {
		return
	}
	setStyle(f, sheet, fmt.Sprintf("%s%s", fmt.Sprintf("%c", 'A'+cpuCount+2), tick), io2State, colors)
}
func PrintProcsStats(f *excelize.File, sheet string, procs []*m.Process, offset int) {
	headers := []string{"Process", "Arrival", "Service", "Waiting", "Finish_time", "Turnaround_(Tr)", "Tr/Ts"}
	printRow(f, sheet, offset, 1, headers)

	for pos, proc := range procs {
		stats := proc.GetStats()
		normalizedTurnaround := float64(stats.TurnaroundTime) / float64(stats.ServiceTime)

		values := []string{
			fmt.Sprintf("%v", stats.ProcId),
			fmt.Sprintf("%v", stats.EntranceTime),
			fmt.Sprintf("%v", stats.ServiceTime),
			fmt.Sprintf("%v", stats.ReadyOrBlockedTime),
			fmt.Sprintf("%v", stats.ExitTime),
			fmt.Sprintf("%v", stats.TurnaroundTime),
			fmt.Sprintf("%v", normalizedTurnaround),
		}

		printRow(f, sheet, offset, pos+2, values)
	}
}
func SaveReport(f *excelize.File, fileName string) {
	if err := f.SaveAs(fileName); err != nil {
		fmt.Println(err)
	}
}
