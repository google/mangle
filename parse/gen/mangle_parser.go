// Code generated from parse/gen/Mangle.g4 by ANTLR 4.13.1. DO NOT EDIT.

package gen // Mangle
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type MangleParser struct {
	*antlr.BaseParser
}

var MangleParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func mangleParserInit() {
	staticData := &MangleParserStaticData
	staticData.LiteralNames = []string{
		"", "'.'", "'descr'", "'inclusion'", "':'", "'{'", "'}'", "'opt'", "",
		"", "'\\u27F8'", "'Package'", "'Use'", "'Decl'", "'bound'", "'let'",
		"'do'", "'('", "')'", "'['", "']'", "'='", "'!='", "','", "'!'", "'<'",
		"'<='", "'>'", "'>='", "':-'", "'\\n'", "'|>'",
	}
	staticData.SymbolicNames = []string{
		"", "", "", "", "", "", "", "", "WHITESPACE", "COMMENT", "LONGLEFTDOUBLEARROW",
		"PACKAGE", "USE", "DECL", "BOUND", "LET", "DO", "LPAREN", "RPAREN",
		"LBRACKET", "RBRACKET", "EQ", "BANGEQ", "COMMA", "BANG", "LESS", "LESSEQ",
		"GREATER", "GREATEREQ", "COLONDASH", "NEWLINE", "PIPEGREATER", "NUMBER",
		"FLOAT", "VARIABLE", "NAME", "TYPENAME", "DOT_TYPE", "CONSTANT", "STRING",
		"BYTESTRING",
	}
	staticData.RuleNames = []string{
		"start", "program", "packageDecl", "useDecl", "decl", "descrBlock",
		"boundsBlock", "constraintsBlock", "clause", "clauseBody", "transform",
		"letStmt", "literalOrFml", "term", "member", "atom", "atoms",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 40, 283, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 1, 0, 1, 0, 1, 0, 1, 1, 3, 1, 39, 8, 1, 1, 1, 5, 1, 42, 8,
		1, 10, 1, 12, 1, 45, 9, 1, 1, 1, 1, 1, 5, 1, 49, 8, 1, 10, 1, 12, 1, 52,
		9, 1, 1, 2, 1, 2, 1, 2, 3, 2, 57, 8, 2, 1, 2, 1, 2, 1, 3, 1, 3, 1, 3, 3,
		3, 64, 8, 3, 1, 3, 1, 3, 1, 4, 1, 4, 1, 4, 3, 4, 71, 8, 4, 1, 4, 5, 4,
		74, 8, 4, 10, 4, 12, 4, 77, 9, 4, 1, 4, 3, 4, 80, 8, 4, 1, 4, 1, 4, 1,
		5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 5, 6, 92, 8, 6, 10, 6, 12,
		6, 95, 9, 6, 1, 6, 3, 6, 98, 8, 6, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 1, 8,
		1, 8, 1, 8, 3, 8, 108, 8, 8, 1, 8, 1, 8, 1, 9, 1, 9, 1, 9, 5, 9, 115, 8,
		9, 10, 9, 12, 9, 118, 9, 9, 1, 9, 3, 9, 121, 8, 9, 1, 9, 1, 9, 5, 9, 125,
		8, 9, 10, 9, 12, 9, 128, 9, 9, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10,
		5, 10, 136, 8, 10, 10, 10, 12, 10, 139, 9, 10, 3, 10, 141, 8, 10, 1, 10,
		1, 10, 1, 10, 5, 10, 146, 8, 10, 10, 10, 12, 10, 149, 9, 10, 3, 10, 151,
		8, 10, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 12, 1, 12, 1, 12, 3, 12, 161,
		8, 12, 1, 12, 1, 12, 3, 12, 165, 8, 12, 1, 13, 1, 13, 1, 13, 1, 13, 1,
		13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 5, 13, 177, 8, 13, 10, 13, 12, 13,
		180, 9, 13, 1, 13, 3, 13, 183, 8, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13,
		1, 13, 1, 13, 5, 13, 192, 8, 13, 10, 13, 12, 13, 195, 9, 13, 1, 13, 1,
		13, 1, 13, 1, 13, 3, 13, 201, 8, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13,
		1, 13, 1, 13, 5, 13, 210, 8, 13, 10, 13, 12, 13, 213, 9, 13, 1, 13, 1,
		13, 1, 13, 1, 13, 3, 13, 219, 8, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13,
		1, 13, 5, 13, 227, 8, 13, 10, 13, 12, 13, 230, 9, 13, 1, 13, 1, 13, 3,
		13, 234, 8, 13, 3, 13, 236, 8, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1,
		13, 5, 13, 244, 8, 13, 10, 13, 12, 13, 247, 9, 13, 1, 13, 3, 13, 250, 8,
		13, 1, 13, 3, 13, 253, 8, 13, 1, 14, 1, 14, 1, 14, 3, 14, 258, 8, 14, 1,
		14, 1, 14, 1, 14, 1, 14, 1, 14, 3, 14, 265, 8, 14, 1, 15, 1, 15, 1, 16,
		1, 16, 1, 16, 1, 16, 5, 16, 273, 8, 16, 10, 16, 12, 16, 276, 9, 16, 1,
		16, 3, 16, 279, 8, 16, 1, 16, 1, 16, 1, 16, 0, 0, 17, 0, 2, 4, 6, 8, 10,
		12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 0, 2, 2, 0, 10, 10, 29, 29,
		2, 0, 21, 22, 25, 28, 311, 0, 34, 1, 0, 0, 0, 2, 38, 1, 0, 0, 0, 4, 53,
		1, 0, 0, 0, 6, 60, 1, 0, 0, 0, 8, 67, 1, 0, 0, 0, 10, 83, 1, 0, 0, 0, 12,
		86, 1, 0, 0, 0, 14, 101, 1, 0, 0, 0, 16, 104, 1, 0, 0, 0, 18, 111, 1, 0,
		0, 0, 20, 150, 1, 0, 0, 0, 22, 152, 1, 0, 0, 0, 24, 164, 1, 0, 0, 0, 26,
		252, 1, 0, 0, 0, 28, 264, 1, 0, 0, 0, 30, 266, 1, 0, 0, 0, 32, 268, 1,
		0, 0, 0, 34, 35, 3, 2, 1, 0, 35, 36, 5, 0, 0, 1, 36, 1, 1, 0, 0, 0, 37,
		39, 3, 4, 2, 0, 38, 37, 1, 0, 0, 0, 38, 39, 1, 0, 0, 0, 39, 43, 1, 0, 0,
		0, 40, 42, 3, 6, 3, 0, 41, 40, 1, 0, 0, 0, 42, 45, 1, 0, 0, 0, 43, 41,
		1, 0, 0, 0, 43, 44, 1, 0, 0, 0, 44, 50, 1, 0, 0, 0, 45, 43, 1, 0, 0, 0,
		46, 49, 3, 8, 4, 0, 47, 49, 3, 16, 8, 0, 48, 46, 1, 0, 0, 0, 48, 47, 1,
		0, 0, 0, 49, 52, 1, 0, 0, 0, 50, 48, 1, 0, 0, 0, 50, 51, 1, 0, 0, 0, 51,
		3, 1, 0, 0, 0, 52, 50, 1, 0, 0, 0, 53, 54, 5, 11, 0, 0, 54, 56, 5, 35,
		0, 0, 55, 57, 3, 32, 16, 0, 56, 55, 1, 0, 0, 0, 56, 57, 1, 0, 0, 0, 57,
		58, 1, 0, 0, 0, 58, 59, 5, 24, 0, 0, 59, 5, 1, 0, 0, 0, 60, 61, 5, 12,
		0, 0, 61, 63, 5, 35, 0, 0, 62, 64, 3, 32, 16, 0, 63, 62, 1, 0, 0, 0, 63,
		64, 1, 0, 0, 0, 64, 65, 1, 0, 0, 0, 65, 66, 5, 24, 0, 0, 66, 7, 1, 0, 0,
		0, 67, 68, 5, 13, 0, 0, 68, 70, 3, 30, 15, 0, 69, 71, 3, 10, 5, 0, 70,
		69, 1, 0, 0, 0, 70, 71, 1, 0, 0, 0, 71, 75, 1, 0, 0, 0, 72, 74, 3, 12,
		6, 0, 73, 72, 1, 0, 0, 0, 74, 77, 1, 0, 0, 0, 75, 73, 1, 0, 0, 0, 75, 76,
		1, 0, 0, 0, 76, 79, 1, 0, 0, 0, 77, 75, 1, 0, 0, 0, 78, 80, 3, 14, 7, 0,
		79, 78, 1, 0, 0, 0, 79, 80, 1, 0, 0, 0, 80, 81, 1, 0, 0, 0, 81, 82, 5,
		1, 0, 0, 82, 9, 1, 0, 0, 0, 83, 84, 5, 2, 0, 0, 84, 85, 3, 32, 16, 0, 85,
		11, 1, 0, 0, 0, 86, 87, 5, 14, 0, 0, 87, 93, 5, 19, 0, 0, 88, 89, 3, 26,
		13, 0, 89, 90, 5, 23, 0, 0, 90, 92, 1, 0, 0, 0, 91, 88, 1, 0, 0, 0, 92,
		95, 1, 0, 0, 0, 93, 91, 1, 0, 0, 0, 93, 94, 1, 0, 0, 0, 94, 97, 1, 0, 0,
		0, 95, 93, 1, 0, 0, 0, 96, 98, 3, 26, 13, 0, 97, 96, 1, 0, 0, 0, 97, 98,
		1, 0, 0, 0, 98, 99, 1, 0, 0, 0, 99, 100, 5, 20, 0, 0, 100, 13, 1, 0, 0,
		0, 101, 102, 5, 3, 0, 0, 102, 103, 3, 32, 16, 0, 103, 15, 1, 0, 0, 0, 104,
		107, 3, 30, 15, 0, 105, 106, 7, 0, 0, 0, 106, 108, 3, 18, 9, 0, 107, 105,
		1, 0, 0, 0, 107, 108, 1, 0, 0, 0, 108, 109, 1, 0, 0, 0, 109, 110, 5, 1,
		0, 0, 110, 17, 1, 0, 0, 0, 111, 116, 3, 24, 12, 0, 112, 113, 5, 23, 0,
		0, 113, 115, 3, 24, 12, 0, 114, 112, 1, 0, 0, 0, 115, 118, 1, 0, 0, 0,
		116, 114, 1, 0, 0, 0, 116, 117, 1, 0, 0, 0, 117, 120, 1, 0, 0, 0, 118,
		116, 1, 0, 0, 0, 119, 121, 5, 23, 0, 0, 120, 119, 1, 0, 0, 0, 120, 121,
		1, 0, 0, 0, 121, 126, 1, 0, 0, 0, 122, 123, 5, 31, 0, 0, 123, 125, 3, 20,
		10, 0, 124, 122, 1, 0, 0, 0, 125, 128, 1, 0, 0, 0, 126, 124, 1, 0, 0, 0,
		126, 127, 1, 0, 0, 0, 127, 19, 1, 0, 0, 0, 128, 126, 1, 0, 0, 0, 129, 130,
		5, 16, 0, 0, 130, 140, 3, 26, 13, 0, 131, 132, 5, 23, 0, 0, 132, 137, 3,
		22, 11, 0, 133, 134, 5, 23, 0, 0, 134, 136, 3, 22, 11, 0, 135, 133, 1,
		0, 0, 0, 136, 139, 1, 0, 0, 0, 137, 135, 1, 0, 0, 0, 137, 138, 1, 0, 0,
		0, 138, 141, 1, 0, 0, 0, 139, 137, 1, 0, 0, 0, 140, 131, 1, 0, 0, 0, 140,
		141, 1, 0, 0, 0, 141, 151, 1, 0, 0, 0, 142, 147, 3, 22, 11, 0, 143, 144,
		5, 23, 0, 0, 144, 146, 3, 22, 11, 0, 145, 143, 1, 0, 0, 0, 146, 149, 1,
		0, 0, 0, 147, 145, 1, 0, 0, 0, 147, 148, 1, 0, 0, 0, 148, 151, 1, 0, 0,
		0, 149, 147, 1, 0, 0, 0, 150, 129, 1, 0, 0, 0, 150, 142, 1, 0, 0, 0, 151,
		21, 1, 0, 0, 0, 152, 153, 5, 15, 0, 0, 153, 154, 5, 34, 0, 0, 154, 155,
		5, 21, 0, 0, 155, 156, 3, 26, 13, 0, 156, 23, 1, 0, 0, 0, 157, 160, 3,
		26, 13, 0, 158, 159, 7, 1, 0, 0, 159, 161, 3, 26, 13, 0, 160, 158, 1, 0,
		0, 0, 160, 161, 1, 0, 0, 0, 161, 165, 1, 0, 0, 0, 162, 163, 5, 24, 0, 0,
		163, 165, 3, 26, 13, 0, 164, 157, 1, 0, 0, 0, 164, 162, 1, 0, 0, 0, 165,
		25, 1, 0, 0, 0, 166, 253, 5, 34, 0, 0, 167, 253, 5, 38, 0, 0, 168, 253,
		5, 32, 0, 0, 169, 253, 5, 33, 0, 0, 170, 253, 5, 39, 0, 0, 171, 253, 5,
		40, 0, 0, 172, 178, 5, 19, 0, 0, 173, 174, 3, 26, 13, 0, 174, 175, 5, 23,
		0, 0, 175, 177, 1, 0, 0, 0, 176, 173, 1, 0, 0, 0, 177, 180, 1, 0, 0, 0,
		178, 176, 1, 0, 0, 0, 178, 179, 1, 0, 0, 0, 179, 182, 1, 0, 0, 0, 180,
		178, 1, 0, 0, 0, 181, 183, 3, 26, 13, 0, 182, 181, 1, 0, 0, 0, 182, 183,
		1, 0, 0, 0, 183, 184, 1, 0, 0, 0, 184, 253, 5, 20, 0, 0, 185, 193, 5, 19,
		0, 0, 186, 187, 3, 26, 13, 0, 187, 188, 5, 4, 0, 0, 188, 189, 3, 26, 13,
		0, 189, 190, 5, 23, 0, 0, 190, 192, 1, 0, 0, 0, 191, 186, 1, 0, 0, 0, 192,
		195, 1, 0, 0, 0, 193, 191, 1, 0, 0, 0, 193, 194, 1, 0, 0, 0, 194, 200,
		1, 0, 0, 0, 195, 193, 1, 0, 0, 0, 196, 197, 3, 26, 13, 0, 197, 198, 5,
		4, 0, 0, 198, 199, 3, 26, 13, 0, 199, 201, 1, 0, 0, 0, 200, 196, 1, 0,
		0, 0, 200, 201, 1, 0, 0, 0, 201, 202, 1, 0, 0, 0, 202, 253, 5, 20, 0, 0,
		203, 211, 5, 5, 0, 0, 204, 205, 3, 26, 13, 0, 205, 206, 5, 4, 0, 0, 206,
		207, 3, 26, 13, 0, 207, 208, 5, 23, 0, 0, 208, 210, 1, 0, 0, 0, 209, 204,
		1, 0, 0, 0, 210, 213, 1, 0, 0, 0, 211, 209, 1, 0, 0, 0, 211, 212, 1, 0,
		0, 0, 212, 218, 1, 0, 0, 0, 213, 211, 1, 0, 0, 0, 214, 215, 3, 26, 13,
		0, 215, 216, 5, 4, 0, 0, 216, 217, 3, 26, 13, 0, 217, 219, 1, 0, 0, 0,
		218, 214, 1, 0, 0, 0, 218, 219, 1, 0, 0, 0, 219, 220, 1, 0, 0, 0, 220,
		253, 5, 6, 0, 0, 221, 222, 5, 37, 0, 0, 222, 228, 5, 25, 0, 0, 223, 224,
		3, 28, 14, 0, 224, 225, 5, 23, 0, 0, 225, 227, 1, 0, 0, 0, 226, 223, 1,
		0, 0, 0, 227, 230, 1, 0, 0, 0, 228, 226, 1, 0, 0, 0, 228, 229, 1, 0, 0,
		0, 229, 235, 1, 0, 0, 0, 230, 228, 1, 0, 0, 0, 231, 233, 3, 28, 14, 0,
		232, 234, 5, 23, 0, 0, 233, 232, 1, 0, 0, 0, 233, 234, 1, 0, 0, 0, 234,
		236, 1, 0, 0, 0, 235, 231, 1, 0, 0, 0, 235, 236, 1, 0, 0, 0, 236, 237,
		1, 0, 0, 0, 237, 253, 5, 27, 0, 0, 238, 239, 5, 35, 0, 0, 239, 245, 5,
		17, 0, 0, 240, 241, 3, 26, 13, 0, 241, 242, 5, 23, 0, 0, 242, 244, 1, 0,
		0, 0, 243, 240, 1, 0, 0, 0, 244, 247, 1, 0, 0, 0, 245, 243, 1, 0, 0, 0,
		245, 246, 1, 0, 0, 0, 246, 249, 1, 0, 0, 0, 247, 245, 1, 0, 0, 0, 248,
		250, 3, 26, 13, 0, 249, 248, 1, 0, 0, 0, 249, 250, 1, 0, 0, 0, 250, 251,
		1, 0, 0, 0, 251, 253, 5, 18, 0, 0, 252, 166, 1, 0, 0, 0, 252, 167, 1, 0,
		0, 0, 252, 168, 1, 0, 0, 0, 252, 169, 1, 0, 0, 0, 252, 170, 1, 0, 0, 0,
		252, 171, 1, 0, 0, 0, 252, 172, 1, 0, 0, 0, 252, 185, 1, 0, 0, 0, 252,
		203, 1, 0, 0, 0, 252, 221, 1, 0, 0, 0, 252, 238, 1, 0, 0, 0, 253, 27, 1,
		0, 0, 0, 254, 257, 3, 26, 13, 0, 255, 256, 5, 4, 0, 0, 256, 258, 3, 26,
		13, 0, 257, 255, 1, 0, 0, 0, 257, 258, 1, 0, 0, 0, 258, 265, 1, 0, 0, 0,
		259, 260, 5, 7, 0, 0, 260, 261, 3, 26, 13, 0, 261, 262, 5, 4, 0, 0, 262,
		263, 3, 26, 13, 0, 263, 265, 1, 0, 0, 0, 264, 254, 1, 0, 0, 0, 264, 259,
		1, 0, 0, 0, 265, 29, 1, 0, 0, 0, 266, 267, 3, 26, 13, 0, 267, 31, 1, 0,
		0, 0, 268, 274, 5, 19, 0, 0, 269, 270, 3, 30, 15, 0, 270, 271, 5, 23, 0,
		0, 271, 273, 1, 0, 0, 0, 272, 269, 1, 0, 0, 0, 273, 276, 1, 0, 0, 0, 274,
		272, 1, 0, 0, 0, 274, 275, 1, 0, 0, 0, 275, 278, 1, 0, 0, 0, 276, 274,
		1, 0, 0, 0, 277, 279, 3, 30, 15, 0, 278, 277, 1, 0, 0, 0, 278, 279, 1,
		0, 0, 0, 279, 280, 1, 0, 0, 0, 280, 281, 5, 20, 0, 0, 281, 33, 1, 0, 0,
		0, 37, 38, 43, 48, 50, 56, 63, 70, 75, 79, 93, 97, 107, 116, 120, 126,
		137, 140, 147, 150, 160, 164, 178, 182, 193, 200, 211, 218, 228, 233, 235,
		245, 249, 252, 257, 264, 274, 278,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// MangleParserInit initializes any static state used to implement MangleParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewMangleParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func MangleParserInit() {
	staticData := &MangleParserStaticData
	staticData.once.Do(mangleParserInit)
}

// NewMangleParser produces a new parser instance for the optional input antlr.TokenStream.
func NewMangleParser(input antlr.TokenStream) *MangleParser {
	MangleParserInit()
	this := new(MangleParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &MangleParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "Mangle.g4"

	return this
}

// MangleParser tokens.
const (
	MangleParserEOF                 = antlr.TokenEOF
	MangleParserT__0                = 1
	MangleParserT__1                = 2
	MangleParserT__2                = 3
	MangleParserT__3                = 4
	MangleParserT__4                = 5
	MangleParserT__5                = 6
	MangleParserT__6                = 7
	MangleParserWHITESPACE          = 8
	MangleParserCOMMENT             = 9
	MangleParserLONGLEFTDOUBLEARROW = 10
	MangleParserPACKAGE             = 11
	MangleParserUSE                 = 12
	MangleParserDECL                = 13
	MangleParserBOUND               = 14
	MangleParserLET                 = 15
	MangleParserDO                  = 16
	MangleParserLPAREN              = 17
	MangleParserRPAREN              = 18
	MangleParserLBRACKET            = 19
	MangleParserRBRACKET            = 20
	MangleParserEQ                  = 21
	MangleParserBANGEQ              = 22
	MangleParserCOMMA               = 23
	MangleParserBANG                = 24
	MangleParserLESS                = 25
	MangleParserLESSEQ              = 26
	MangleParserGREATER             = 27
	MangleParserGREATEREQ           = 28
	MangleParserCOLONDASH           = 29
	MangleParserNEWLINE             = 30
	MangleParserPIPEGREATER         = 31
	MangleParserNUMBER              = 32
	MangleParserFLOAT               = 33
	MangleParserVARIABLE            = 34
	MangleParserNAME                = 35
	MangleParserTYPENAME            = 36
	MangleParserDOT_TYPE            = 37
	MangleParserCONSTANT            = 38
	MangleParserSTRING              = 39
	MangleParserBYTESTRING          = 40
)

// MangleParser rules.
const (
	MangleParserRULE_start            = 0
	MangleParserRULE_program          = 1
	MangleParserRULE_packageDecl      = 2
	MangleParserRULE_useDecl          = 3
	MangleParserRULE_decl             = 4
	MangleParserRULE_descrBlock       = 5
	MangleParserRULE_boundsBlock      = 6
	MangleParserRULE_constraintsBlock = 7
	MangleParserRULE_clause           = 8
	MangleParserRULE_clauseBody       = 9
	MangleParserRULE_transform        = 10
	MangleParserRULE_letStmt          = 11
	MangleParserRULE_literalOrFml     = 12
	MangleParserRULE_term             = 13
	MangleParserRULE_member           = 14
	MangleParserRULE_atom             = 15
	MangleParserRULE_atoms            = 16
)

// IStartContext is an interface to support dynamic dispatch.
type IStartContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Program() IProgramContext
	EOF() antlr.TerminalNode

	// IsStartContext differentiates from other interfaces.
	IsStartContext()
}

type StartContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStartContext() *StartContext {
	var p = new(StartContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_start
	return p
}

func InitEmptyStartContext(p *StartContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_start
}

func (*StartContext) IsStartContext() {}

func NewStartContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StartContext {
	var p = new(StartContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_start

	return p
}

func (s *StartContext) GetParser() antlr.Parser { return s.parser }

func (s *StartContext) Program() IProgramContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IProgramContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IProgramContext)
}

func (s *StartContext) EOF() antlr.TerminalNode {
	return s.GetToken(MangleParserEOF, 0)
}

func (s *StartContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StartContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *StartContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterStart(s)
	}
}

