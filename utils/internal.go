package utils

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"rest/models"
	"strconv"
	"strings"
)

const (
	path = "identifiers"
)

// CountBits returns number of active bits
func CountBits(n int64) int {
	count := 0
	for n > 0 {
		count++
		n = n & (n - 1)
	}

	return count
}

// IsLatin checks if string passed contains only alphabetic characters
func IsLatin(s string) bool {
	for _, char := range s {
		if char < 'A' || char > 'Z' && char < 'a' || char > 'z' {
			return false
		}
	}
	return true
}

// ValidateUser validates user info
func ValidateUser(u models.User) bool {
	firstName, lastName := u.FirstName, u.LastName
	if firstName == "" || lastName == "" || !IsLatin(firstName) || !IsLatin(lastName) {
		return false
	}
	return true
}

// ValidateID validates ID
func ValidateID(ID string) bool {
	if _, err := strconv.Atoi(ID); err != nil || ID == "" {
		return false
	}
	return true
}

// LongestSbstring returns substring with highest number of unique characters.
// If more than one are ound with same maximum length, first one is returned
func LongestSubstring(s string) string {
	maxLen := 0
	longestSubstr := ""
	start := 0
	m := make(map[byte]int)
	for i := 0; i < len(s); i++ {
		if ind, ok := m[s[i]]; ok {
			if start < ind {
				start = ind
			}
		}
		if substr := s[start : i+1]; len(substr) > maxLen {
			longestSubstr = substr
			maxLen = len(substr)
		}
		m[s[i]] = i + 1
	}
	return longestSubstr

}

// ValidateIIN validates IIN
// This may not work for those born in 2000s
func ValidateIIN(s string) bool {
	if len(s) != 12 {
		return false
	}
	if _, err := strconv.Atoi(s); err != nil || s[0] == '-' || s[0] == '+' {
		return false
	}
	year := 10*int(s[0]-'0') + int(s[1]-'0')
	var month int
	if month := 10*int(s[2]-'0') + int(s[3]-'0'); month > 12 {
		return false
	}
	day := 10*int(s[4]-'0') + int(s[5]-'0')
	if (month == 4 || month == 6 || month == 9 || month == 11) && day > 30 {
		return false
	} else if month == 2 {
		if !isLeapYear(year) {
			if day > 29 {
				return false
			}
		} else {
			if day > 28 {
				return false
			}
		}
	} else {
		if day > 31 {
			fmt.Println(">31")
			return false
		}
	}
	if s[6] > '6' {
		return false
	}
	var mod int
	if mod = (int(s[0]-'0') + 2*int(s[1]-'0') + 3*int(s[2]-'0') + 4*int(s[3]-'0') + 5*int(s[4]-'0') + 6*int(s[5]-'0') + 7*int(s[6]-'0') + 8*int(s[7]-'0') + 9*int(s[8]-'0') + 10*int(s[9]-'0') + 11*int(s[10]-'0')) % 11; mod == 10 {
		mod = (3*int(s[0]-'0') + 4*int(s[1]-'0') + 5*int(s[2]-'0') + 6*int(s[3]-'0') + 7*int(s[4]-'0') + 8*int(s[5]-'0') + 9*int(s[6]-'0') + 10*int(s[7]-'0') + 11*int(s[8]-'0') + int(s[9]-'0') + 2*int(s[10]-'0')) % 11
	}
	return mod == int(s[11]-'0')
}

// isLeapYear checks if year is leap
func isLeapYear(y int) bool {
	if y%4 == 0 && y%100 != 0 || y%400 == 0 {
		return true
	} else {
		return false
	}
}

// GetIdentifiers returns all identifiers with specified name
func GetIdentifiers(str, searchDir string) ([]byte, error) {
	fileList := []string{}
	if err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	}); err != nil {
		return nil, err
	}
	for _, file := range fileList {
		if err := walk(file, str); err != nil {
			log.Println(err)
		}
	}
	res, err := ReadFile()
	if err != nil {
		return nil, err
	}
	return res, nil
}

// walk walks through entire file tree
func walk(fname, str string) error {
	file, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer file.Close()

	// read the whole file in
	srcbuf, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	src := string(srcbuf)

	// file set
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "lib.go", src, 0)
	if err != nil {
		return err
	}

	// main inspection
	ast.Inspect(f, func(n ast.Node) bool {
		var res string
		switch fn := n.(type) {

		// catching all function declarations
		// other intersting things to catch FuncLit and FuncType
		case *ast.FuncDecl:
			fName := ""
			// fmt.Println("FNAME", fName)
			// if a method
			if fn.Recv != nil {
				fName = fmt.Sprintf("method: %s()", fn.Name.Name)
			} else {
				fName = fmt.Sprintf("func: %s()", fn.Name.Name)
			}
			// fmt.Println(fName)
			res += fName + "\n"
			// fmt.Println()

		case *ast.GenDecl:
			for _, spec := range fn.Specs {
				switch spec := spec.(type) {
				case *ast.ImportSpec:
					if name := spec.Path.Value; strings.Contains(name, str) {
						// fmt.Println("import:", spec.Path.Value)
						res += fmt.Sprintf("import: %s", spec.Path.Value) + "\n"
					}
				case *ast.TypeSpec:
					if name := spec.Name.String(); strings.Contains(name, str) {
						// fmt.Println("type:", spec.Name.String())
						res += fmt.Sprintf("type: %s", spec.Name.String()) + "\n"
					}
				case *ast.ValueSpec:
					for _, id := range spec.Names {
						if name := id.Name; strings.Contains(name, str) {
							// fmt.Printf("%v: %v\n", id.Obj.Kind, id.Name)
							res += fmt.Sprintf("%v: %v\n", id.Obj.Kind, id.Name) + "\n"
						}
						//fmt.Printf("Var %s: %v", id.Name, id.Obj.Decl.(*ast.ValueSpec).Values[0].(*ast.BasicLit).Value)
					}
				}
			}
		}
		fileN, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err) // hmmmm
		}
		defer fileN.Close()

		if _, err = fileN.WriteString(res); err != nil {
			panic(err) // hmmmm
		}

		return true
	})
	return nil
}

// expr parses ast.Expr
func expr(e ast.Expr) (ret string) {
	switch x := e.(type) {
	case *ast.StarExpr:
		return fmt.Sprintf("%s*%v", ret, x.X)
	case *ast.Ident:
		return fmt.Sprintf("%s%v", ret, x.Name)
	case *ast.ArrayType:
		if x.Len != nil {
			return "some array"
		}
		res := expr(x.Elt)
		return fmt.Sprintf("%s[]%v", ret, res)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", expr(x.Key), expr(x.Value))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", expr(x.X), expr(x.Sel))
	default:
		log.Printf("couldn't recognize type: %#v\n", x) // hmmmmm
	}
	return
}

// fields parses ast.FieldList
func fields(fl ast.FieldList) (ret string) {
	pcomma := ""
	for i, f := range fl.List {
		// get all the names if present
		var names string
		ncomma := ""
		for j, n := range f.Names {
			if j > 0 {
				ncomma = ", "
			}
			names = fmt.Sprintf("%s%s%s ", names, ncomma, n)
		}
		if i > 0 {
			pcomma = ", "
		}
		ret = fmt.Sprintf("%s%s%s%s", ret, pcomma, names, expr(f.Type))
	}
	return ret
}

// ReadFile parses output file and returns its contents
func ReadFile() ([]byte, error) {
	body, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return body, nil
}
