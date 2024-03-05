package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/antlr4-go/antlr/v4"
	"github.com/octoberswimmer/apexfmt/formatter"
	"github.com/octoberswimmer/apexfmt/parser"
)

type soqlNoWhereListener struct {
	*parser.BaseApexParserListener
	filename string
}

type errorListener struct {
	*antlr.DefaultErrorListener
	errors []string
}

func (e *errorListener) SyntaxError(_ antlr.Recognizer, _ interface{}, line, column int, msg string, _ antlr.RecognitionException) {
	e.errors = append(e.errors, fmt.Sprintln("line "+strconv.Itoa(line)+":"+strconv.Itoa(column)+" "+msg))
}

func (l *soqlNoWhereListener) EnterSoqlLiteral(ctx *parser.SoqlLiteralContext) {
	query := ctx.Query()
	if query.WhereClause() == nil {
		p := ctx.GetParser()
		tokens := p.GetTokenStream()
		input := tokens.GetTokenSource().GetInputStream()
		lexer := parser.NewApexLexer(input)
		stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

		v := formatter.NewFormatVisitor(stream)
		out, ok := v.VisitRule(ctx).(string)
		if !ok {
			fmt.Errorf("Unexpected result formatting")
		}
		token := ctx.GetStart()

		// Get the line number associated with the token
		lineNumber := token.GetLine()

		fmt.Println("No WHERE in", l.filename, "on line", lineNumber, ":", out)
	}
}

var RootCmd = &cobra.Command{
	Use:   "soqlnowhere [file...]",
	Short: "Find SOQL With No WHERE",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, filename := range args {
			err := checkFile(filename)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
		return nil
	},
	DisableFlagsInUseLine: true,
}

func checkFile(filename string) error {
	src, err := readFile(filename)
	if err != nil {
		return fmt.Errorf("Failed to read file %s: %w", filename, err)
	}
	input := antlr.NewInputStream(string(src))
	lexer := parser.NewApexLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := parser.NewApexParser(stream)
	p.RemoveErrorListeners()
	e := &errorListener{}
	p.AddErrorListener(e)

	l := new(soqlNoWhereListener)
	l.filename = filename
	antlr.ParseTreeWalkerDefault.Walk(l, p.CompilationUnit())
	if len(e.errors) > 0 {
		return errors.New(strings.Join(e.errors, "\n"))
	}
	return nil
}

func readFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	src, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return src, nil
}

func main() {
	Execute()
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