func (s *StartContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitStart(s)
	}
}

func (s *StartContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitStart(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) Start_() (localctx IStartContext) {
	localctx = NewStartContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, MangleParserRULE_start)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(34)
		p.Program()
	}
	{
		p.SetState(35)
		p.Match(MangleParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IProgramContext is an interface to support dynamic dispatch.
type IProgramContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	PackageDecl() IPackageDeclContext
	AllUseDecl() []IUseDeclContext
	UseDecl(i int) IUseDeclContext
	AllDecl() []IDeclContext
	Decl(i int) IDeclContext
	AllClause() []IClauseContext
	Clause(i int) IClauseContext

	// IsProgramContext differentiates from other interfaces.
	IsProgramContext()
}

type ProgramContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyProgramContext() *ProgramContext {
	var p = new(ProgramContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_program
	return p
}

func InitEmptyProgramContext(p *ProgramContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_program
}

func (*ProgramContext) IsProgramContext() {}

func NewProgramContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ProgramContext {
	var p = new(ProgramContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_program

	return p
}

func (s *ProgramContext) GetParser() antlr.Parser { return s.parser }

func (s *ProgramContext) PackageDecl() IPackageDeclContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPackageDeclContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPackageDeclContext)
}

func (s *ProgramContext) AllUseDecl() []IUseDeclContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IUseDeclContext); ok {
			len++
		}
	}

	tst := make([]IUseDeclContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IUseDeclContext); ok {
			tst[i] = t.(IUseDeclContext)
			i++
		}
	}

	return tst
}

