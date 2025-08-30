package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SocialCalc format utilities for Go backend

// ParseTouchCalcMSC parses the simple MSC format from touchcalc.msc.txt
func ParseTouchCalcMSC(mscData string) (string, error) {
	lines := strings.Split(strings.TrimSpace(mscData), "\n")
	var socialCalcLines []string
	
	// Add SocialCalc version header
	socialCalcLines = append(socialCalcLines, "version:1.4")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		
		coord := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Validate coordinate format (A1, B2, etc.)
		if !isValidCoordinate(coord) {
			continue
		}
		
		// Create SocialCalc cell entry
		if strings.HasPrefix(value, "=") {
			// Formula
			formula := strings.TrimPrefix(value, "=")
			socialCalcLines = append(socialCalcLines, fmt.Sprintf("cell:%s:vtf:n:0:%s", coord, formula))
		} else {
			// Text value
			escapedValue := escapeForSocialCalc(value)
			socialCalcLines = append(socialCalcLines, fmt.Sprintf("cell:%s:t:%s", coord, escapedValue))
		}
	}
	
	// Add sheet dimensions
	socialCalcLines = append(socialCalcLines, "sheet:c:10:r:20")
	
	return strings.Join(socialCalcLines, "\n"), nil
}

// ConvertSocialCalcToMSC converts SocialCalc format to simple MSC format
func ConvertSocialCalcToMSC(socialCalcData string) string {
	lines := strings.Split(socialCalcData, "\n")
	var mscLines []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "cell:") {
			continue
		}
		
		parts := strings.Split(line, ":")
		if len(parts) < 4 {
			continue
		}
		
		coord := parts[1]
		cellType := parts[2]
		
		var value string
		switch cellType {
		case "t": // text
			if len(parts) >= 4 {
				value = unescapeFromSocialCalc(parts[3])
			}
		case "v": // numeric value
			if len(parts) >= 4 {
				value = parts[3]
			}
		case "vtf": // formula
			if len(parts) >= 6 {
				value = "=" + parts[5]
			}
		default:
			continue
		}
		
		if value != "" {
			mscLines = append(mscLines, fmt.Sprintf("%s:%s", coord, value))
		}
	}
	
	return strings.Join(mscLines, "\n")
}

// CreateDefaultTouchCalcData creates default spreadsheet data
func CreateDefaultTouchCalcData(storageBackend string) string {
	defaultData := []string{
		"version:1.4",
		"cell:A1:t:TouchCalc Spreadsheet",
		"cell:B1:t:Welcome to your cloud spreadsheet!",
		"cell:A2:t:Cell A2",
		"cell:B2:t:Cell B2", 
		"cell:C2:vtf:n:0:A2+B2",
		"cell:A3:t:Data automatically saves to the cloud",
		"cell:B3:t:Storage Backend:",
		fmt.Sprintf("cell:C3:t:%s", storageBackend),
		"cell:A5:t:Try these features:",
		"cell:B5:t:• Edit any cell by clicking",
		"cell:B6:t:• Use formulas like =A1+B1",
		"cell:B7:t:• Auto-save every 30 seconds",
		"cell:B8:t:• Export to CSV or Excel",
		"sheet:c:10:r:20",
	}
	
	return strings.Join(defaultData, "\n")
}

// Helper functions

func isValidCoordinate(coord string) bool {
	// Match patterns like A1, B2, AA1, etc.
	match, _ := regexp.MatchString(`^[A-Z]+[0-9]+$`, coord)
	return match
}

func escapeForSocialCalc(value string) string {
	// Escape special characters for SocialCalc format
	value = strings.ReplaceAll(value, "\\", "\\b")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, ":", "\\c")
	return value
}

func unescapeFromSocialCalc(value string) string {
	// Unescape special characters from SocialCalc format
	value = strings.ReplaceAll(value, "\\c", ":")
	value = strings.ReplaceAll(value, "\\n", "\n")
	value = strings.ReplaceAll(value, "\\b", "\\")
	return value
}

// CoordinateToRowCol converts Excel-style coordinate to row/column numbers
func CoordinateToRowCol(coord string) (int, int, error) {
	if !isValidCoordinate(coord) {
		return 0, 0, fmt.Errorf("invalid coordinate: %s", coord)
	}
	
	// Extract column letters and row number
	var colPart, rowPart string
	for i, char := range coord {
		if char >= '0' && char <= '9' {
			colPart = coord[:i]
			rowPart = coord[i:]
			break
		}
	}
	
	// Convert column letters to number (A=1, B=2, ..., AA=27, etc.)
	col := 0
	for _, char := range colPart {
		col = col*26 + int(char-'A') + 1
	}
	
	// Convert row string to number
	row, err := strconv.Atoi(rowPart)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid row number in coordinate: %s", coord)
	}
	
	return row, col, nil
}

// RowColToCoordinate converts row/column numbers to Excel-style coordinate
func RowColToCoordinate(row, col int) string {
	if row < 1 || col < 1 {
		return ""
	}
	
	// Convert column number to letters
	colStr := ""
	for col > 0 {
		col--
		colStr = string(rune('A'+col%26)) + colStr
		col /= 26
	}
	
	return fmt.Sprintf("%s%d", colStr, row)
}

// ValidateSocialCalcData validates SocialCalc format data
func ValidateSocialCalcData(data string) error {
	lines := strings.Split(data, "\n")
	hasVersion := false
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		if strings.HasPrefix(line, "version:") {
			hasVersion = true
			continue
		}
		
		if strings.HasPrefix(line, "cell:") {
			parts := strings.Split(line, ":")
			if len(parts) < 4 {
				return fmt.Errorf("invalid cell definition at line %d: %s", i+1, line)
			}
			
			coord := parts[1]
			if !isValidCoordinate(coord) {
				return fmt.Errorf("invalid coordinate at line %d: %s", i+1, coord)
			}
		}
	}
	
	if !hasVersion {
		return fmt.Errorf("missing version header")
	}
	
	return nil
}