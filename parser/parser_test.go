// Copyright 2015 The Neugram Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser_test

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"neugram.io/ng/format"
	"neugram.io/ng/parser"
	"neugram.io/ng/syntax/expr"
	"neugram.io/ng/syntax/stmt"
	"neugram.io/ng/syntax/tipe"
	"neugram.io/ng/syntax/token"
)

type parserTest struct {
	input string
	want  expr.Expr
}

var parserTests = []parserTest{
	{"foo", &expr.Ident{Name: "foo"}},
	{"x + y", &expr.Binary{Op: token.Add, Left: &expr.Ident{Name: "x"}, Right: &expr.Ident{Name: "y"}}},
	{
		"x + y + 9",
		&expr.Binary{
			Op:    token.Add,
			Left:  &expr.Binary{Op: token.Add, Left: &expr.Ident{Name: "x"}, Right: &expr.Ident{Name: "y"}},
			Right: &expr.BasicLiteral{Value: big.NewInt(9)},
		},
	},
	{
		"x + (y + 7)",
		&expr.Binary{
			Op:   token.Add,
			Left: &expr.Ident{Name: "x"},
			Right: &expr.Unary{
				Op: token.LeftParen,
				Expr: &expr.Binary{
					Op:    token.Add,
					Left:  &expr.Ident{Name: "y"},
					Right: &expr.BasicLiteral{Value: big.NewInt(7)},
				},
			},
		},
	},
	{
		"x + y * z",
		&expr.Binary{
			Op:    token.Add,
			Left:  &expr.Ident{Name: "x"},
			Right: &expr.Binary{Op: token.Mul, Left: &expr.Ident{Name: "y"}, Right: &expr.Ident{Name: "z"}},
		},
	},
	{
		"x | y",
		&expr.Binary{
			Op:    token.Pipe,
			Left:  &expr.Ident{Name: "x"},
			Right: &expr.Ident{Name: "y"},
		},
	},
	{
		"x ^ y",
		&expr.Binary{
			Op:    token.Pow,
			Left:  &expr.Ident{Name: "x"},
			Right: &expr.Ident{Name: "y"},
		},
	},
	{
		"x & y",
		&expr.Binary{
			Op:    token.Ref,
			Left:  &expr.Ident{Name: "x"},
			Right: &expr.Ident{Name: "y"},
		},
	},
	{
		"x % y",
		&expr.Binary{
			Op:    token.Rem,
			Left:  &expr.Ident{Name: "x"},
			Right: &expr.Ident{Name: "y"},
		},
	},
	{
		"x &^ y",
		&expr.Binary{
			Op:    token.RefPow,
			Left:  &expr.Ident{Name: "x"},
			Right: &expr.Ident{Name: "y"},
		},
	},
	{
		"x << y",
		&expr.Binary{
			Op:    token.TwoLess,
			Left:  &expr.Ident{Name: "x"},
			Right: &expr.Ident{Name: "y"},
		},
	},
	{
		"x >> y",
		&expr.Binary{
			Op:    token.TwoGreater,
			Left:  &expr.Ident{Name: "x"},
			Right: &expr.Ident{Name: "y"},
		},
	},
	{
		"quit()",
		&expr.Call{Func: &expr.Ident{Name: "quit"}},
	},
	{
		"foo(4)",
		&expr.Call{
			Func: &expr.Ident{Name: "foo"},
			Args: []expr.Expr{&expr.BasicLiteral{Value: big.NewInt(4)}},
		},
	},
	{
		"min(1, 2)",
		&expr.Call{
			Func: &expr.Ident{Name: "min"},
			Args: []expr.Expr{
				&expr.BasicLiteral{Value: big.NewInt(1)},
				&expr.BasicLiteral{Value: big.NewInt(2)},
			},
		},
	},
	{
		"func() integer { return 7 }",
		&expr.FuncLiteral{
			Type: &tipe.Func{
				Params:  &tipe.Tuple{},
				Results: &tipe.Tuple{Elems: []tipe.Type{tinteger}},
			},
			Body: &stmt.Block{Stmts: []stmt.Stmt{
				&stmt.Return{Exprs: []expr.Expr{&expr.BasicLiteral{Value: big.NewInt(7)}}},
			}},
		},
	},
	{
		"func(x, y val) (r0 val, r1 val) { return x, y }",
		&expr.FuncLiteral{
			Type: &tipe.Func{
				Params: &tipe.Tuple{Elems: []tipe.Type{
					&tipe.Unresolved{Name: "val"},
					&tipe.Unresolved{Name: "val"},
				}},
				Results: &tipe.Tuple{Elems: []tipe.Type{
					&tipe.Unresolved{Name: "val"},
					&tipe.Unresolved{Name: "val"},
				}},
			},
			ParamNames:  []string{"x", "y"},
			ResultNames: []string{"r0", "r1"},
			Body: &stmt.Block{Stmts: []stmt.Stmt{
				&stmt.Return{Exprs: []expr.Expr{
					&expr.Ident{Name: "x"},
					&expr.Ident{Name: "y"},
				}},
			}},
		},
	},
	{
		`func() int64 {
			x := 7
			return x
		}`,
		&expr.FuncLiteral{
			Type: &tipe.Func{
				Params:  &tipe.Tuple{},
				Results: &tipe.Tuple{Elems: []tipe.Type{tint64}},
			},
			ResultNames: []string{""},
			Body: &stmt.Block{Stmts: []stmt.Stmt{
				&stmt.Assign{
					Decl:  true,
					Left:  []expr.Expr{&expr.Ident{Name: "x"}},
					Right: []expr.Expr{&expr.BasicLiteral{Value: big.NewInt(7)}},
				},
				&stmt.Return{Exprs: []expr.Expr{&expr.Ident{Name: "x"}}},
			}},
		},
	},
	{
		`func() int64 {
			if x := 9; x > 3 {
				return x
			} else {
				return 1-x
			}
		}`,
		&expr.FuncLiteral{
			Type: &tipe.Func{
				Params:  &tipe.Tuple{},
				Results: &tipe.Tuple{Elems: []tipe.Type{tint64}},
			},
			ResultNames: []string{""},
			Body: &stmt.Block{Stmts: []stmt.Stmt{&stmt.If{
				Init: &stmt.Assign{
					Decl:  true,
					Left:  []expr.Expr{&expr.Ident{Name: "x"}},
					Right: []expr.Expr{&expr.BasicLiteral{Value: big.NewInt(9)}},
				},
				Cond: &expr.Binary{
					Op:    token.Greater,
					Left:  &expr.Ident{Name: "x"},
					Right: &expr.BasicLiteral{Value: big.NewInt(3)},
				},
				Body: &stmt.Block{Stmts: []stmt.Stmt{
					&stmt.Return{Exprs: []expr.Expr{&expr.Ident{Name: "x"}}},
				}},
				Else: &stmt.Block{Stmts: []stmt.Stmt{
					&stmt.Return{Exprs: []expr.Expr{
						&expr.Binary{
							Op:    token.Sub,
							Left:  &expr.BasicLiteral{Value: big.NewInt(1)},
							Right: &expr.Ident{Name: "x"},
						},
					}},
				}},
			}}},
		},
	},
	{
		"func(x val) val { return 3+x }(1)",
		&expr.Call{
			Func: &expr.FuncLiteral{
				Type: &tipe.Func{
					Params:  &tipe.Tuple{Elems: []tipe.Type{&tipe.Unresolved{Name: "val"}}},
					Results: &tipe.Tuple{Elems: []tipe.Type{&tipe.Unresolved{Name: "val"}}},
				},
				ParamNames:  []string{""},
				ResultNames: []string{""},
				Body: &stmt.Block{Stmts: []stmt.Stmt{
					&stmt.Return{Exprs: []expr.Expr{
						&expr.Binary{
							Op:    token.Add,
							Left:  &expr.BasicLiteral{Value: big.NewInt(3)},
							Right: &expr.Ident{Name: "x"},
						},
					}},
				}},
			},
			Args: []expr.Expr{&expr.BasicLiteral{Value: big.NewInt(1)}},
		},
	},
	{
		"func() { x = -x }",
		&expr.FuncLiteral{
			Type: &tipe.Func{
				Params: &tipe.Tuple{},
			},
			Body: &stmt.Block{Stmts: []stmt.Stmt{&stmt.Assign{
				Left:  []expr.Expr{&expr.Ident{Name: "x"}},
				Right: []expr.Expr{&expr.Unary{Op: token.Sub, Expr: &expr.Ident{Name: "x"}}},
			}}},
		},
	},
	{"x.y.z", &expr.Selector{Left: &expr.Selector{Left: &expr.Ident{Name: "x"}, Right: &expr.Ident{Name: "y"}}, Right: &expr.Ident{Name: "z"}}},
	{"y * /* comment */ z", &expr.Binary{Op: token.Mul, Left: &expr.Ident{Name: "y"}, Right: &expr.Ident{Name: "z"}}},
	{"y * z//comment", &expr.Binary{Op: token.Mul, Left: &expr.Ident{Name: "y"}, Right: &expr.Ident{Name: "z"}}},
	{`"hello"`, &expr.BasicLiteral{Value: "hello"}},
	{`"hello \"neugram\""`, &expr.BasicLiteral{Value: `hello "neugram"`}},
	//TODO{`"\""`, &expr.BasicLiteral{Value:`"\""`}}
	{"x[4]", &expr.Index{Left: &expr.Ident{Name: "x"}, Indicies: []expr.Expr{basic(4)}}},
	{"x[1+2]", &expr.Index{
		Left: &expr.Ident{Name: "x"},
		Indicies: []expr.Expr{&expr.Binary{Op: token.Add,
			Left:  basic(1),
			Right: basic(2),
		}},
	}},
	{"x[1:3]", &expr.Index{
		Left:     &expr.Ident{Name: "x"},
		Indicies: []expr.Expr{&expr.Slice{Low: basic(1), High: basic(3)}},
	}},
	{"x[1:]", &expr.Index{Left: &expr.Ident{Name: "x"}, Indicies: []expr.Expr{&expr.Slice{Low: basic(1)}}}},
	{"x[:3]", &expr.Index{Left: &expr.Ident{Name: "x"}, Indicies: []expr.Expr{&expr.Slice{High: basic(3)}}}},
	{"x[:]", &expr.Index{Left: &expr.Ident{Name: "x"}, Indicies: []expr.Expr{&expr.Slice{}}}},
	{"x[:,:]", &expr.Index{Left: &expr.Ident{Name: "x"}, Indicies: []expr.Expr{&expr.Slice{}, &expr.Slice{}}}},
	{"x[1:,:3]", &expr.Index{Left: &expr.Ident{Name: "x"}, Indicies: []expr.Expr{&expr.Slice{Low: basic(1)}, &expr.Slice{High: basic(3)}}}},
	{"x[1:3,5:7]", &expr.Index{Left: &expr.Ident{Name: "x"}, Indicies: []expr.Expr{&expr.Slice{Low: basic(1), High: basic(3)}, &expr.Slice{Low: basic(5), High: basic(7)}}}},
	/* TODO
	{`x["C1"|"C2"]`, &expr.TableIndex{Expr: &expr.Ident{Name: "x"}, ColNames: []string{"C1", "C2"}}},
	{`x["C1",1:]`, &expr.TableIndex{
		Expr:     &expr.Ident{Name: "x"},
		ColNames: []string{"C1"},
		Rows:     expr.Range{Start: &expr.BasicLiteral{Value:big.NewInt(1)}},
	}},
	/*{"[|]num{}", &expr.TableLiteral{Type: &tipe.Table{tipe.Num}}},
	{"[|]num{{0, 1, 2}}", &expr.TableLiteral{
		Type: &tipe.Table{tipe.Num},
		Rows: [][]expr.Expr{{basic(0), basic(1), basic(2)}},
	}},
	{`[|]num{{|"Col1"|}, {1}, {2}}`, &expr.TableLiteral{
		Type:     &tipe.Table{tipe.Num},
		ColNames: []expr.Expr{basic("Col1")},
		Rows:     [][]expr.Expr{{basic(1)}, {basic(2)}},
	}},
	*/
	{"($$ls$$)", &expr.Unary{ // for Issue #50
		Op: token.LeftParen,
		Expr: &expr.Shell{
			Cmds: []*expr.ShellList{{AndOr: []*expr.ShellAndOr{{Pipeline: []*expr.ShellPipeline{{
				Cmd: []*expr.ShellCmd{{
					SimpleCmd: &expr.ShellSimpleCmd{
						Args: []string{"ls"},
					},
				},
				},
			}}}}}},
			TrapOut: true,
		}},
	},
}