func (s *ProgramContext) UseDecl(i int) IUseDeclContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUseDeclContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUseDeclContext)
}

func (s *ProgramContext) AllDecl() []IDeclContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IDeclContext); ok {
			len++
		}
	}

	tst := make([]IDeclContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IDeclContext); ok {
			tst[i] = t.(IDeclContext)
			i++
		}
	}

	return tst
}

func (s *ProgramContext) Decl(i int) IDeclContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDeclContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDeclContext)
}

func (s *ProgramContext) AllClause() []IClauseContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IClauseContext); ok {
			len++
		}
	}

	tst := make([]IClauseContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IClauseContext); ok {
			tst[i] = t.(IClauseContext)
			i++
		}
	}

	return tst
}

func (s *ProgramContext) Clause(i int) IClauseContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IClauseContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IClauseContext)
}

func (s *ProgramContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ProgramContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ProgramContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterProgram(s)
	}
}

func (s *ProgramContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitProgram(s)
	}
}

func (s *ProgramContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitProgram(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) Program() (localctx IProgramContext) {
	localctx = NewProgramContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, MangleParserRULE_program)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(38)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MangleParserPACKAGE {
		{
			p.SetState(37)
			p.PackageDecl()
		}

	}
	p.SetState(43)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == MangleParserUSE {
		{
			p.SetState(40)
			p.UseDecl()
		}

		p.SetState(45)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(50)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2126009344032) != 0 {
		p.SetState(48)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case MangleParserDECL:
			{
				p.SetState(46)
				p.Decl()
			}

		case MangleParserT__4, MangleParserLBRACKET, MangleParserNUMBER, MangleParserFLOAT, MangleParserVARIABLE, MangleParserNAME, MangleParserDOT_TYPE, MangleParserCONSTANT, MangleParserSTRING, MangleParserBYTESTRING:
			{
				p.SetState(47)
				p.Clause()
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(52)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPackageDeclContext is an interface to support dynamic dispatch.
type IPackageDeclContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	PACKAGE() antlr.TerminalNode
	NAME() antlr.TerminalNode
	BANG() antlr.TerminalNode
	Atoms() IAtomsContext

	// IsPackageDeclContext differentiates from other interfaces.
	IsPackageDeclContext()
}

type PackageDeclContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPackageDeclContext() *PackageDeclContext {
	var p = new(PackageDeclContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_packageDecl
	return p
}

func InitEmptyPackageDeclContext(p *PackageDeclContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_packageDecl
}

func (*PackageDeclContext) IsPackageDeclContext() {}

func NewPackageDeclContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PackageDeclContext {
	var p = new(PackageDeclContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_packageDecl

	return p
}

func (s *PackageDeclContext) GetParser() antlr.Parser { return s.parser }

func (s *PackageDeclContext) PACKAGE() antlr.TerminalNode {
	return s.GetToken(MangleParserPACKAGE, 0)
}

func (s *PackageDeclContext) NAME() antlr.TerminalNode {
	return s.GetToken(MangleParserNAME, 0)
}

func (s *PackageDeclContext) BANG() antlr.TerminalNode {
	return s.GetToken(MangleParserBANG, 0)
}

func (s *PackageDeclContext) Atoms() IAtomsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAtomsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAtomsContext)
}

func (s *PackageDeclContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PackageDeclContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PackageDeclContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterPackageDecl(s)
	}
}

func (s *PackageDeclContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitPackageDecl(s)
	}
}

