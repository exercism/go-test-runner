package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"strconv"
	"strings"
)

type subTData struct {
	subTKey    string     // subtest name key
	origTDName string     // original test data []struct name
	newTDName  string     // new test data struct name
	TD         []ast.Expr // original test data node
	subTest    ast.Stmt   // Run() function literal node
}

// return the code of the "test" function from a file
func getFuncCode(test string, fstr string) string {
	fset := token.NewFileSet()
	ppc := parser.ParseComments
	if file, err := parser.ParseFile(fset, fstr, nil, ppc); err == nil {
		for _, d := range file.Decls {
			if f, ok := d.(*ast.FuncDecl); ok && f.Name.Name == test {
				fun := &printer.CommentedNode{Node: f, Comments: file.Comments}
				var buf bytes.Buffer
				printer.Fprint(&buf, fset, fun)
				return buf.String()
			}
		}
	} else {
		log.Printf(
			"warning: '%s' not parsed from '%s': %s", test, fstr, err,
		)
	}
	return ""
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

	if 2 != len(fbAST) {
		log.Println("warning: subtests are constrained to two top level nodes")
		return ""
	}

	// Ensure the first two statements are an assignment and a range
	tdast, ok := fbAST[0].(*ast.AssignStmt) // f.Decls[0].Body.List[0]
	if !ok {
		log.Println("warning: first subtest statement must be an assignment")
		return ""
	}
	rast, ok := fbAST[1].(*ast.RangeStmt) // f.Decls[0].Body.List[1]
	if !ok {
		log.Println("warning: second subtest statement must be a range keyword")
		return ""
	}

	// process the test data assignment
	metadata, ok := processTestDataAssgn(sub, tdast)
	if !ok {
		return ""
	}
	lhs1 := tdast.Lhs[0].(*ast.Ident)        // f.Decls[0].Body.List[0].Lhs[0]
	rhs1 := tdast.Rhs[0].(*ast.CompositeLit) // f.Decls[0].Body.List[0].Rhs[0]

	// process the range statement
	ok = processRange(metadata, rast)
	if !ok {
		return ""
	}

	// rename the test data to match the variable assigned in the range stmt
	lhs1.Name = metadata.newTDName
	// assign the subtest data to the new test data variable
	rhs1.Elts = metadata.TD
	// create a new assignment statement to replace the original
	newassgn := &ast.AssignStmt{
		Lhs:    []ast.Expr{lhs1},
		TokPos: tdast.TokPos,
		Tok:    tdast.Tok,
		Rhs:    []ast.Expr{rhs1},
	}
	// swap the new assignment statement for the original
	tdast = newassgn

	// swap the original range statement for the extracted subtest
	fbAST[1] = metadata.subTest

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		log.Println("warning: failed to format extracted AST for subtest")
		return ""
	}
	return buf.String()
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
			altsub := strconv.Quote(strings.Replace(sub, " ", "_", -1))
			// still check the original subtest name, in case it had underscores
			if strconv.Quote(sub) == value.Value || altsub == value.Value {
				metadata.subTKey = kv.Key.(*ast.Ident).Name // subtest data "name"
				// TD is the "parent" array of KeyValueExprs
				metadata.TD = vals.Elts // test data element for the requested subtest
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

	body := runfunclit.(*ast.FuncLit).Body.List[0]
	metadata.subTest = body
	return true
}