var tint64 = &tipe.Unresolved{Name: "int64"}
var tinteger = &tipe.Unresolved{Name: "integer"}
var tfloat64 = &tipe.Unresolved{Name: "float64"}

func TestParseExpr(t *testing.T) {
	for _, test := range parserTests {
		t.Logf("Parsing %q\n", test.input)
		s, err := parser.ParseStmt([]byte(test.input))
		if err != nil {
			t.Errorf("ParseExpr(%q): error: %v", test.input, err)
			continue
		}
		if s == nil {
			t.Errorf("ParseExpr(%q): nil stmt", test.input)
			continue
		}
		got := s.(*stmt.Simple).Expr
		if !parser.EqualExpr(got, test.want) {
			diff := format.Diff(test.want, got)
			if diff == "" {
				t.Errorf("ParseExpr(%q): format.Diff empty but expressions not equal", test.input)
			} else {
				t.Errorf("ParseExpr(%q):\n%v", test.input, diff)
			}
		}
	}
}

type parserErrTest struct {
	input     string
	errsubstr string
}

var parserErrTests = []parserErrTest{
	{`\`, `unknown token: '\'`},
}

func TestParseError(t *testing.T) {
	for _, test := range parserErrTests {
		t.Logf("Parsing %q\n", test.input)
		_, err := parser.ParseStmt([]byte(test.input))
		if err == nil {
			t.Errorf("ParseStmt(%q): missing expected error", test.input)
			continue
		}
		if got := err.Error(); !strings.Contains(got, test.errsubstr) {
			t.Errorf("ParseStmt(%q): error %q does not contain %q", test.input, got, test.errsubstr)
		}
	}
}

var shellTests = []parserTest{
	{``, &expr.Shell{}},
	{`ls -l`, simplesh("ls", "-l")},
	{`ls | head`, &expr.Shell{
		Cmds: []*expr.ShellList{{
			AndOr: []*expr.ShellAndOr{{
				Pipeline: []*expr.ShellPipeline{{
					Bang: false,
					Cmd: []*expr.ShellCmd{
						{SimpleCmd: &expr.ShellSimpleCmd{Args: []string{"ls"}}},
						{SimpleCmd: &expr.ShellSimpleCmd{Args: []string{"head"}}},
					},
				}},
			}},
		}},
	}},
	{`ls > flist`, &expr.Shell{
		Cmds: []*expr.ShellList{{
			AndOr: []*expr.ShellAndOr{{
				Pipeline: []*expr.ShellPipeline{{
					Bang: false,
					Cmd: []*expr.ShellCmd{{
						SimpleCmd: &expr.ShellSimpleCmd{
							Redirect: []*expr.ShellRedirect{{Token: token.Greater, Filename: "flist"}},
							Args:     []string{"ls"},
						},
					}},
				}},
			}},
		}},
	}},
	{`echo hi | cat && true || false`, &expr.Shell{
		Cmds: []*expr.ShellList{{
			AndOr: []*expr.ShellAndOr{{
				Pipeline: []*expr.ShellPipeline{
					{
						Cmd: []*expr.ShellCmd{
							{
								SimpleCmd: &expr.ShellSimpleCmd{
									Args: []string{"echo", "hi"},
								},
							},
							{
								SimpleCmd: &expr.ShellSimpleCmd{
									Args: []string{"cat"},
								},
							},
						},
					},
					{
						Cmd: []*expr.ShellCmd{{
							SimpleCmd: &expr.ShellSimpleCmd{
								Args: []string{"true"},
							},
						}},
					},
					{
						Cmd: []*expr.ShellCmd{{
							SimpleCmd: &expr.ShellSimpleCmd{
								Args: []string{"false"},
							},
						}},
					},
				},
				Sep: []token.Token{token.LogicalAnd, token.LogicalOr},
			}},
		}},
	}},
	{`echo one && echo two > f || echo 3
	echo -n 4;
	echo 5 | wc; echo 6 & echo 7; echo 8 &`, &expr.Shell{
		Cmds: []*expr.ShellList{
			{
				AndOr: []*expr.ShellAndOr{{
					Pipeline: []*expr.ShellPipeline{
						{
							Cmd: []*expr.ShellCmd{{
								SimpleCmd: &expr.ShellSimpleCmd{
									Args: []string{"echo", "one"},
								},
							}},
						},
						{
							Cmd: []*expr.ShellCmd{{
								SimpleCmd: &expr.ShellSimpleCmd{
									Redirect: []*expr.ShellRedirect{{
										Token:    token.Greater,
										Filename: "f",
									}},
									Args: []string{"echo", "two"},
								},
							}},
						},
						{
							Cmd: []*expr.ShellCmd{{
								SimpleCmd: &expr.ShellSimpleCmd{
									Args: []string{"echo", "3"},
								},
							}},
						},
					},
					Sep: []token.Token{token.LogicalAnd, token.LogicalOr},
				}},
			},
			{
				AndOr: []*expr.ShellAndOr{
					{
						Pipeline: []*expr.ShellPipeline{{
							Cmd: []*expr.ShellCmd{{
								SimpleCmd: &expr.ShellSimpleCmd{
									Args: []string{"echo", "-n", "4"},
								},
							}},
						}},
					},
					{
						Pipeline: []*expr.ShellPipeline{{
							Cmd: []*expr.ShellCmd{
								{
									SimpleCmd: &expr.ShellSimpleCmd{
										Args: []string{"echo", "5"},
									},
								},
								{
									SimpleCmd: &expr.ShellSimpleCmd{
										Args: []string{"wc"},
									},
								},
							},
						}},
					},
					{
						Pipeline: []*expr.ShellPipeline{{
							Cmd: []*expr.ShellCmd{{
								SimpleCmd: &expr.ShellSimpleCmd{
									Args: []string{"echo", "6"},
								},
							}},
						}},
						Background: true,
					},
					{
						Pipeline: []*expr.ShellPipeline{{
							Cmd: []*expr.ShellCmd{{
								SimpleCmd: &expr.ShellSimpleCmd{
									Args: []string{"echo", "7"},
								},
							}},
						}},
					},
					{
						Pipeline: []*expr.ShellPipeline{{
							Cmd: []*expr.ShellCmd{{
								SimpleCmd: &expr.ShellSimpleCmd{
									Args: []string{"echo", "8"},
								},
							}},
						}},
						Background: true,
					},
				},
			},
		},
	}},
	{`echo start; (echo a; echo b 2>&1); echo end`, &expr.Shell{Cmds: []*expr.ShellList{{
		AndOr: []*expr.ShellAndOr{
			{Pipeline: []*expr.ShellPipeline{{
				Bang: false,
				Cmd: []*expr.ShellCmd{{
					SimpleCmd: &expr.ShellSimpleCmd{
						Args: []string{"echo", "start"},
					},
				}},
			}}},
			{Pipeline: []*expr.ShellPipeline{{
				Cmd: []*expr.ShellCmd{{
					SimpleCmd: (*expr.ShellSimpleCmd)(nil),
					Subshell: &expr.ShellList{
						AndOr: []*expr.ShellAndOr{
							{
								Pipeline: []*expr.ShellPipeline{{
									Cmd: []*expr.ShellCmd{{
										SimpleCmd: &expr.ShellSimpleCmd{
											Args: []string{"echo", "a"},
										},
									}},
								}},
							},
							{
								Pipeline: []*expr.ShellPipeline{{
									Cmd: []*expr.ShellCmd{{
										SimpleCmd: &expr.ShellSimpleCmd{
											Redirect: []*expr.ShellRedirect{{
												Number:   intp(2),
												Token:    token.GreaterAnd,
												Filename: "1",
											}},
											Args: []string{"echo", "b"},
										},
									}},
								}},
							},
						},
					},
				}},
			}}},
			{Pipeline: []*expr.ShellPipeline{{
				Cmd: []*expr.ShellCmd{{
					SimpleCmd: &expr.ShellSimpleCmd{
						Args: []string{"echo", "end"},
					},
				}},
			}}},
		},
	}}}},
	{`GOOS=linux GOARCH=arm64 go build`, &expr.Shell{Cmds: []*expr.ShellList{{
		AndOr: []*expr.ShellAndOr{{Pipeline: []*expr.ShellPipeline{{
			Cmd: []*expr.ShellCmd{{SimpleCmd: &expr.ShellSimpleCmd{
				Assign: []expr.ShellAssign{
					{Key: "GOOS", Value: "linux"},
					{Key: "GOARCH", Value: "arm64"},
				},
				Args: []string{"go", "build"},
			}}},
		}}}},
	}}}},
	{`grep -R "fun*foo" .`, simplesh("grep", "-R", `"fun*foo"`, ".")},
	{`echo -n not_a_file_*`, simplesh("echo", "-n", "not_a_file_*")},
	{`echo -n "\""`, simplesh("echo", "-n", `"\""`)},
	{`echo "a b \"" 'c \' \d "e f'g"`, simplesh(
		"echo", `"a b \""`, `'c \'`, `\d`, `"e f'g"`,
	)},
	{`go build "-ldflags=-v -extldflags=-v" pkg`, simplesh("go", "build", `"-ldflags=-v -extldflags=-v"`, "pkg")},
	{`find . -name \*.c -exec grep -H {} \;
	ls`, &expr.Shell{Cmds: []*expr.ShellList{
		{
			AndOr: []*expr.ShellAndOr{{Pipeline: []*expr.ShellPipeline{{
				Cmd: []*expr.ShellCmd{{SimpleCmd: &expr.ShellSimpleCmd{
					Args: []string{"find", ".", "-name", `\*.c`, "-exec", "grep", "-H", "{}", `\;`},
				}}},
			}}}},
		},
		{
			AndOr: []*expr.ShellAndOr{{Pipeline: []*expr.ShellPipeline{{
				Cmd: []*expr.ShellCmd{{SimpleCmd: &expr.ShellSimpleCmd{
					Args: []string{"ls"},
				}}},
			}}}},
		},
	}}},
	{`echo -n a${VAL}c `, simplesh("echo", "-n", "a${VAL}c")},
	// TODO {`ls \
	//-l`, simplesh(`ls`, `-l`)},
	// TODO: test unbalanced paren errors
}

func simplesh(args ...string) *expr.Shell {
	return &expr.Shell{Cmds: []*expr.ShellList{{
		AndOr: []*expr.ShellAndOr{{Pipeline: []*expr.ShellPipeline{{
			Cmd: []*expr.ShellCmd{{SimpleCmd: &expr.ShellSimpleCmd{
				Args: args,
			}}},
		}}}},
	}}}
}

func TestParseShell(t *testing.T) {
	for _, test := range shellTests {
		fmt.Printf("Parsing %q\n", test.input)
		s, err := parser.ParseStmt([]byte("($$ " + test.input + " $$)"))
		if err != nil {
			t.Errorf("ParseExpr(%q): error: %v", test.input, err)
			continue
		}
		if s == nil {
			t.Errorf("ParseExpr(%q): nil stmt", test.input)
			continue
		}
		got := s.(*stmt.Simple).Expr.(*expr.Unary).Expr.(*expr.Shell)
		if !parser.EqualExpr(got, test.want) {
			t.Errorf("ParseExpr(%q) = %v\ndiff: %s", test.input, format.Debug(got), format.Diff(test.want, got))
		}
	}
}

type stmtTest struct {
	input string
	want  stmt.Stmt
}

var stmtTests = []stmtTest{
	{"for {}", &stmt.For{Body: &stmt.Block{}}},
	{"for ;; {}", &stmt.For{Body: &stmt.Block{}}},
	{"for true {}", &stmt.For{Cond: &expr.Ident{Name: "true"}, Body: &stmt.Block{}}},
	{"for ; true; {}", &stmt.For{Cond: &expr.Ident{Name: "true"}, Body: &stmt.Block{}}},
	{"for range x {}", &stmt.Range{Expr: &expr.Ident{Name: "x"}, Body: &stmt.Block{}}},
	{"for k, v := range x {}", &stmt.Range{
		Key:  &expr.Ident{Name: "k"},
		Val:  &expr.Ident{Name: "v"},
		Expr: &expr.Ident{Name: "x"},
		Body: &stmt.Block{},
	}},
	{"for k := range x {}", &stmt.Range{
		Key:  &expr.Ident{Name: "k"},
		Expr: &expr.Ident{Name: "x"},
		Body: &stmt.Block{},
	}},
	{
		"for i := 0; i < 10; i++ { x = i }",
		&stmt.For{
			Init: &stmt.Assign{
				Decl:  true,
				Left:  []expr.Expr{&expr.Ident{Name: "i"}},
				Right: []expr.Expr{&expr.BasicLiteral{Value: big.NewInt(0)}},
			},
			Cond: &expr.Binary{
				Op:    token.Less,
				Left:  &expr.Ident{Name: "i"},
				Right: &expr.BasicLiteral{Value: big.NewInt(10)},
			},
			Post: &stmt.Assign{
				Left: []expr.Expr{&expr.Ident{Name: "i"}},
				Right: []expr.Expr{
					&expr.Binary{
						Op:    token.Add,
						Left:  &expr.Ident{Name: "i"},
						Right: &expr.BasicLiteral{Value: big.NewInt(1)},
					},
				},
			},
			Body: &stmt.Block{Stmts: []stmt.Stmt{&stmt.Assign{
				Left:  []expr.Expr{&expr.Ident{Name: "x"}},
				Right: []expr.Expr{&expr.Ident{Name: "i"}},
			}}},
		},
	},
	{
		"const x = 4",
		&stmt.Const{NameList: []string{"x"}, Values: []expr.Expr{basic(4)}},
	},
	{
		"const x int64 = 4",
		&stmt.Const{NameList: []string{"x"}, Type: tint64, Values: []expr.Expr{basic(4)}},
	},
	{
		"const i, j = 4, 5",
		&stmt.Const{
			NameList: []string{"i", "j"},
			Values:   []expr.Expr{basic(4), basic(5)},
		},
	},
	{
		"const i, j int64 = 4, 5",
		&stmt.Const{
			NameList: []string{"i", "j"},
			Type:     tint64,
			Values:   []expr.Expr{basic(4), basic(5)},
		},
	},
	{
		"const i, j = 4, 5.0",
		&stmt.Const{
			NameList: []string{"i", "j"},
			Values:   []expr.Expr{basic(4), basic(5.0)},
		},
	},
	{
		`const (
			x = 4
		)`,
		&stmt.ConstSet{
			Consts: []*stmt.Const{
				{NameList: []string{"x"}, Values: []expr.Expr{basic(4)}},
			},
		},
	},
	{
		`const (
			x int64 = 4
		)`,
		&stmt.ConstSet{
			Consts: []*stmt.Const{
				{NameList: []string{"x"}, Type: tint64, Values: []expr.Expr{basic(4)}},
			},
		},
	},
	{
		`const (
			x, y = 4, 5
		)`,
		&stmt.ConstSet{
			Consts: []*stmt.Const{
				{NameList: []string{"x", "y"}, Values: []expr.Expr{basic(4), basic(5)}},
			},
		},
	},
	{
		`const (
			x, y int64 = 4, 5
		)`,
		&stmt.ConstSet{
			Consts: []*stmt.Const{
				{NameList: []string{"x", "y"}, Type: tint64, Values: []expr.Expr{basic(4), basic(5)}},
			},
		},
	},
	{
		`const (
			x, y = 1, 2
			z, w = 3, 4
		)`,
		&stmt.ConstSet{
			Consts: []*stmt.Const{
				{NameList: []string{"x", "y"}, Values: []expr.Expr{basic(1), basic(2)}},
				{NameList: []string{"z", "w"}, Values: []expr.Expr{basic(3), basic(4)}},
			},
		},
	},
	{
		`const (
			x, y int64   = 1, 2
			z, w float64 = 3, 4
		)`,
		&stmt.ConstSet{
			Consts: []*stmt.Const{
				{NameList: []string{"x", "y"}, Type: tint64, Values: []expr.Expr{basic(1), basic(2)}},
				{NameList: []string{"z", "w"}, Type: tfloat64, Values: []expr.Expr{basic(3), basic(4)}},
			},
		},
	},
	{"x.y", &stmt.Simple{Expr: &expr.Selector{Left: &expr.Ident{Name: "x"}, Right: &expr.Ident{Name: "y"}}}},
	{
		`type A integer`,
		&stmt.TypeDecl{Name: "A", Type: &tipe.Named{Name: "A", Type: tinteger}},
	},
	{
		"type Array [2]int",
		&stmt.TypeDecl{
			Name: "Array",
			Type: &tipe.Named{
				Name: "Array",
				Type: &tipe.Array{Len: 2, Elem: &tipe.Unresolved{Name: "int"}},
			}},
	},
	{
		`type S struct { x integer }`,
		&stmt.TypeDecl{
			Name: "S",
			Type: &tipe.Named{
				Name: "S",
				Type: &tipe.Struct{Fields: []tipe.StructField{{Name: "x", Type: tinteger}}},
			},
		},
	},
	{"type T struct { S }", &stmt.TypeDecl{
		Name: "T",
		Type: &tipe.Named{
			Type: &tipe.Struct{Fields: []tipe.StructField{{Name: "S", Type: &tipe.Unresolved{Name: "S"}, Embedded: true}}},
			Name: "T",
		},
	}},
	{`type T struct {
		S
	}`, &stmt.TypeDecl{
		Name: "T",
		Type: &tipe.Named{
			Type: &tipe.Struct{Fields: []tipe.StructField{{Name: "S", Type: &tipe.Unresolved{Name: "S"}, Embedded: true}}},
			Name: "T",
		},
	}},
	{"type T struct { *S }", &stmt.TypeDecl{
		Name: "T",
		Type: &tipe.Named{
			Type: &tipe.Struct{Fields: []tipe.StructField{{Name: "S", Type: &tipe.Pointer{Elem: &tipe.Unresolved{Name: "S"}}, Embedded: true}}},
			Name: "T",
		},
	}},
	{`type T struct {
		*S
	}`, &stmt.TypeDecl{
		Name: "T",
		Type: &tipe.Named{
			Type: &tipe.Struct{Fields: []tipe.StructField{{Name: "S", Type: &tipe.Pointer{Elem: &tipe.Unresolved{Name: "S"}}, Embedded: true}}},
			Name: "T",
		},
	}},
	{"type T struct { A string `json` }", &stmt.TypeDecl{
		Name: "T",
		Type: &tipe.Named{
			Type: &tipe.Struct{Fields: []tipe.StructField{{
				Name: "A",
				Type: &tipe.Unresolved{Name: "string"},
				Tag:  `json`,
			}}},
		},
	}},
	{"type T struct { A string \"json\" }", &stmt.TypeDecl{
		Name: "T",
		Type: &tipe.Named{
			Type: &tipe.Struct{Fields: []tipe.StructField{{
				Name: "A",
				Type: &tipe.Unresolved{Name: "string"},
				Tag:  "json",
			}}},
		},
	}},
	{"type T struct { A string `json:\"a\"` }", &stmt.TypeDecl{
		Name: "T",
		Type: &tipe.Named{
			Type: &tipe.Struct{Fields: []tipe.StructField{{
				Name: "A",
				Type: &tipe.Unresolved{Name: "string"},
				Tag:  `json:"a"`,
			}}},
		},
	}},
	{`type T struct {
		_ [4]byte
		N string
		_ [4]byte
	}`, &stmt.TypeDecl{
		Name: "T",
		Type: &tipe.Named{
			Type: &tipe.Struct{Fields: []tipe.StructField{
				{
					Name: "_",
					Type: &tipe.Array{Len: 4, Elem: &tipe.Unresolved{Name: "byte"}},
				},
				{
					Name: "N",
					Type: &tipe.Unresolved{Name: "string"},
				},
				{
					Name: "_",
					Type: &tipe.Array{Len: 4, Elem: &tipe.Unresolved{Name: "byte"}},
				},
			}},
		},
	}},
	{`type T struct { X, Y int }`, &stmt.TypeDecl{
		Name: "T",
		Type: &tipe.Named{
			Type: &tipe.Struct{Fields: []tipe.StructField{
				{Name: "X", Type: &tipe.Unresolved{Name: "int"}},
				{Name: "Y", Type: &tipe.Unresolved{Name: "int"}},
			}},
		},
	}},
	{`type T struct {
		X, Y int
	}`, &stmt.TypeDecl{
		Name: "T",
		Type: &tipe.Named{
			Type: &tipe.Struct{Fields: []tipe.StructField{
				{Name: "X", Type: &tipe.Unresolved{Name: "int"}},
				{Name: "Y", Type: &tipe.Unresolved{Name: "int"}},
			}},
		},
	}},
	{
		`type (
			T int64
			S struct { x int64 }
		)`,
		&stmt.TypeDeclSet{TypeDecls: []*stmt.TypeDecl{
			{Name: "T", Type: &tipe.Named{
				Name: "T",
				Type: &tipe.Unresolved{Name: "int64"},
			}},
			{Name: "S", Type: &tipe.Named{
				Name: "S",
				Type: &tipe.Struct{Fields: []tipe.StructField{{
					Name: "x",
					Type: &tipe.Unresolved{Name: "int64"},
				}}},
			}}}},
	},
	{
		`methodik AnInt integer {
			func (a) f() integer { return a }
		}
		`,
		&stmt.MethodikDecl{
			Name: "AnInt",
			Type: &tipe.Named{
				Type:        tinteger,
				MethodNames: []string{"f"},
				Methods: []*tipe.Func{{
					Params:  &tipe.Tuple{},
					Results: &tipe.Tuple{Elems: []tipe.Type{tinteger}},
				}},
			},
			Methods: []*expr.FuncLiteral{{
				Name:         "f",
				ReceiverName: "a",
				Type: &tipe.Func{
					Params:  &tipe.Tuple{},
					Results: &tipe.Tuple{Elems: []tipe.Type{tinteger}},
				},
				Body: &stmt.Block{Stmts: []stmt.Stmt{
					&stmt.Return{Exprs: []expr.Expr{&expr.Ident{Name: "a"}}},
				}},
			}},
		},
	},
	{
		`methodik T *struct{
			x integer
			y [|]int64
		} {
			func (a) f(x integer) integer {
				return a.x
			}
		}
		`,
		&stmt.MethodikDecl{
			Name: "T",
			Type: &tipe.Named{
				Type: &tipe.Pointer{Elem: &tipe.Struct{Fields: []tipe.StructField{
					{Name: "x", Type: tinteger},
					{Name: "y", Type: &tipe.Table{Type: tint64}},
				}}},
				MethodNames: []string{"f"},
				Methods: []*tipe.Func{{
					Params:  &tipe.Tuple{Elems: []tipe.Type{tinteger}},
					Results: &tipe.Tuple{Elems: []tipe.Type{tinteger}},
				}},
			},
			Methods: []*expr.FuncLiteral{{
				Name:            "f",
				ReceiverName:    "a",
				PointerReceiver: true,
				Type: &tipe.Func{
					Params:  &tipe.Tuple{Elems: []tipe.Type{tinteger}},
					Results: &tipe.Tuple{Elems: []tipe.Type{tinteger}},
				},
				ParamNames: []string{"x"},
				Body: &stmt.Block{Stmts: []stmt.Stmt{
					&stmt.Return{Exprs: []expr.Expr{&expr.Selector{
						Left:  &expr.Ident{Name: "a"},
						Right: &expr.Ident{Name: "x"},
					}}},
				}},
			}},
		},
	},
	{"S{ X: 7 }", &stmt.Simple{Expr: &expr.CompLiteral{
		Type:   &tipe.Unresolved{Name: "S"},
		Keys:   []expr.Expr{&expr.Ident{Name: "X"}},
		Values: []expr.Expr{&expr.BasicLiteral{Value: big.NewInt(7)}},
	}}},
	{`map[string]string{ "foo": "bar" }`, &stmt.Simple{Expr: &expr.MapLiteral{
		Type:   &tipe.Map{Key: &tipe.Unresolved{Name: "string"}, Value: &tipe.Unresolved{Name: "string"}},
		Keys:   []expr.Expr{basic("foo")},
		Values: []expr.Expr{basic("bar")},
	}}},
	{"x.y", &stmt.Simple{Expr: &expr.Selector{Left: &expr.Ident{Name: "x"}, Right: &expr.Ident{Name: "y"}}}},
	{"sync.Mutex{}", &stmt.Simple{Expr: &expr.CompLiteral{
		Type: &tipe.Unresolved{Package: "sync", Name: "Mutex"},
	}}},
	{"_ = 5", &stmt.Assign{Left: []expr.Expr{&expr.Ident{Name: "_"}}, Right: []expr.Expr{basic(5)}}},
	{"x, _ := 4, 5", &stmt.Assign{
		Decl:  true,
		Left:  []expr.Expr{&expr.Ident{Name: "x"}, &expr.Ident{Name: "_"}},
		Right: []expr.Expr{basic(4), basic(5)},
	}},
	{`if x == y && y == z {}`, &stmt.If{
		Cond: &expr.Binary{
			Op:    token.LogicalAnd,
			Left:  &expr.Binary{Op: token.Equal, Left: &expr.Ident{Name: "x"}, Right: &expr.Ident{Name: "y"}},
			Right: &expr.Binary{Op: token.Equal, Left: &expr.Ident{Name: "y"}, Right: &expr.Ident{Name: "z"}},
		},
		Body: &stmt.Block{},
	}},
	{`if (x == T{}) {}`, &stmt.If{
		Cond: &expr.Unary{
			Op: token.LeftParen,
			Expr: &expr.Binary{
				Op:    token.Equal,
				Left:  &expr.Ident{Name: "x"},
				Right: &expr.CompLiteral{Type: &tipe.Unresolved{Name: "T"}},
			},
		},
		Body: &stmt.Block{},
	}},
	{
		`f(x, // a comment
		y)`,
		&stmt.Simple{Expr: &expr.Call{
			Func: &expr.Ident{Name: "f"},
			Args: []expr.Expr{&expr.Ident{Name: "x"}, &expr.Ident{Name: "y"}},
		}},
	},
	{
		`for {
			x := 4 // a comment
			x = 5
		}`,
		&stmt.For{
			Body: &stmt.Block{Stmts: []stmt.Stmt{
				&stmt.Assign{Left: []expr.Expr{&expr.Ident{Name: "x"}}, Right: []expr.Expr{basic(4)}},
				&stmt.Assign{Left: []expr.Expr{&expr.Ident{Name: "x"}}, Right: []expr.Expr{basic(5)}},
			}},
		},
	},
	{`go func() {}()`, &stmt.Go{Call: &expr.Call{
		Func: &expr.FuncLiteral{
			Type: &tipe.Func{Params: &tipe.Tuple{}},
			Body: &stmt.Block{},
		},
	}}},
	{"switch {}", &stmt.Switch{}},
	{`switch {
	case true:
		print(true)
	case false:
		print(false)
	default:
		print(42)
	}`,
		&stmt.Switch{
			Cases: []stmt.SwitchCase{
				{
					Conds: []expr.Expr{
						&expr.Ident{
							Name: "true",
						},
					},
					Body: &stmt.Block{
						Stmts: []stmt.Stmt{
							&stmt.Simple{
								Expr: &expr.Call{
									Func: &expr.Ident{Name: "print"},
									Args: []expr.Expr{&expr.Ident{Name: "true"}},
								},
							},
						},
					},
				},
				{
					Conds: []expr.Expr{&expr.Ident{Name: "false"}},
					Body: &stmt.Block{
						Stmts: []stmt.Stmt{
							&stmt.Simple{
								Expr: &expr.Call{
									Func: &expr.Ident{Name: "print"},
									Args: []expr.Expr{&expr.Ident{Name: "false"}},
								},
							},
						},
					},
				},
				{
					Default: true,
					Body: &stmt.Block{
						Stmts: []stmt.Stmt{
							&stmt.Simple{
								Expr: &expr.Call{
									Func: &expr.Ident{Name: "print"},
									Args: []expr.Expr{&expr.BasicLiteral{Value: big.NewInt(42)}},
								},
							},
						},
					},
				},
			},
		},
	},
	{`switch i := fct(); i {
	case 42, 66:
		print(i)
	default:
		print(ok)
	}`,
		&stmt.Switch{
			Init: &stmt.Assign{
				Decl: true,
				Left: []expr.Expr{
					&expr.Ident{
						Name: "i",
					},
				},
				Right: []expr.Expr{
					&expr.Call{
						Func: &expr.Ident{
							Name: "fct",
						},
					},
				},
			},
			Cond: &expr.Ident{
				Name: "i",
			},
			Cases: []stmt.SwitchCase{
				{
					Conds: []expr.Expr{
						&expr.BasicLiteral{Value: big.NewInt(42)},
						&expr.BasicLiteral{Value: big.NewInt(66)},
					},
					Body: &stmt.Block{
						Stmts: []stmt.Stmt{
							&stmt.Simple{
								Expr: &expr.Call{
									Func: &expr.Ident{Name: "print"},
									Args: []expr.Expr{&expr.Ident{Name: "i"}},
								},
							},
						},
					},
				},
				{
					Default: true,
					Body: &stmt.Block{
						Stmts: []stmt.Stmt{
							&stmt.Simple{
								Expr: &expr.Call{
									Func: &expr.Ident{Name: "print"},
									Args: []expr.Expr{&expr.Ident{Name: "ok"}},
								},
							},
						},
					},
				},
			},
		},
	},
	{
		"switch v.(type) {}",
		&stmt.TypeSwitch{
			Init:   nil,
			Assign: &stmt.Simple{Expr: &expr.TypeAssert{Left: &expr.Ident{Name: "v"}}},
		},
	},
	{
		"switch x := v.(type) {}",
		&stmt.TypeSwitch{
			Init: nil,
			Assign: &stmt.Assign{
				Decl: true,
				Left: []expr.Expr{
					&expr.Ident{Name: "x"},
				},
				Right: []expr.Expr{
					&expr.TypeAssert{Left: &expr.Ident{Name: "v"}},
				},
			},
		},
	},
	{
		"switch x := fct(); x.(type) {}",
		&stmt.TypeSwitch{
			Init: &stmt.Assign{
				Decl:  true,
				Left:  []expr.Expr{&expr.Ident{Name: "x"}},
				Right: []expr.Expr{&expr.Call{Func: &expr.Ident{Name: "fct"}}},
			},
			Assign: &stmt.Simple{Expr: &expr.TypeAssert{Left: &expr.Ident{Name: "x"}}},
		},
	},
	{
		"switch x := fct(); v := x.(type) {}",
		&stmt.TypeSwitch{
			Init: &stmt.Assign{
				Decl:  true,
				Left:  []expr.Expr{&expr.Ident{Name: "x"}},
				Right: []expr.Expr{&expr.Call{Func: &expr.Ident{Name: "fct"}}},
			},
			Assign: &stmt.Assign{
				Decl: true,
				Left: []expr.Expr{
					&expr.Ident{Name: "v"},
				},
				Right: []expr.Expr{
					&expr.TypeAssert{Left: &expr.Ident{Name: "x"}},
				},
			},
		},
	},
	{
		"switch x, y := f(); v := g(x, y).(type) {}",
		&stmt.TypeSwitch{
			Init: &stmt.Assign{
				Decl:  true,
				Left:  []expr.Expr{&expr.Ident{Name: "x"}, &expr.Ident{Name: "y"}},
				Right: []expr.Expr{&expr.Call{Func: &expr.Ident{Name: "f"}}},
			},
			Assign: &stmt.Assign{
				Decl: true,
				Left: []expr.Expr{
					&expr.Ident{Name: "v"},
				},
				Right: []expr.Expr{
					&expr.TypeAssert{Left: &expr.Call{
						Func: &expr.Ident{Name: "g"},
						Args: []expr.Expr{&expr.Ident{Name: "x"}, &expr.Ident{Name: "y"}},
					}},
				},
			},
		},
	},
	{
		`switch x := fct(); x.(type) {
		case int, float64:
		case *int:
		default:
		}
		`,
		&stmt.TypeSwitch{
			Init: &stmt.Assign{
				Decl:  true,
				Left:  []expr.Expr{&expr.Ident{Name: "x"}},
				Right: []expr.Expr{&expr.Call{Func: &expr.Ident{Name: "fct"}}},
			},
			Assign: &stmt.Simple{Expr: &expr.TypeAssert{Left: &expr.Ident{Name: "x"}}},
			Cases: []stmt.TypeSwitchCase{
				{
					Types: []tipe.Type{&tipe.Unresolved{Package: "", Name: "int"}, &tipe.Unresolved{Package: "", Name: "float64"}},
					Body:  &stmt.Block{},
				},
				{
					Types: []tipe.Type{&tipe.Pointer{Elem: &tipe.Unresolved{Package: "", Name: "int"}}},
					Body:  &stmt.Block{},
				},
				{
					Default: true,
					Body:    &stmt.Block{},
				},
			},
		},
	},
	{"select {}", &stmt.Select{}},
	{`select {
	case v := <-ch1:
		print(v)
	case v, ok := <-ch2:
		print(v, ok)
	case ch3 <- vv:
		print(ch3)
	case <-ch4:
		print(ch4)
	default:
		print(42)
	}`,
		&stmt.Select{
			Cases: []stmt.SelectCase{
				{
					Stmt: &stmt.Assign{
						Decl: true,
						Left: []expr.Expr{
							&expr.Ident{
								Name: "v",
							},
						},
						Right: []expr.Expr{
							&expr.Unary{
								Op: token.ChanOp,
								Expr: &expr.Ident{
									Name: "ch1",
								},
							},
						},
					},
					Body: &stmt.Block{
						Stmts: []stmt.Stmt{
							&stmt.Simple{
								Expr: &expr.Call{
									Func: &expr.Ident{Name: "print"},
									Args: []expr.Expr{&expr.Ident{Name: "v"}},
								},
							},
						},
					},
				},
				{
					Stmt: &stmt.Assign{
						Decl:  true,
						Left:  []expr.Expr{&expr.Ident{Name: "v"}, &expr.Ident{Name: "ok"}},
						Right: []expr.Expr{&expr.Unary{Op: token.ChanOp, Expr: &expr.Ident{Name: "ch2"}}},
					},
					Body: &stmt.Block{
						Stmts: []stmt.Stmt{
							&stmt.Simple{
								Expr: &expr.Call{
									Func: &expr.Ident{Name: "print"},
									Args: []expr.Expr{&expr.Ident{Name: "v"}, &expr.Ident{Name: "ok"}},
								},
							},
						},
					},
				},
				{
					Stmt: &stmt.Send{Chan: &expr.Ident{Name: "ch3"}, Value: &expr.Ident{Name: "vv"}},
					Body: &stmt.Block{
						Stmts: []stmt.Stmt{
							&stmt.Simple{
								Expr: &expr.Call{
									Func: &expr.Ident{Name: "print"},
									Args: []expr.Expr{&expr.Ident{Name: "ch3"}},
								},
							},
						},
					},
				},
				{
					Stmt: &stmt.Simple{Expr: &expr.Unary{Op: token.ChanOp, Expr: &expr.Ident{Name: "ch4"}}},
					Body: &stmt.Block{
						Stmts: []stmt.Stmt{
							&stmt.Simple{
								Expr: &expr.Call{
									Func: &expr.Ident{Name: "print"},
									Args: []expr.Expr{&expr.Ident{Name: "ch4"}},
								},
							},
						},
					},
				},
				{
					Default: true,
					Body: &stmt.Block{
						Stmts: []stmt.Stmt{
							&stmt.Simple{
								Expr: &expr.Call{
									Func: &expr.Ident{Name: "print"},
									Args: []expr.Expr{&expr.BasicLiteral{Value: big.NewInt(42)}},
								},
							},
						},
					},
				},
			},
		},
	},
	{
		`
		select {
		default:
		return
		}
		`, &stmt.Select{
			Cases: []stmt.SelectCase{
				{
					Default: true,
					Body: &stmt.Block{
						Stmts: []stmt.Stmt{
							&stmt.Return{},
						},
					},
				},
			},
		},
	},
	{
		`
		select {
		default:
		return 1
		}
		`, &stmt.Select{
			Cases: []stmt.SelectCase{
				{
					Default: true,
					Body: &stmt.Block{
						Stmts: []stmt.Stmt{
							&stmt.Return{
								Exprs: []expr.Expr{
									&expr.BasicLiteral{
										Value: big.NewInt(1),
									},
								},
							},
						},
					},
				},
			},
		},
	},
	{"return", &stmt.Return{}},
	{"return 1", &stmt.Return{Exprs: []expr.Expr{&expr.BasicLiteral{Value: big.NewInt(1)}}}},
	{"{ return }", &stmt.Block{Stmts: []stmt.Stmt{&stmt.Return{}}}},
	{"{ return 1 }", &stmt.Block{Stmts: []stmt.Stmt{&stmt.Return{Exprs: []expr.Expr{&expr.BasicLiteral{Value: big.NewInt(1)}}}}}},
	{"var i = 10", &stmt.Var{
		NameList: []string{"i"},
		Values:   []expr.Expr{basic(10)},
	}},
	{"var i int", &stmt.Var{
		NameList: []string{"i"},
		Type:     &tipe.Unresolved{Name: "int"},
	}},
	{"var i int = 11", &stmt.Var{
		NameList: []string{"i"},
		Values:   []expr.Expr{basic(11)},
		Type:     &tipe.Unresolved{Name: "int"},
	}},
	{"var i, j = 1, 2", &stmt.Var{
		NameList: []string{"i", "j"},
		Values:   []expr.Expr{basic(1), basic(2)},
	}},
	{"var i, j int64 = 1, 2", &stmt.Var{
		NameList: []string{"i", "j"},
		Type:     tint64,
		Values:   []expr.Expr{basic(1), basic(2)},
	}},
	{"var i, j int64", &stmt.Var{
		NameList: []string{"i", "j"},
		Type:     tint64,
	}},
	{"var i map[string]int", &stmt.Var{
		NameList: []string{"i"},
		Type: &tipe.Map{
			Key:   &tipe.Unresolved{Name: "string"},
			Value: &tipe.Unresolved{Name: "int"},
		},
	}},
	{"var i chan int", &stmt.Var{
		NameList: []string{"i"},
		Type:     &tipe.Chan{Elem: &tipe.Unresolved{Name: "int"}},
	}},
	{"var i []int", &stmt.Var{
		NameList: []string{"i"},
		Type:     &tipe.Slice{Elem: &tipe.Unresolved{Name: "int"}},
	}},
	{"var i [2]int", &stmt.Var{
		NameList: []string{"i"},
		Type: &tipe.Array{
			Len:  2,
			Elem: &tipe.Unresolved{Name: "int"},
		},
	}},
	{"var i = [2]int{1,2}", &stmt.Var{
		NameList: []string{"i"},
		Values: []expr.Expr{&expr.ArrayLiteral{
			Type: &tipe.Array{
				Len:  2,
				Elem: &tipe.Unresolved{Name: "int"},
			},
			Values: []expr.Expr{basic(1), basic(2)},
		}},
	}},
	{"var i = [2]int{1:2}", &stmt.Var{
		NameList: []string{"i"},
		Values: []expr.Expr{&expr.ArrayLiteral{
			Type: &tipe.Array{
				Len:  2,
				Elem: &tipe.Unresolved{Name: "int"},
			},
			Keys:   []expr.Expr{basic(1)},
			Values: []expr.Expr{basic(2)},
		}},
	}},
	{"var i = [...]int{1,2}", &stmt.Var{
		NameList: []string{"i"},
		Values: []expr.Expr{&expr.ArrayLiteral{
			Type: &tipe.Array{
				Len:      2,
				Elem:     &tipe.Unresolved{Name: "int"},
				Ellipsis: true,
			},
			Values: []expr.Expr{basic(1), basic(2)},
		}},
	}},
	{"var i = [...]int{1:2}", &stmt.Var{
		NameList: []string{"i"},
		Values: []expr.Expr{&expr.ArrayLiteral{
			Type: &tipe.Array{
				Len:      2,
				Elem:     &tipe.Unresolved{Name: "int"},
				Ellipsis: true,
			},
			Keys:   []expr.Expr{basic(1)},
			Values: []expr.Expr{basic(2)},
		}},
	}},
	{"var i = []int{1:2}", &stmt.Var{
		NameList: []string{"i"},
		Values: []expr.Expr{&expr.SliceLiteral{
			Type: &tipe.Slice{
				Elem: &tipe.Unresolved{Name: "int"},
			},
			Keys:   []expr.Expr{basic(1)},
			Values: []expr.Expr{basic(2)},
		}},
	}},
	{"var i string", &stmt.Var{
		NameList: []string{"i"},
		Type:     &tipe.Unresolved{Name: "string"},
	}},
	{"var i struct{}", &stmt.Var{
		NameList: []string{"i"},
		Type:     &tipe.Struct{},
	}},
	{
		`var (
			i int = 11
			j = 22
			k float64
		)
		`, &stmt.VarSet{
			Vars: []*stmt.Var{
				{
					NameList: []string{"i"},
					Values:   []expr.Expr{basic(11)},
					Type:     &tipe.Unresolved{Name: "int"},
				},
				{
					NameList: []string{"j"},
					Values:   []expr.Expr{basic(22)},
				},
				{
					NameList: []string{"k"},
					Type:     tfloat64,
				},
			},
		},
	},
	{"0x0", &stmt.Simple{Expr: basic(0x0)}},
	{"0x1", &stmt.Simple{Expr: basic(0x1)}},
	{"0xdeadbeef", &stmt.Simple{Expr: basic(0xdeadbeef)}},
	{"0xDEADBEEF", &stmt.Simple{Expr: basic(0xDEADBEEF)}},
	{"0xdEadb33f", &stmt.Simple{Expr: basic(0xdEadb33f)}},
	{"0X0", &stmt.Simple{Expr: basic(0X0)}},
	{"0X1", &stmt.Simple{Expr: basic(0X1)}},
	{"0Xdeadbeef", &stmt.Simple{Expr: basic(0Xdeadbeef)}},
	{"0XDEADBEEF", &stmt.Simple{Expr: basic(0XDEADBEEF)}},
	{"0XdEadb33f", &stmt.Simple{Expr: basic(0XdEadb33f)}},
	{"defer f()", &stmt.Defer{Expr: &expr.Call{Func: &expr.Ident{Name: "f"}}}},
	{"defer f.Close()", &stmt.Defer{Expr: &expr.Call{
		Func: &expr.Selector{
			Left:  &expr.Ident{Name: "f"},
			Right: &expr.Ident{Name: "Close"},
		},
	}}},
	{"defer f(a, b)", &stmt.Defer{Expr: &expr.Call{
		Func: &expr.Ident{Name: "f"},
		Args: []expr.Expr{&expr.Ident{Name: "a"}, &expr.Ident{Name: "b"}},
	}}},
	{"defer func(){}()", &stmt.Defer{Expr: &expr.Call{
		Func: &expr.FuncLiteral{
			Type: &tipe.Func{Params: &tipe.Tuple{}},
			Body: &stmt.Block{},
		},
	}}},
}

func TestParseStmt(t *testing.T) {
	for _, test := range stmtTests {
		fmt.Printf("Parsing stmt %q\n", test.input)
		got, err := parser.ParseStmt([]byte(test.input))
		if err != nil {
			t.Errorf("ParseStmt(%q): error: %v", test.input, err)
			continue
		}
		if got == nil {
			t.Errorf("ParseStmt(%q): nil stmt", test.input)
			continue
		}
		if !parser.EqualStmt(got, test.want) {
			diff := format.Diff(test.want, got)
			if diff == "" {
				t.Errorf("ParseStmt(%q): format.Diff empty but statements not equal", test.input)
			} else {
				t.Errorf("ParseStmt(%q):\n%v", test.input, diff)
			}
		}
	}
}

func basic(x interface{}) *expr.BasicLiteral {
	switch x := x.(type) {
	case int:
		return &expr.BasicLiteral{Value: big.NewInt(int64(x))}
	case int64:
		return &expr.BasicLiteral{Value: big.NewInt(x)}
	case string:
		return &expr.BasicLiteral{Value: x}
	case float64:
		return &expr.BasicLiteral{Value: big.NewFloat(x)}
	default:
		panic(fmt.Sprintf("unknown basic %v (%T)", x, x))
	}
}

func intp(x int) *int {
	return &x
}