func (s *PackageDeclContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitPackageDecl(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) PackageDecl() (localctx IPackageDeclContext) {
	localctx = NewPackageDeclContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, MangleParserRULE_packageDecl)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(53)
		p.Match(MangleParserPACKAGE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(54)
		p.Match(MangleParserNAME)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(56)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MangleParserLBRACKET {
		{
			p.SetState(55)
			p.Atoms()
		}

	}
	{
		p.SetState(58)
		p.Match(MangleParserBANG)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IUseDeclContext is an interface to support dynamic dispatch.
type IUseDeclContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	USE() antlr.TerminalNode
	NAME() antlr.TerminalNode
	BANG() antlr.TerminalNode
	Atoms() IAtomsContext

	// IsUseDeclContext differentiates from other interfaces.
	IsUseDeclContext()
}

type UseDeclContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUseDeclContext() *UseDeclContext {
	var p = new(UseDeclContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_useDecl
	return p
}

func InitEmptyUseDeclContext(p *UseDeclContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_useDecl
}

func (*UseDeclContext) IsUseDeclContext() {}

func NewUseDeclContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *UseDeclContext {
	var p = new(UseDeclContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_useDecl

	return p
}

func (s *UseDeclContext) GetParser() antlr.Parser { return s.parser }

func (s *UseDeclContext) USE() antlr.TerminalNode {
	return s.GetToken(MangleParserUSE, 0)
}

func (s *UseDeclContext) NAME() antlr.TerminalNode {
	return s.GetToken(MangleParserNAME, 0)
}

func (s *UseDeclContext) BANG() antlr.TerminalNode {
	return s.GetToken(MangleParserBANG, 0)
}

func (s *UseDeclContext) Atoms() IAtomsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAtomsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAtomsContext)
}

func (s *UseDeclContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UseDeclContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *UseDeclContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterUseDecl(s)
	}
}

func (s *UseDeclContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitUseDecl(s)
	}
}

func (s *UseDeclContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitUseDecl(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) UseDecl() (localctx IUseDeclContext) {
	localctx = NewUseDeclContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, MangleParserRULE_useDecl)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(60)
		p.Match(MangleParserUSE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(61)
		p.Match(MangleParserNAME)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(63)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MangleParserLBRACKET {
		{
			p.SetState(62)
			p.Atoms()
		}

	}
	{
		p.SetState(65)
		p.Match(MangleParserBANG)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IDeclContext is an interface to support dynamic dispatch.
type IDeclContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	DECL() antlr.TerminalNode
	Atom() IAtomContext
	DescrBlock() IDescrBlockContext
	AllBoundsBlock() []IBoundsBlockContext
	BoundsBlock(i int) IBoundsBlockContext
	ConstraintsBlock() IConstraintsBlockContext

	// IsDeclContext differentiates from other interfaces.
	IsDeclContext()
}

type DeclContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDeclContext() *DeclContext {
	var p = new(DeclContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_decl
	return p
}

func InitEmptyDeclContext(p *DeclContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_decl
}

func (*DeclContext) IsDeclContext() {}

func NewDeclContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DeclContext {
	var p = new(DeclContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_decl

	return p
}

func (s *DeclContext) GetParser() antlr.Parser { return s.parser }

func (s *DeclContext) DECL() antlr.TerminalNode {
	return s.GetToken(MangleParserDECL, 0)
}

func (s *DeclContext) Atom() IAtomContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAtomContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAtomContext)
}

func (s *DeclContext) DescrBlock() IDescrBlockContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDescrBlockContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDescrBlockContext)
}

func (s *DeclContext) AllBoundsBlock() []IBoundsBlockContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IBoundsBlockContext); ok {
			len++
		}
	}

	tst := make([]IBoundsBlockContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IBoundsBlockContext); ok {
			tst[i] = t.(IBoundsBlockContext)
			i++
		}
	}

	return tst
}

func (s *DeclContext) BoundsBlock(i int) IBoundsBlockContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBoundsBlockContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBoundsBlockContext)
}

func (s *DeclContext) ConstraintsBlock() IConstraintsBlockContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConstraintsBlockContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConstraintsBlockContext)
}

func (s *DeclContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DeclContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DeclContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterDecl(s)
	}
}

func (s *DeclContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitDecl(s)
	}
}

func (s *DeclContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitDecl(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) Decl() (localctx IDeclContext) {
	localctx = NewDeclContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, MangleParserRULE_decl)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(67)
		p.Match(MangleParserDECL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(68)
		p.Atom()
	}
	p.SetState(70)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MangleParserT__1 {
		{
			p.SetState(69)
			p.DescrBlock()
		}

	}
	p.SetState(75)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == MangleParserBOUND {
		{
			p.SetState(72)
			p.BoundsBlock()
		}

		p.SetState(77)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(79)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MangleParserT__2 {
		{
			p.SetState(78)
			p.ConstraintsBlock()
		}

	}
	{
		p.SetState(81)
		p.Match(MangleParserT__0)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IDescrBlockContext is an interface to support dynamic dispatch.
type IDescrBlockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Atoms() IAtomsContext

	// IsDescrBlockContext differentiates from other interfaces.
	IsDescrBlockContext()
}

type DescrBlockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDescrBlockContext() *DescrBlockContext {
	var p = new(DescrBlockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_descrBlock
	return p
}

func InitEmptyDescrBlockContext(p *DescrBlockContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_descrBlock
}

func (*DescrBlockContext) IsDescrBlockContext() {}

func NewDescrBlockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DescrBlockContext {
	var p = new(DescrBlockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_descrBlock

	return p
}

func (s *DescrBlockContext) GetParser() antlr.Parser { return s.parser }

func (s *DescrBlockContext) Atoms() IAtomsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAtomsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAtomsContext)
}

func (s *DescrBlockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DescrBlockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DescrBlockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterDescrBlock(s)
	}
}

func (s *DescrBlockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitDescrBlock(s)
	}
}

