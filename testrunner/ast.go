package testrunner

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"path/filepath"
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
	pkgName  string
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
			printer.Fprint(&buf, fset, fun)

			tests = append(tests, rootLevelTest{
				name:     f.Name.Name,
				fileName: fileName,
				code:     buf.String(),
				taskID:   taskID,
				pkgName:  file.Name.Name,
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
func getSubCode(test string, sub string, code string, file string, pkgName string) string {
	pkgLine := fmt.Sprintf("package %s\n", pkgName)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(
		fset, file, pkgLine+code, parser.ParseComments,
	)
	if err != nil {
		log.Printf("warning: '%s' not parsed from '%s': %s", test, file, err)
		return ""
	}

	resolveTestData(fset, f, file)

	fAST, ok := f.Decls[0].(*ast.FuncDecl)
	if !ok {
		log.Println("warning: first subtest declaration must be a function")
		return ""
	}

	fbAST := fAST.Body.List // f.Decls[0].Body.List

	astInfo, err := findTestDataAndRange(fbAST, fset)
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
	if astInfo.testDataAstIdx != -1 { // testDataAst is already in the test function
		return strings.TrimSpace(strings.TrimPrefix(buf.String(), pkgLine))
	}
	return insertTestDataASTIntoFunc(fset, astInfo.testDataAst, fAST.Body, buf.Bytes(), pkgLine)
}

func findTestDataAndRange(stmtList []ast.Stmt, fset *token.FileSet) (subTestAstInfo, error) {
	result := subTestAstInfo{}
	posToIndex := make(map[token.Position]int)
	for i := range stmtList {
		posToIndex[fset.Position(stmtList[i].Pos())] = i
		if rangeCandidate, ok := stmtList[i].(*ast.RangeStmt); ok {
			assignCandidate := getTestDataAssignFromRange(rangeCandidate)
			if assignCandidate != nil {
				// check if assignCandidate is in the same function with rangeCandidate
				if idx, ok := posToIndex[fset.Position(assignCandidate.Pos())]; ok &&
					fset.File(assignCandidate.Pos()).Name() == fset.File(rangeCandidate.Pos()).Name() {
					result.testDataAstIdx = idx
				} else {
					result.testDataAstIdx = -1
				}
				result.testDataAst = assignCandidate
				result.rangeAst = rangeCandidate
				result.rangeAstIdx = i
				return result, nil
			}
			return subTestAstInfo{}, errors.New("failed to find assignment in sub-test")
		}
	}

	if result.testDataAst == nil {
		return subTestAstInfo{}, errors.New("failed to find assignment in sub-test")
	}

	return subTestAstInfo{}, errors.New("failed to find range statement in sub-test")
}
func getTestDataAssignFromRange(rangeAst *ast.RangeStmt) *ast.AssignStmt {
	spec := rangeAst.X.(*ast.Ident).Obj.Decl
	if assignStmt, ok := spec.(*ast.AssignStmt); ok {
		return assignStmt
	}
	if valueSpec, ok := spec.(*ast.ValueSpec); ok {
		lhs := make([]ast.Expr, len(valueSpec.Names))
		for i, name := range valueSpec.Names {
			lhs[i] = name
		}
		return &ast.AssignStmt{
			Lhs: lhs,
			Tok: token.DEFINE,
			Rhs: valueSpec.Values,
		}
	}
	return nil
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

	// Loop for all of the test data structs
	for _, td := range rhs1.Elts {
		vals, ok := td.(*ast.CompositeLit)
		if !ok {
			continue
		}

		// Loop for each KeyValueExpr in the struct
		for _, tv := range vals.Elts {
			kv, ok := tv.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			value, ok := kv.Value.(*ast.BasicLit)
			if !ok {
				continue
			}
			if token.STRING != value.Kind {
				continue
			}
			// spaces are replaced with underscores in subtest names
			// caveat: a subtest name mixing spaces and underscores cannot be found!
			altsub := strconv.Quote(strings.Replace(sub, "_", " ", -1))
			// still check the original subtest name, in case it had underscores
			if strconv.Quote(sub) == value.Value || altsub == value.Value {
				metadata.subTKey = kv.Key.(*ast.Ident).Name // subtest data "name"
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

	if "Run" != runcall.(*ast.SelectorExpr).Sel.Name {
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

// resolveTestData resolves test data variable declared in cases_test.go (if exists)
func resolveTestData(fset *token.FileSet, f *ast.File, file string) {
	filedata := filepath.Join(filepath.Dir(file), "cases_test.go")
	fdata, _ := parser.ParseFile(fset, filedata, nil, parser.ParseComments)
	if fdata != nil {
		ast.NewPackage(fset, map[string]*ast.File{file: f, filedata: fdata}, nil, nil)
	} else {
		ast.NewPackage(fset, map[string]*ast.File{file: f}, nil, nil)
	}
}

// insertTestDataASTIntoFunc inserts testDataAst into the first line of fbAST function's body
func insertTestDataASTIntoFunc(fset *token.FileSet, testDataAst *ast.AssignStmt, fbAST *ast.BlockStmt, fileText []byte, pkgLine string) string {
	buf := bytes.Buffer{}

	p := fset.Position(fbAST.Lbrace).Offset + 1

	// write the beginning of fileText to func (...) {
	buf.Write(fileText[:p+1])
	
	// write test data assign stmt
	if err := format.Node(&buf, fset, testDataAst); err != nil {
		log.Println("warning: failed to format extracted AST for subtest")
		return ""
	}
	// write the rest of fileText
	buf.Write(fileText[p+1:])
	
	// because assign stmt is extracted from different file, its indentation is different from fileText 
	// so need to reformat
	src, err := format.Source((buf.Bytes()))
	if err != nil {
		log.Println("warning: failed to format extracted AST for subtest")
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(string(src), pkgLine))
}
