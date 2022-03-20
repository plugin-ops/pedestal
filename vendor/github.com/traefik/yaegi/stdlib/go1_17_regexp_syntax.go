// Code generated by 'yaegi extract regexp/syntax'. DO NOT EDIT.

//go:build go1.17
// +build go1.17

package stdlib

import (
	"reflect"
	"regexp/syntax"
)

func init() {
	Symbols["regexp/syntax/syntax"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"ClassNL":                  reflect.ValueOf(syntax.ClassNL),
		"Compile":                  reflect.ValueOf(syntax.Compile),
		"DotNL":                    reflect.ValueOf(syntax.DotNL),
		"EmptyBeginLine":           reflect.ValueOf(syntax.EmptyBeginLine),
		"EmptyBeginText":           reflect.ValueOf(syntax.EmptyBeginText),
		"EmptyEndLine":             reflect.ValueOf(syntax.EmptyEndLine),
		"EmptyEndText":             reflect.ValueOf(syntax.EmptyEndText),
		"EmptyNoWordBoundary":      reflect.ValueOf(syntax.EmptyNoWordBoundary),
		"EmptyOpContext":           reflect.ValueOf(syntax.EmptyOpContext),
		"EmptyWordBoundary":        reflect.ValueOf(syntax.EmptyWordBoundary),
		"ErrInternalError":         reflect.ValueOf(syntax.ErrInternalError),
		"ErrInvalidCharClass":      reflect.ValueOf(syntax.ErrInvalidCharClass),
		"ErrInvalidCharRange":      reflect.ValueOf(syntax.ErrInvalidCharRange),
		"ErrInvalidEscape":         reflect.ValueOf(syntax.ErrInvalidEscape),
		"ErrInvalidNamedCapture":   reflect.ValueOf(syntax.ErrInvalidNamedCapture),
		"ErrInvalidPerlOp":         reflect.ValueOf(syntax.ErrInvalidPerlOp),
		"ErrInvalidRepeatOp":       reflect.ValueOf(syntax.ErrInvalidRepeatOp),
		"ErrInvalidRepeatSize":     reflect.ValueOf(syntax.ErrInvalidRepeatSize),
		"ErrInvalidUTF8":           reflect.ValueOf(syntax.ErrInvalidUTF8),
		"ErrMissingBracket":        reflect.ValueOf(syntax.ErrMissingBracket),
		"ErrMissingParen":          reflect.ValueOf(syntax.ErrMissingParen),
		"ErrMissingRepeatArgument": reflect.ValueOf(syntax.ErrMissingRepeatArgument),
		"ErrTrailingBackslash":     reflect.ValueOf(syntax.ErrTrailingBackslash),
		"ErrUnexpectedParen":       reflect.ValueOf(syntax.ErrUnexpectedParen),
		"FoldCase":                 reflect.ValueOf(syntax.FoldCase),
		"InstAlt":                  reflect.ValueOf(syntax.InstAlt),
		"InstAltMatch":             reflect.ValueOf(syntax.InstAltMatch),
		"InstCapture":              reflect.ValueOf(syntax.InstCapture),
		"InstEmptyWidth":           reflect.ValueOf(syntax.InstEmptyWidth),
		"InstFail":                 reflect.ValueOf(syntax.InstFail),
		"InstMatch":                reflect.ValueOf(syntax.InstMatch),
		"InstNop":                  reflect.ValueOf(syntax.InstNop),
		"InstRune":                 reflect.ValueOf(syntax.InstRune),
		"InstRune1":                reflect.ValueOf(syntax.InstRune1),
		"InstRuneAny":              reflect.ValueOf(syntax.InstRuneAny),
		"InstRuneAnyNotNL":         reflect.ValueOf(syntax.InstRuneAnyNotNL),
		"IsWordChar":               reflect.ValueOf(syntax.IsWordChar),
		"Literal":                  reflect.ValueOf(syntax.Literal),
		"MatchNL":                  reflect.ValueOf(syntax.MatchNL),
		"NonGreedy":                reflect.ValueOf(syntax.NonGreedy),
		"OneLine":                  reflect.ValueOf(syntax.OneLine),
		"OpAlternate":              reflect.ValueOf(syntax.OpAlternate),
		"OpAnyChar":                reflect.ValueOf(syntax.OpAnyChar),
		"OpAnyCharNotNL":           reflect.ValueOf(syntax.OpAnyCharNotNL),
		"OpBeginLine":              reflect.ValueOf(syntax.OpBeginLine),
		"OpBeginText":              reflect.ValueOf(syntax.OpBeginText),
		"OpCapture":                reflect.ValueOf(syntax.OpCapture),
		"OpCharClass":              reflect.ValueOf(syntax.OpCharClass),
		"OpConcat":                 reflect.ValueOf(syntax.OpConcat),
		"OpEmptyMatch":             reflect.ValueOf(syntax.OpEmptyMatch),
		"OpEndLine":                reflect.ValueOf(syntax.OpEndLine),
		"OpEndText":                reflect.ValueOf(syntax.OpEndText),
		"OpLiteral":                reflect.ValueOf(syntax.OpLiteral),
		"OpNoMatch":                reflect.ValueOf(syntax.OpNoMatch),
		"OpNoWordBoundary":         reflect.ValueOf(syntax.OpNoWordBoundary),
		"OpPlus":                   reflect.ValueOf(syntax.OpPlus),
		"OpQuest":                  reflect.ValueOf(syntax.OpQuest),
		"OpRepeat":                 reflect.ValueOf(syntax.OpRepeat),
		"OpStar":                   reflect.ValueOf(syntax.OpStar),
		"OpWordBoundary":           reflect.ValueOf(syntax.OpWordBoundary),
		"POSIX":                    reflect.ValueOf(syntax.POSIX),
		"Parse":                    reflect.ValueOf(syntax.Parse),
		"Perl":                     reflect.ValueOf(syntax.Perl),
		"PerlX":                    reflect.ValueOf(syntax.PerlX),
		"Simple":                   reflect.ValueOf(syntax.Simple),
		"UnicodeGroups":            reflect.ValueOf(syntax.UnicodeGroups),
		"WasDollar":                reflect.ValueOf(syntax.WasDollar),

		// type definitions
		"EmptyOp":   reflect.ValueOf((*syntax.EmptyOp)(nil)),
		"Error":     reflect.ValueOf((*syntax.Error)(nil)),
		"ErrorCode": reflect.ValueOf((*syntax.ErrorCode)(nil)),
		"Flags":     reflect.ValueOf((*syntax.Flags)(nil)),
		"Inst":      reflect.ValueOf((*syntax.Inst)(nil)),
		"InstOp":    reflect.ValueOf((*syntax.InstOp)(nil)),
		"Op":        reflect.ValueOf((*syntax.Op)(nil)),
		"Prog":      reflect.ValueOf((*syntax.Prog)(nil)),
		"Regexp":    reflect.ValueOf((*syntax.Regexp)(nil)),
	}
}