func (s *DescrBlockContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitDescrBlock(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) DescrBlock() (localctx IDescrBlockContext) {
	localctx = NewDescrBlockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, MangleParserRULE_descrBlock)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(83)
		p.Match(MangleParserT__1)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(84)
		p.Atoms()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBoundsBlockContext is an interface to support dynamic dispatch.
type IBoundsBlockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	BOUND() antlr.TerminalNode
	LBRACKET() antlr.TerminalNode
	RBRACKET() antlr.TerminalNode
	AllTerm() []ITermContext
	Term(i int) ITermContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsBoundsBlockContext differentiates from other interfaces.
	IsBoundsBlockContext()
}

type BoundsBlockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBoundsBlockContext() *BoundsBlockContext {
	var p = new(BoundsBlockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_boundsBlock
	return p
}

func InitEmptyBoundsBlockContext(p *BoundsBlockContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_boundsBlock
}

func (*BoundsBlockContext) IsBoundsBlockContext() {}

func NewBoundsBlockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *BoundsBlockContext {
	var p = new(BoundsBlockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_boundsBlock

	return p
}

func (s *BoundsBlockContext) GetParser() antlr.Parser { return s.parser }

func (s *BoundsBlockContext) BOUND() antlr.TerminalNode {
	return s.GetToken(MangleParserBOUND, 0)
}

func (s *BoundsBlockContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(MangleParserLBRACKET, 0)
}

func (s *BoundsBlockContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(MangleParserRBRACKET, 0)
}

func (s *BoundsBlockContext) AllTerm() []ITermContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITermContext); ok {
			len++
		}
	}

	tst := make([]ITermContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITermContext); ok {
			tst[i] = t.(ITermContext)
			i++
		}
	}

	return tst
}

func (s *BoundsBlockContext) Term(i int) ITermContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermContext)
}

func (s *BoundsBlockContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MangleParserCOMMA)
}

func (s *BoundsBlockContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MangleParserCOMMA, i)
}

func (s *BoundsBlockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BoundsBlockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *BoundsBlockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterBoundsBlock(s)
	}
}

func (s *BoundsBlockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitBoundsBlock(s)
	}
}

func (s *BoundsBlockContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitBoundsBlock(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) BoundsBlock() (localctx IBoundsBlockContext) {
	localctx = NewBoundsBlockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, MangleParserRULE_boundsBlock)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(86)
		p.Match(MangleParserBOUND)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(87)
		p.Match(MangleParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(93)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 9, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(88)
				p.Term()
			}
			{
				p.SetState(89)
				p.Match(MangleParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		p.SetState(95)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 9, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	p.SetState(97)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2126009335840) != 0 {
		{
			p.SetState(96)
			p.Term()
		}

	}
	{
		p.SetState(99)
		p.Match(MangleParserRBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IConstraintsBlockContext is an interface to support dynamic dispatch.
type IConstraintsBlockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Atoms() IAtomsContext

	// IsConstraintsBlockContext differentiates from other interfaces.
	IsConstraintsBlockContext()
}

type ConstraintsBlockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyConstraintsBlockContext() *ConstraintsBlockContext {
	var p = new(ConstraintsBlockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_constraintsBlock
	return p
}

func InitEmptyConstraintsBlockContext(p *ConstraintsBlockContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_constraintsBlock
}

func (*ConstraintsBlockContext) IsConstraintsBlockContext() {}

func NewConstraintsBlockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ConstraintsBlockContext {
	var p = new(ConstraintsBlockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_constraintsBlock

	return p
}

func (s *ConstraintsBlockContext) GetParser() antlr.Parser { return s.parser }

func (s *ConstraintsBlockContext) Atoms() IAtomsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAtomsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAtomsContext)
}

func (s *ConstraintsBlockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConstraintsBlockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ConstraintsBlockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterConstraintsBlock(s)
	}
}

func (s *ConstraintsBlockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitConstraintsBlock(s)
	}
}

func (s *ConstraintsBlockContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitConstraintsBlock(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) ConstraintsBlock() (localctx IConstraintsBlockContext) {
	localctx = NewConstraintsBlockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, MangleParserRULE_constraintsBlock)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(101)
		p.Match(MangleParserT__2)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(102)
		p.Atoms()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IClauseContext is an interface to support dynamic dispatch.
type IClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Atom() IAtomContext
	ClauseBody() IClauseBodyContext
	COLONDASH() antlr.TerminalNode
	LONGLEFTDOUBLEARROW() antlr.TerminalNode

	// IsClauseContext differentiates from other interfaces.
	IsClauseContext()
}

type ClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyClauseContext() *ClauseContext {
	var p = new(ClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_clause
	return p
}

func InitEmptyClauseContext(p *ClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_clause
}

func (*ClauseContext) IsClauseContext() {}

func NewClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ClauseContext {
	var p = new(ClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_clause

	return p
}

func (s *ClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *ClauseContext) Atom() IAtomContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAtomContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAtomContext)
}

func (s *ClauseContext) ClauseBody() IClauseBodyContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IClauseBodyContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IClauseBodyContext)
}

func (s *ClauseContext) COLONDASH() antlr.TerminalNode {
	return s.GetToken(MangleParserCOLONDASH, 0)
}

func (s *ClauseContext) LONGLEFTDOUBLEARROW() antlr.TerminalNode {
	return s.GetToken(MangleParserLONGLEFTDOUBLEARROW, 0)
}

func (s *ClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterClause(s)
	}
}

func (s *ClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitClause(s)
	}
}

func (s *ClauseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitClause(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) Clause() (localctx IClauseContext) {
	localctx = NewClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, MangleParserRULE_clause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(104)
		p.Atom()
	}
	p.SetState(107)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MangleParserLONGLEFTDOUBLEARROW || _la == MangleParserCOLONDASH {
		{
			p.SetState(105)
			_la = p.GetTokenStream().LA(1)

			if !(_la == MangleParserLONGLEFTDOUBLEARROW || _la == MangleParserCOLONDASH) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(106)
			p.ClauseBody()
		}

	}
	{
		p.SetState(109)
		p.Match(MangleParserT__0)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IClauseBodyContext is an interface to support dynamic dispatch.
type IClauseBodyContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllLiteralOrFml() []ILiteralOrFmlContext
	LiteralOrFml(i int) ILiteralOrFmlContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	AllPIPEGREATER() []antlr.TerminalNode
	PIPEGREATER(i int) antlr.TerminalNode
	AllTransform() []ITransformContext
	Transform(i int) ITransformContext

	// IsClauseBodyContext differentiates from other interfaces.
	IsClauseBodyContext()
}

type ClauseBodyContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyClauseBodyContext() *ClauseBodyContext {
	var p = new(ClauseBodyContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_clauseBody
	return p
}

func InitEmptyClauseBodyContext(p *ClauseBodyContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_clauseBody
}

func (*ClauseBodyContext) IsClauseBodyContext() {}

func NewClauseBodyContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ClauseBodyContext {
	var p = new(ClauseBodyContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_clauseBody

	return p
}

func (s *ClauseBodyContext) GetParser() antlr.Parser { return s.parser }

func (s *ClauseBodyContext) AllLiteralOrFml() []ILiteralOrFmlContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ILiteralOrFmlContext); ok {
			len++
		}
	}

	tst := make([]ILiteralOrFmlContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ILiteralOrFmlContext); ok {
			tst[i] = t.(ILiteralOrFmlContext)
			i++
		}
	}

	return tst
}

func (s *ClauseBodyContext) LiteralOrFml(i int) ILiteralOrFmlContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteralOrFmlContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILiteralOrFmlContext)
}

func (s *ClauseBodyContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MangleParserCOMMA)
}

func (s *ClauseBodyContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MangleParserCOMMA, i)
}

func (s *ClauseBodyContext) AllPIPEGREATER() []antlr.TerminalNode {
	return s.GetTokens(MangleParserPIPEGREATER)
}

func (s *ClauseBodyContext) PIPEGREATER(i int) antlr.TerminalNode {
	return s.GetToken(MangleParserPIPEGREATER, i)
}

func (s *ClauseBodyContext) AllTransform() []ITransformContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITransformContext); ok {
			len++
		}
	}

	tst := make([]ITransformContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITransformContext); ok {
			tst[i] = t.(ITransformContext)
			i++
		}
	}

	return tst
}

func (s *ClauseBodyContext) Transform(i int) ITransformContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITransformContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITransformContext)
}

