package testrunner

import (
	"bytes"
	"errors"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type subTData struct {
	subTKey    string            // subtest name key
	origTDName string            // original test data []struct name
	newTDName  string            // new test data struct name
	TD         *ast.CompositeLit // original test data node
	subTest    []ast.Stmt        // statements comprising the body of Run() function node
}

type subTestAstInfo struct {
	testDataAst    *ast.AssignStmt
	testDataAstIdx int
	rangeAst       *ast.RangeStmt
	rangeAstIdx    int
}

type rootLevelTest struct {
	name     string
	fileName string
	code     string
	taskID   uint64
}

// FindAllRootLevelTests parses the test file and extracts the name,
// test code and task id for each top level test (parent test) in the file.
func FindAllRootLevelTests(fileName string) []rootLevelTest {
	defer handleASTPanic()
	tests := []rootLevelTest{}
	fset := token.NewFileSet()
	ppc := parser.ParseComments
	file, err := parser.ParseFile(fset, fileName, nil, ppc)
	if err != nil {
		log.Printf("error: not able to parse '%s': %s", fileName, err)
		return nil
	}
	for _, d := range file.Decls {
		if f, ok := d.(*ast.FuncDecl); ok && strings.HasPrefix(f.Name.Name, "Test") {
			taskID := findTaskID(f.Doc)
			fun := &printer.CommentedNode{Node: f, Comments: file.Comments}
			var buf bytes.Buffer
			printErr := printer.Fprint(&buf, fset, fun)
			if printErr != nil {
				log.Printf("warning: failed to print AST for test %s in %s: %s",
					f.Name.Name, fileName, err,
				)
			}

			tests = append(tests, rootLevelTest{
				name:     f.Name.Name,
				fileName: fileName,
				code:     buf.String(),
				taskID:   taskID,
			})
		}
	}
	return tests
}

func ConvertToMapByTestName(tests []rootLevelTest) map[string]rootLevelTest {
	result := map[string]rootLevelTest{}
	for i := range tests {
		result[tests[i].name] = tests[i]
	}
	return result
}

var taskIDFormat = regexp.MustCompile(`testRunnerTaskID=([0-9]+)`)

// findTaskID checks whether there is a task ID set in a function comment,
// e.g. "testRunnerTaskID=2".
// If no task ID was identified, 0 is returned.
func findTaskID(doc *ast.CommentGroup) uint64 {
	matches := taskIDFormat.FindStringSubmatch(doc.Text())
	if len(matches) != 2 {
		return 0
	}

	taskID, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		log.Println("warning: failed to parse testRunnerTaskID value")
		return 0
	}

	return taskID
}

// generate simplified test code corresponding to a subtest
func getSubCode(test string, sub string, code string, file string) string {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(
		fset, file, "package main\n"+code, parser.ParseComments,
	)
	if err != nil {
		log.Printf("warning: '%s' not parsed from '%s': %s", test, file, err)
		return ""
	}

	fAST, ok := f.Decls[0].(*ast.FuncDecl)
	if !ok {
		log.Println("warning: first subtest declaration must be a function")
		return ""
	}

	fbAST := fAST.Body.List // f.Decls[0].Body.List

	astInfo, err := findTestDataAndRange(fbAST)
	if err != nil {
		log.Printf("warning: could not find test table and/or range: %v\n", err)
		return ""
	}

	// process the test data assignment
	metadata, ok := processTestDataAssgn(sub, astInfo.testDataAst)
	if !ok {
		return ""
	}
	lhs1 := astInfo.testDataAst.Lhs[0].(*ast.Ident)        // f.Decls[0].Body.List[0].Lhs[0]
	rhs1 := astInfo.testDataAst.Rhs[0].(*ast.CompositeLit) // f.Decls[0].Body.List[0].Rhs[0]

	// process the range statement
	ok = processRange(metadata, astInfo.rangeAst)
	if !ok {
		return ""
	}

	// rename the test data to match the variable assigned in the range stmt
	lhs1.Name = metadata.newTDName
	// assign the subtest data to the new test data variable
	*rhs1 = *metadata.TD

	// splice the statements of the extracted subtest in place of the original `for...range` statement
	fAST.Body.List = append(fbAST[:astInfo.rangeAstIdx], append(metadata.subTest, fbAST[astInfo.rangeAstIdx+1:]...)...)

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		log.Println("warning: failed to format extracted AST for subtest")
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(buf.String(), "package main"))
}

func findTestDataAndRange(stmtList []ast.Stmt) (subTestAstInfo, error) {
	result := subTestAstInfo{}

	for i := range stmtList {
		assignCandidate, ok := stmtList[i].(*ast.AssignStmt)
		if ok && result.testDataAst == nil {
			result.testDataAst = assignCandidate
			result.testDataAstIdx = i
		} else if ok {
			identifier, isIdentifier := assignCandidate.Lhs[0].(*ast.Ident)
			if !isIdentifier {
				continue
			}
			// Overwrite the assignment we already found in case there is an
			// assignment to a "tests" variable.
			if identifier.Name == "tests" {
				result.testDataAst = assignCandidate
				result.testDataAstIdx = i
			}
		}

		rangeCandidate, ok := stmtList[i].(*ast.RangeStmt)
		// If we found a range after we already found an assignment, we are good to go.
		if ok && result.testDataAst != nil {
			result.rangeAst = rangeCandidate
			result.rangeAstIdx = i
			return result, nil
		}
	}

	if result.testDataAst == nil {
		return subTestAstInfo{}, errors.New("failed to find assignment in sub-test")
	}

	return subTestAstInfo{}, errors.New("failed to find range statement in sub-test")
}