func (s *ClauseBodyContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ClauseBodyContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ClauseBodyContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterClauseBody(s)
	}
}

func (s *ClauseBodyContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitClauseBody(s)
	}
}

func (s *ClauseBodyContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitClauseBody(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) ClauseBody() (localctx IClauseBodyContext) {
	localctx = NewClauseBodyContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, MangleParserRULE_clauseBody)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(111)
		p.LiteralOrFml()
	}
	p.SetState(116)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 12, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(112)
				p.Match(MangleParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(113)
				p.LiteralOrFml()
			}

		}
		p.SetState(118)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 12, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	p.SetState(120)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MangleParserCOMMA {
		{
			p.SetState(119)
			p.Match(MangleParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	p.SetState(126)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == MangleParserPIPEGREATER {
		{
			p.SetState(122)
			p.Match(MangleParserPIPEGREATER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(123)
			p.Transform()
		}

		p.SetState(128)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITransformContext is an interface to support dynamic dispatch.
type ITransformContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	DO() antlr.TerminalNode
	Term() ITermContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	AllLetStmt() []ILetStmtContext
	LetStmt(i int) ILetStmtContext

	// IsTransformContext differentiates from other interfaces.
	IsTransformContext()
}

type TransformContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTransformContext() *TransformContext {
	var p = new(TransformContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_transform
	return p
}

func InitEmptyTransformContext(p *TransformContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_transform
}

func (*TransformContext) IsTransformContext() {}

func NewTransformContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TransformContext {
	var p = new(TransformContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_transform

	return p
}

func (s *TransformContext) GetParser() antlr.Parser { return s.parser }

func (s *TransformContext) DO() antlr.TerminalNode {
	return s.GetToken(MangleParserDO, 0)
}

func (s *TransformContext) Term() ITermContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermContext)
}

func (s *TransformContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MangleParserCOMMA)
}

func (s *TransformContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MangleParserCOMMA, i)
}

func (s *TransformContext) AllLetStmt() []ILetStmtContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ILetStmtContext); ok {
			len++
		}
	}

	tst := make([]ILetStmtContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ILetStmtContext); ok {
			tst[i] = t.(ILetStmtContext)
			i++
		}
	}

	return tst
}

func (s *TransformContext) LetStmt(i int) ILetStmtContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILetStmtContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILetStmtContext)
}

func (s *TransformContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TransformContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TransformContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterTransform(s)
	}
}

func (s *TransformContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitTransform(s)
	}
}

func (s *TransformContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitTransform(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) Transform() (localctx ITransformContext) {
	localctx = NewTransformContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, MangleParserRULE_transform)
	var _la int

	p.SetState(150)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MangleParserDO:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(129)
			p.Match(MangleParserDO)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(130)
			p.Term()
		}
		p.SetState(140)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == MangleParserCOMMA {
			{
				p.SetState(131)
				p.Match(MangleParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(132)
				p.LetStmt()
			}
			p.SetState(137)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == MangleParserCOMMA {
				{
					p.SetState(133)
					p.Match(MangleParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(134)
					p.LetStmt()
				}

				p.SetState(139)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		}

	case MangleParserLET:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(142)
			p.LetStmt()
		}
		p.SetState(147)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == MangleParserCOMMA {
			{
				p.SetState(143)
				p.Match(MangleParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(144)
				p.LetStmt()
			}

			p.SetState(149)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILetStmtContext is an interface to support dynamic dispatch.
type ILetStmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LET() antlr.TerminalNode
	VARIABLE() antlr.TerminalNode
	EQ() antlr.TerminalNode
	Term() ITermContext

	// IsLetStmtContext differentiates from other interfaces.
	IsLetStmtContext()
}

type LetStmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLetStmtContext() *LetStmtContext {
	var p = new(LetStmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_letStmt
	return p
}

func InitEmptyLetStmtContext(p *LetStmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_letStmt
}

func (*LetStmtContext) IsLetStmtContext() {}

func NewLetStmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LetStmtContext {
	var p = new(LetStmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_letStmt

	return p
}

func (s *LetStmtContext) GetParser() antlr.Parser { return s.parser }

func (s *LetStmtContext) LET() antlr.TerminalNode {
	return s.GetToken(MangleParserLET, 0)
}

func (s *LetStmtContext) VARIABLE() antlr.TerminalNode {
	return s.GetToken(MangleParserVARIABLE, 0)
}

func (s *LetStmtContext) EQ() antlr.TerminalNode {
	return s.GetToken(MangleParserEQ, 0)
}

func (s *LetStmtContext) Term() ITermContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermContext)
}

func (s *LetStmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LetStmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LetStmtContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterLetStmt(s)
	}
}

func (s *LetStmtContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitLetStmt(s)
	}
}

func (s *LetStmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitLetStmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) LetStmt() (localctx ILetStmtContext) {
	localctx = NewLetStmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, MangleParserRULE_letStmt)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(152)
		p.Match(MangleParserLET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(153)
		p.Match(MangleParserVARIABLE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(154)
		p.Match(MangleParserEQ)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(155)
		p.Term()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILiteralOrFmlContext is an interface to support dynamic dispatch.
type ILiteralOrFmlContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllTerm() []ITermContext
	Term(i int) ITermContext
	EQ() antlr.TerminalNode
	BANGEQ() antlr.TerminalNode
	LESS() antlr.TerminalNode
	LESSEQ() antlr.TerminalNode
	GREATER() antlr.TerminalNode
	GREATEREQ() antlr.TerminalNode
	BANG() antlr.TerminalNode

	// IsLiteralOrFmlContext differentiates from other interfaces.
	IsLiteralOrFmlContext()
}

type LiteralOrFmlContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLiteralOrFmlContext() *LiteralOrFmlContext {
	var p = new(LiteralOrFmlContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_literalOrFml
	return p
}

func InitEmptyLiteralOrFmlContext(p *LiteralOrFmlContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_literalOrFml
}

func (*LiteralOrFmlContext) IsLiteralOrFmlContext() {}

func NewLiteralOrFmlContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LiteralOrFmlContext {
	var p = new(LiteralOrFmlContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_literalOrFml

	return p
}

func (s *LiteralOrFmlContext) GetParser() antlr.Parser { return s.parser }

func (s *LiteralOrFmlContext) AllTerm() []ITermContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITermContext); ok {
			len++
		}
	}

	tst := make([]ITermContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITermContext); ok {
			tst[i] = t.(ITermContext)
			i++
		}
	}

	return tst
}

func (s *LiteralOrFmlContext) Term(i int) ITermContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermContext)
}

func (s *LiteralOrFmlContext) EQ() antlr.TerminalNode {
	return s.GetToken(MangleParserEQ, 0)
}

func (s *LiteralOrFmlContext) BANGEQ() antlr.TerminalNode {
	return s.GetToken(MangleParserBANGEQ, 0)
}

func (s *LiteralOrFmlContext) LESS() antlr.TerminalNode {
	return s.GetToken(MangleParserLESS, 0)
}

func (s *LiteralOrFmlContext) LESSEQ() antlr.TerminalNode {
	return s.GetToken(MangleParserLESSEQ, 0)
}

func (s *LiteralOrFmlContext) GREATER() antlr.TerminalNode {
	return s.GetToken(MangleParserGREATER, 0)
}

func (s *LiteralOrFmlContext) GREATEREQ() antlr.TerminalNode {
	return s.GetToken(MangleParserGREATEREQ, 0)
}

func (s *LiteralOrFmlContext) BANG() antlr.TerminalNode {
	return s.GetToken(MangleParserBANG, 0)
}

func (s *LiteralOrFmlContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LiteralOrFmlContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LiteralOrFmlContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterLiteralOrFml(s)
	}
}

func (s *LiteralOrFmlContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitLiteralOrFml(s)
	}
}

func (s *LiteralOrFmlContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitLiteralOrFml(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) LiteralOrFml() (localctx ILiteralOrFmlContext) {
	localctx = NewLiteralOrFmlContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, MangleParserRULE_literalOrFml)
	var _la int

	p.SetState(164)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MangleParserT__4, MangleParserLBRACKET, MangleParserNUMBER, MangleParserFLOAT, MangleParserVARIABLE, MangleParserNAME, MangleParserDOT_TYPE, MangleParserCONSTANT, MangleParserSTRING, MangleParserBYTESTRING:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(157)
			p.Term()
		}
		p.SetState(160)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&509607936) != 0 {
			{
				p.SetState(158)
				_la = p.GetTokenStream().LA(1)

				if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&509607936) != 0) {
					p.GetErrorHandler().RecoverInline(p)
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
			}
			{
				p.SetState(159)
				p.Term()
			}

		}

	case MangleParserBANG:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(162)
			p.Match(MangleParserBANG)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(163)
			p.Term()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITermContext is an interface to support dynamic dispatch.
type ITermContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsTermContext differentiates from other interfaces.
	IsTermContext()
}

type TermContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTermContext() *TermContext {
	var p = new(TermContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_term
	return p
}

func InitEmptyTermContext(p *TermContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_term
}

func (*TermContext) IsTermContext() {}

func NewTermContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TermContext {
	var p = new(TermContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_term

	return p
}

func (s *TermContext) GetParser() antlr.Parser { return s.parser }

func (s *TermContext) CopyAll(ctx *TermContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *TermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TermContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type StrContext struct {
	TermContext
}

func NewStrContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *StrContext {
	var p = new(StrContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *StrContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StrContext) STRING() antlr.TerminalNode {
	return s.GetToken(MangleParserSTRING, 0)
}

func (s *StrContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterStr(s)
	}
}

func (s *StrContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitStr(s)
	}
}

func (s *StrContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitStr(s)

	default:
		return t.VisitChildren(s)
	}
}

type BStrContext struct {
	TermContext
}

func NewBStrContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *BStrContext {
	var p = new(BStrContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *BStrContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BStrContext) BYTESTRING() antlr.TerminalNode {
	return s.GetToken(MangleParserBYTESTRING, 0)
}

func (s *BStrContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterBStr(s)
	}
}

func (s *BStrContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitBStr(s)
	}
}

func (s *BStrContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitBStr(s)

	default:
		return t.VisitChildren(s)
	}
}

type FloatContext struct {
	TermContext
}

func NewFloatContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *FloatContext {
	var p = new(FloatContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *FloatContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FloatContext) FLOAT() antlr.TerminalNode {
	return s.GetToken(MangleParserFLOAT, 0)
}

func (s *FloatContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterFloat(s)
	}
}

func (s *FloatContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitFloat(s)
	}
}

func (s *FloatContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitFloat(s)

	default:
		return t.VisitChildren(s)
	}
}

type ApplContext struct {
	TermContext
}

func NewApplContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ApplContext {
	var p = new(ApplContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *ApplContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ApplContext) NAME() antlr.TerminalNode {
	return s.GetToken(MangleParserNAME, 0)
}

func (s *ApplContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(MangleParserLPAREN, 0)
}

func (s *ApplContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(MangleParserRPAREN, 0)
}

func (s *ApplContext) AllTerm() []ITermContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITermContext); ok {
			len++
		}
	}

	tst := make([]ITermContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITermContext); ok {
			tst[i] = t.(ITermContext)
			i++
		}
	}

	return tst
}

func (s *ApplContext) Term(i int) ITermContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermContext)
}

func (s *ApplContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MangleParserCOMMA)
}

func (s *ApplContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MangleParserCOMMA, i)
}

func (s *ApplContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterAppl(s)
	}
}

func (s *ApplContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitAppl(s)
	}
}

func (s *ApplContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitAppl(s)

	default:
		return t.VisitChildren(s)
	}
}

type VarContext struct {
	TermContext
}

func NewVarContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *VarContext {
	var p = new(VarContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *VarContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VarContext) VARIABLE() antlr.TerminalNode {
	return s.GetToken(MangleParserVARIABLE, 0)
}

func (s *VarContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterVar(s)
	}
}

func (s *VarContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitVar(s)
	}
}

func (s *VarContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitVar(s)

	default:
		return t.VisitChildren(s)
	}
}

type ConstContext struct {
	TermContext
}

func NewConstContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConstContext {
	var p = new(ConstContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *ConstContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConstContext) CONSTANT() antlr.TerminalNode {
	return s.GetToken(MangleParserCONSTANT, 0)
}

func (s *ConstContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterConst(s)
	}
}

func (s *ConstContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitConst(s)
	}
}

func (s *ConstContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitConst(s)

	default:
		return t.VisitChildren(s)
	}
}

type NumContext struct {
	TermContext
}

func NewNumContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *NumContext {
	var p = new(NumContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *NumContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NumContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(MangleParserNUMBER, 0)
}

func (s *NumContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterNum(s)
	}
}

func (s *NumContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitNum(s)
	}
}

func (s *NumContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitNum(s)

	default:
		return t.VisitChildren(s)
	}
}

type ListContext struct {
	TermContext
}

func NewListContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ListContext {
	var p = new(ListContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *ListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ListContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(MangleParserLBRACKET, 0)
}

func (s *ListContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(MangleParserRBRACKET, 0)
}

func (s *ListContext) AllTerm() []ITermContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITermContext); ok {
			len++
		}
	}

	tst := make([]ITermContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITermContext); ok {
			tst[i] = t.(ITermContext)
			i++
		}
	}

	return tst
}

func (s *ListContext) Term(i int) ITermContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermContext)
}

func (s *ListContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MangleParserCOMMA)
}

func (s *ListContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MangleParserCOMMA, i)
}

func (s *ListContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterList(s)
	}
}

func (s *ListContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitList(s)
	}
}

func (s *ListContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitList(s)

	default:
		return t.VisitChildren(s)
	}
}

type MapContext struct {
	TermContext
}

func NewMapContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *MapContext {
	var p = new(MapContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *MapContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MapContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(MangleParserLBRACKET, 0)
}

func (s *MapContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(MangleParserRBRACKET, 0)
}

func (s *MapContext) AllTerm() []ITermContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITermContext); ok {
			len++
		}
	}

	tst := make([]ITermContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITermContext); ok {
			tst[i] = t.(ITermContext)
			i++
		}
	}

	return tst
}

func (s *MapContext) Term(i int) ITermContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermContext)
}

func (s *MapContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MangleParserCOMMA)
}

func (s *MapContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MangleParserCOMMA, i)
}

func (s *MapContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterMap(s)
	}
}

func (s *MapContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitMap(s)
	}
}

func (s *MapContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitMap(s)

	default:
		return t.VisitChildren(s)
	}
}

type StructContext struct {
	TermContext
}

func NewStructContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *StructContext {
	var p = new(StructContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *StructContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StructContext) AllTerm() []ITermContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITermContext); ok {
			len++
		}
	}

	tst := make([]ITermContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITermContext); ok {
			tst[i] = t.(ITermContext)
			i++
		}
	}

	return tst
}

func (s *StructContext) Term(i int) ITermContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermContext)
}

func (s *StructContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MangleParserCOMMA)
}

func (s *StructContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MangleParserCOMMA, i)
}

func (s *StructContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterStruct(s)
	}
}

func (s *StructContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitStruct(s)
	}
}

func (s *StructContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitStruct(s)

	default:
		return t.VisitChildren(s)
	}
}

type DotTypeContext struct {
	TermContext
}

func NewDotTypeContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DotTypeContext {
	var p = new(DotTypeContext)

	InitEmptyTermContext(&p.TermContext)
	p.parser = parser
	p.CopyAll(ctx.(*TermContext))

	return p
}

func (s *DotTypeContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DotTypeContext) DOT_TYPE() antlr.TerminalNode {
	return s.GetToken(MangleParserDOT_TYPE, 0)
}

func (s *DotTypeContext) LESS() antlr.TerminalNode {
	return s.GetToken(MangleParserLESS, 0)
}

func (s *DotTypeContext) GREATER() antlr.TerminalNode {
	return s.GetToken(MangleParserGREATER, 0)
}

func (s *DotTypeContext) AllMember() []IMemberContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IMemberContext); ok {
			len++
		}
	}

	tst := make([]IMemberContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IMemberContext); ok {
			tst[i] = t.(IMemberContext)
			i++
		}
	}

	return tst
}

func (s *DotTypeContext) Member(i int) IMemberContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMemberContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMemberContext)
}

func (s *DotTypeContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MangleParserCOMMA)
}

func (s *DotTypeContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MangleParserCOMMA, i)
}

func (s *DotTypeContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterDotType(s)
	}
}

func (s *DotTypeContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitDotType(s)
	}
}