// validate the test data assignment and return the associated metadata
func processTestDataAssgn(sub string, assgn *ast.AssignStmt) (*subTData, bool) {
	lhs1, ok := assgn.Lhs[0].(*ast.Ident) // f.Decls[0].Body.List[0].Lhs[0]
	if !ok {
		log.Println("warning: test data assignment not found")
		return nil, false
	}
	if ast.Var != lhs1.Obj.Kind {
		log.Println("warning: test data assignment must be a var")
		return nil, false
	}
	metadata := subTData{origTDName: lhs1.Name}

	rhs1, ok := assgn.Rhs[0].(*ast.CompositeLit) // f.Decls[0].Body.List[0].Rhs[0]
	if !ok {
		log.Println("warning: test data assignment must be a composite literal")
		return nil, false
	}
	fieldNames := getAllFieldNames(rhs1.Type)
	// Loop for all of the test data structs
	for _, td := range rhs1.Elts {
		vals, ok := td.(*ast.CompositeLit)
		if !ok {
			continue
		}
		// Loop for each KeyValueExpr in the struct
		for i, tv := range vals.Elts {
			var key string
			var value *ast.BasicLit
			kv, ok := tv.(*ast.KeyValueExpr)
			if ok {
				key = kv.Key.(*ast.Ident).Name
				value, ok = kv.Value.(*ast.BasicLit)
			} else if fieldNames != nil {
				key = fieldNames[i]
				value, ok = tv.(*ast.BasicLit)
			}
			if !ok {
				continue
			}
			if token.STRING != value.Kind {
				continue
			}
			// spaces are replaced with underscores in subtest names
			// caveat: a subtest name mixing spaces and underscores cannot be found!
			altsub := strconv.Quote(strings.ReplaceAll(sub, "_", " "))
			// still check the original subtest name, in case it had underscores
			if strconv.Quote(sub) == value.Value || altsub == value.Value {
				metadata.subTKey = key // subtest data "name"
				// TD is the "parent" array of KeyValueExprs
				metadata.TD = vals // test data element for the requested subtest
				// re-assign the type from an array to the underlying test data struct
				metadata.TD.Type = rhs1.Type.(*ast.ArrayType).Elt
				return &metadata, true
			}
		}
	}
	log.Printf("warning: could not find test data struct for subtest: %s", sub)
	return nil, false
}

// getAllFieldNames returns all the field names of anonymous struct type
// not support for named struct type yet
func getAllFieldNames(exp ast.Expr) []string {
	structType, ok := exp.(*ast.ArrayType).Elt.(*ast.StructType)
	if !ok {
		return nil
	}
	keys := make([]string, 0)
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			keys = append(keys, name.Name)
		}
	}
	return keys
}

// validate the range over the test data and store associated metadata
func processRange(metadata *subTData, rastmt *ast.RangeStmt) bool {
	// Confirm that the range is over the test data
	if rastmt.X.(*ast.Ident).Name != metadata.origTDName {
		log.Printf("warning: test data (%s) and range value (%s) mismatch",
			rastmt.X.(*ast.Ident).Name, metadata.origTDName,
		)
		return false
	}

	// Pull the name of the subtest data being used
	metadata.newTDName = rastmt.Value.(*ast.Ident).Name

	// Parse the Run() call within the range statement
	rblexp := rastmt.Body.List[0].(*ast.ExprStmt).X

	// Parse the function literal from the Run() call within the range statement
	runcall := rblexp.(*ast.CallExpr).Fun

	if runcall.(*ast.SelectorExpr).Sel.Name != "Run" {
		log.Printf("warning: Run() call must follow range loop: (%s)",
			runcall.(*ast.SelectorExpr).Sel.Name,
		)
		return false
	}

	runselector := rblexp.(*ast.CallExpr).Args[0]
	runfunclit := rblexp.(*ast.CallExpr).Args[1]

	if metadata.newTDName != runselector.(*ast.SelectorExpr).X.(*ast.Ident).Name {
		log.Printf("warning: Run() call not passing expected test data %s: %s",
			metadata.newTDName, runselector.(*ast.SelectorExpr).X.(*ast.Ident).Name,
		)
		return false
	}

	if metadata.subTKey != runselector.(*ast.SelectorExpr).Sel.Name {
		log.Printf("warning: Run() call name (%s) must match test data struct: %s",
			runselector.(*ast.SelectorExpr).X.(*ast.Ident).Name, metadata.subTKey,
		)
		return false
	}

	body := runfunclit.(*ast.FuncLit).Body.List
	metadata.subTest = body
	return true
}