func (s *DotTypeContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitDotType(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) Term() (localctx ITermContext) {
	localctx = NewTermContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, MangleParserRULE_term)
	var _la int

	var _alt int

	p.SetState(252)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 32, p.GetParserRuleContext()) {
	case 1:
		localctx = NewVarContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(166)
			p.Match(MangleParserVARIABLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		localctx = NewConstContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(167)
			p.Match(MangleParserCONSTANT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		localctx = NewNumContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(168)
			p.Match(MangleParserNUMBER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		localctx = NewFloatContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(169)
			p.Match(MangleParserFLOAT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 5:
		localctx = NewStrContext(p, localctx)
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(170)
			p.Match(MangleParserSTRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 6:
		localctx = NewBStrContext(p, localctx)
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(171)
			p.Match(MangleParserBYTESTRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 7:
		localctx = NewListContext(p, localctx)
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(172)
			p.Match(MangleParserLBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(178)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 21, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(173)
					p.Term()
				}
				{
					p.SetState(174)
					p.Match(MangleParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(180)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 21, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(182)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2126009335840) != 0 {
			{
				p.SetState(181)
				p.Term()
			}

		}
		{
			p.SetState(184)
			p.Match(MangleParserRBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 8:
		localctx = NewMapContext(p, localctx)
		p.EnterOuterAlt(localctx, 8)
		{
			p.SetState(185)
			p.Match(MangleParserLBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(193)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 23, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(186)
					p.Term()
				}
				{
					p.SetState(187)
					p.Match(MangleParserT__3)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(188)
					p.Term()
				}
				{
					p.SetState(189)
					p.Match(MangleParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(195)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 23, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(200)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2126009335840) != 0 {
			{
				p.SetState(196)
				p.Term()
			}
			{
				p.SetState(197)
				p.Match(MangleParserT__3)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(198)
				p.Term()
			}

		}
		{
			p.SetState(202)
			p.Match(MangleParserRBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 9:
		localctx = NewStructContext(p, localctx)
		p.EnterOuterAlt(localctx, 9)
		{
			p.SetState(203)
			p.Match(MangleParserT__4)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(211)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 25, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(204)
					p.Term()
				}
				{
					p.SetState(205)
					p.Match(MangleParserT__3)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(206)
					p.Term()
				}
				{
					p.SetState(207)
					p.Match(MangleParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(213)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 25, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(218)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2126009335840) != 0 {
			{
				p.SetState(214)
				p.Term()
			}
			{
				p.SetState(215)
				p.Match(MangleParserT__3)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(216)
				p.Term()
			}

		}
		{
			p.SetState(220)
			p.Match(MangleParserT__5)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 10:
		localctx = NewDotTypeContext(p, localctx)
		p.EnterOuterAlt(localctx, 10)
		{
			p.SetState(221)
			p.Match(MangleParserDOT_TYPE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(222)
			p.Match(MangleParserLESS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(228)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 27, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(223)
					p.Member()
				}
				{
					p.SetState(224)
					p.Match(MangleParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(230)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 27, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(235)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2126009335968) != 0 {
			{
				p.SetState(231)
				p.Member()
			}
			p.SetState(233)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			if _la == MangleParserCOMMA {
				{
					p.SetState(232)
					p.Match(MangleParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}

		}
		{
			p.SetState(237)
			p.Match(MangleParserGREATER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 11:
		localctx = NewApplContext(p, localctx)
		p.EnterOuterAlt(localctx, 11)
		{
			p.SetState(238)
			p.Match(MangleParserNAME)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(239)
			p.Match(MangleParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(245)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 30, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(240)
					p.Term()
				}
				{
					p.SetState(241)
					p.Match(MangleParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(247)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 30, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(249)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2126009335840) != 0 {
			{
				p.SetState(248)
				p.Term()
			}

		}
		{
			p.SetState(251)
			p.Match(MangleParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMemberContext is an interface to support dynamic dispatch.
type IMemberContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllTerm() []ITermContext
	Term(i int) ITermContext

	// IsMemberContext differentiates from other interfaces.
	IsMemberContext()
}

type MemberContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMemberContext() *MemberContext {
	var p = new(MemberContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_member
	return p
}

func InitEmptyMemberContext(p *MemberContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_member
}

func (*MemberContext) IsMemberContext() {}

func NewMemberContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MemberContext {
	var p = new(MemberContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_member

	return p
}

func (s *MemberContext) GetParser() antlr.Parser { return s.parser }

func (s *MemberContext) AllTerm() []ITermContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITermContext); ok {
			len++
		}
	}

	tst := make([]ITermContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITermContext); ok {
			tst[i] = t.(ITermContext)
			i++
		}
	}

	return tst
}

func (s *MemberContext) Term(i int) ITermContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermContext)
}

func (s *MemberContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MemberContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MemberContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterMember(s)
	}
}

func (s *MemberContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitMember(s)
	}
}

func (s *MemberContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitMember(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) Member() (localctx IMemberContext) {
	localctx = NewMemberContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, MangleParserRULE_member)
	var _la int

	p.SetState(264)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MangleParserT__4, MangleParserLBRACKET, MangleParserNUMBER, MangleParserFLOAT, MangleParserVARIABLE, MangleParserNAME, MangleParserDOT_TYPE, MangleParserCONSTANT, MangleParserSTRING, MangleParserBYTESTRING:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(254)
			p.Term()
		}
		p.SetState(257)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == MangleParserT__3 {
			{
				p.SetState(255)
				p.Match(MangleParserT__3)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(256)
				p.Term()
			}

		}

	case MangleParserT__6:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(259)
			p.Match(MangleParserT__6)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(260)
			p.Term()
		}
		{
			p.SetState(261)
			p.Match(MangleParserT__3)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(262)
			p.Term()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAtomContext is an interface to support dynamic dispatch.
type IAtomContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Term() ITermContext

	// IsAtomContext differentiates from other interfaces.
	IsAtomContext()
}

type AtomContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAtomContext() *AtomContext {
	var p = new(AtomContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_atom
	return p
}

func InitEmptyAtomContext(p *AtomContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_atom
}

func (*AtomContext) IsAtomContext() {}

func NewAtomContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AtomContext {
	var p = new(AtomContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_atom

	return p
}

func (s *AtomContext) GetParser() antlr.Parser { return s.parser }

func (s *AtomContext) Term() ITermContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITermContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITermContext)
}

func (s *AtomContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AtomContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AtomContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterAtom(s)
	}
}

func (s *AtomContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitAtom(s)
	}
}

func (s *AtomContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitAtom(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) Atom() (localctx IAtomContext) {
	localctx = NewAtomContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, MangleParserRULE_atom)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(266)
		p.Term()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAtomsContext is an interface to support dynamic dispatch.
type IAtomsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LBRACKET() antlr.TerminalNode
	RBRACKET() antlr.TerminalNode
	AllAtom() []IAtomContext
	Atom(i int) IAtomContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsAtomsContext differentiates from other interfaces.
	IsAtomsContext()
}

type AtomsContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAtomsContext() *AtomsContext {
	var p = new(AtomsContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_atoms
	return p
}

func InitEmptyAtomsContext(p *AtomsContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_atoms
}

func (*AtomsContext) IsAtomsContext() {}

func NewAtomsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AtomsContext {
	var p = new(AtomsContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_atoms

	return p
}

func (s *AtomsContext) GetParser() antlr.Parser { return s.parser }

func (s *AtomsContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(MangleParserLBRACKET, 0)
}

func (s *AtomsContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(MangleParserRBRACKET, 0)
}

func (s *AtomsContext) AllAtom() []IAtomContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IAtomContext); ok {
			len++
		}
	}

	tst := make([]IAtomContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IAtomContext); ok {
			tst[i] = t.(IAtomContext)
			i++
		}
	}

	return tst
}

func (s *AtomsContext) Atom(i int) IAtomContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAtomContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAtomContext)
}

func (s *AtomsContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MangleParserCOMMA)
}

func (s *AtomsContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MangleParserCOMMA, i)
}

func (s *AtomsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AtomsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AtomsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterAtoms(s)
	}
}

func (s *AtomsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitAtoms(s)
	}
}

func (s *AtomsContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitAtoms(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) Atoms() (localctx IAtomsContext) {
	localctx = NewAtomsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 32, MangleParserRULE_atoms)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(268)
		p.Match(MangleParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(274)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 35, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(269)
				p.Atom()
			}
			{
				p.SetState(270)
				p.Match(MangleParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		p.SetState(276)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 35, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	p.SetState(278)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2126009335840) != 0 {
		{
			p.SetState(277)
			p.Atom()
		}

	}
	{
		p.SetState(280)
		p.Match(MangleParserRBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}
