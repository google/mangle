// Code generated from parse/gen/Mangle.g4 by ANTLR 4.13.2. DO NOT EDIT.

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
		"", "'temporal'", "'.'", "'descr'", "'inclusion'", "'now'", "':'", "'opt'",
		"", "", "'\\u27F8'", "'Package'", "'Use'", "'Decl'", "'bound'", "'let'",
		"'do'", "'('", "')'", "'['", "']'", "'{'", "'}'", "'='", "'!='", "','",
		"'!'", "'<='", "'<'", "'>='", "'>'", "':-'", "'\\n'", "'|>'", "'@'",
		"'<-'", "'[-'",
	}
	staticData.SymbolicNames = []string{
		"", "", "", "", "", "", "", "", "WHITESPACE", "COMMENT", "LONGLEFTDOUBLEARROW",
		"PACKAGE", "USE", "DECL", "BOUND", "LET", "DO", "LPAREN", "RPAREN",
		"LBRACKET", "RBRACKET", "LBRACE", "RBRACE", "EQ", "BANGEQ", "COMMA",
		"BANG", "LESSEQ", "LESS", "GREATEREQ", "GREATER", "COLONDASH", "NEWLINE",
		"PIPEGREATER", "AT", "DIAMONDMINUS", "BOXMINUS", "TIMESTAMP", "DURATION",
		"NUMBER", "FLOAT", "VARIABLE", "NAME", "TYPENAME", "DOT_TYPE", "CONSTANT",
		"STRING", "BYTESTRING",
	}
	staticData.RuleNames = []string{
		"start", "program", "packageDecl", "useDecl", "decl", "descrBlock",
		"boundsBlock", "constraintsBlock", "clause", "temporalAnnotation", "temporalBound",
		"clauseBody", "transform", "letStmt", "literalOrFml", "temporalOperator",
		"term", "member", "atom", "atoms",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 47, 328, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 1, 0, 1, 0, 1,
		0, 1, 1, 3, 1, 45, 8, 1, 1, 1, 5, 1, 48, 8, 1, 10, 1, 12, 1, 51, 9, 1,
		1, 1, 1, 1, 5, 1, 55, 8, 1, 10, 1, 12, 1, 58, 9, 1, 1, 2, 1, 2, 1, 2, 3,
		2, 63, 8, 2, 1, 2, 1, 2, 1, 3, 1, 3, 1, 3, 3, 3, 70, 8, 3, 1, 3, 1, 3,
		1, 4, 1, 4, 1, 4, 3, 4, 77, 8, 4, 1, 4, 3, 4, 80, 8, 4, 1, 4, 5, 4, 83,
		8, 4, 10, 4, 12, 4, 86, 9, 4, 1, 4, 3, 4, 89, 8, 4, 1, 4, 1, 4, 1, 5, 1,
		5, 1, 5, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 5, 6, 101, 8, 6, 10, 6, 12, 6, 104,
		9, 6, 1, 6, 3, 6, 107, 8, 6, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 1, 8, 1, 8,
		3, 8, 116, 8, 8, 1, 8, 1, 8, 3, 8, 120, 8, 8, 1, 8, 1, 8, 1, 9, 1, 9, 1,
		9, 1, 9, 1, 9, 3, 9, 129, 8, 9, 1, 9, 1, 9, 1, 10, 1, 10, 1, 11, 1, 11,
		1, 11, 5, 11, 138, 8, 11, 10, 11, 12, 11, 141, 9, 11, 1, 11, 3, 11, 144,
		8, 11, 1, 11, 1, 11, 5, 11, 148, 8, 11, 10, 11, 12, 11, 151, 9, 11, 1,
		12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 5, 12, 159, 8, 12, 10, 12, 12, 12,
		162, 9, 12, 3, 12, 164, 8, 12, 1, 12, 1, 12, 1, 12, 5, 12, 169, 8, 12,
		10, 12, 12, 12, 172, 9, 12, 3, 12, 174, 8, 12, 1, 13, 1, 13, 1, 13, 1,
		13, 1, 13, 1, 14, 3, 14, 182, 8, 14, 1, 14, 1, 14, 3, 14, 186, 8, 14, 1,
		14, 1, 14, 3, 14, 190, 8, 14, 1, 14, 1, 14, 3, 14, 194, 8, 14, 1, 15, 1,
		15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15,
		1, 15, 1, 15, 3, 15, 210, 8, 15, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1,
		16, 1, 16, 1, 16, 1, 16, 1, 16, 5, 16, 222, 8, 16, 10, 16, 12, 16, 225,
		9, 16, 1, 16, 3, 16, 228, 8, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1,
		16, 1, 16, 5, 16, 237, 8, 16, 10, 16, 12, 16, 240, 9, 16, 1, 16, 1, 16,
		1, 16, 1, 16, 3, 16, 246, 8, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1,
		16, 1, 16, 5, 16, 255, 8, 16, 10, 16, 12, 16, 258, 9, 16, 1, 16, 1, 16,
		1, 16, 1, 16, 3, 16, 264, 8, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1,
		16, 5, 16, 272, 8, 16, 10, 16, 12, 16, 275, 9, 16, 1, 16, 1, 16, 3, 16,
		279, 8, 16, 3, 16, 281, 8, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16,
		5, 16, 289, 8, 16, 10, 16, 12, 16, 292, 9, 16, 1, 16, 3, 16, 295, 8, 16,
		1, 16, 3, 16, 298, 8, 16, 1, 17, 1, 17, 1, 17, 3, 17, 303, 8, 17, 1, 17,
		1, 17, 1, 17, 1, 17, 1, 17, 3, 17, 310, 8, 17, 1, 18, 1, 18, 1, 19, 1,
		19, 1, 19, 1, 19, 5, 19, 318, 8, 19, 10, 19, 12, 19, 321, 9, 19, 1, 19,
		3, 19, 324, 8, 19, 1, 19, 1, 19, 1, 19, 0, 0, 20, 0, 2, 4, 6, 8, 10, 12,
		14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 0, 3, 2, 0, 10, 10,
		31, 31, 3, 0, 5, 5, 37, 38, 41, 41, 2, 0, 23, 24, 27, 30, 359, 0, 40, 1,
		0, 0, 0, 2, 44, 1, 0, 0, 0, 4, 59, 1, 0, 0, 0, 6, 66, 1, 0, 0, 0, 8, 73,
		1, 0, 0, 0, 10, 92, 1, 0, 0, 0, 12, 95, 1, 0, 0, 0, 14, 110, 1, 0, 0, 0,
		16, 113, 1, 0, 0, 0, 18, 123, 1, 0, 0, 0, 20, 132, 1, 0, 0, 0, 22, 134,
		1, 0, 0, 0, 24, 173, 1, 0, 0, 0, 26, 175, 1, 0, 0, 0, 28, 193, 1, 0, 0,
		0, 30, 209, 1, 0, 0, 0, 32, 297, 1, 0, 0, 0, 34, 309, 1, 0, 0, 0, 36, 311,
		1, 0, 0, 0, 38, 313, 1, 0, 0, 0, 40, 41, 3, 2, 1, 0, 41, 42, 5, 0, 0, 1,
		42, 1, 1, 0, 0, 0, 43, 45, 3, 4, 2, 0, 44, 43, 1, 0, 0, 0, 44, 45, 1, 0,
		0, 0, 45, 49, 1, 0, 0, 0, 46, 48, 3, 6, 3, 0, 47, 46, 1, 0, 0, 0, 48, 51,
		1, 0, 0, 0, 49, 47, 1, 0, 0, 0, 49, 50, 1, 0, 0, 0, 50, 56, 1, 0, 0, 0,
		51, 49, 1, 0, 0, 0, 52, 55, 3, 8, 4, 0, 53, 55, 3, 16, 8, 0, 54, 52, 1,
		0, 0, 0, 54, 53, 1, 0, 0, 0, 55, 58, 1, 0, 0, 0, 56, 54, 1, 0, 0, 0, 56,
		57, 1, 0, 0, 0, 57, 3, 1, 0, 0, 0, 58, 56, 1, 0, 0, 0, 59, 60, 5, 11, 0,
		0, 60, 62, 5, 42, 0, 0, 61, 63, 3, 38, 19, 0, 62, 61, 1, 0, 0, 0, 62, 63,
		1, 0, 0, 0, 63, 64, 1, 0, 0, 0, 64, 65, 5, 26, 0, 0, 65, 5, 1, 0, 0, 0,
		66, 67, 5, 12, 0, 0, 67, 69, 5, 42, 0, 0, 68, 70, 3, 38, 19, 0, 69, 68,
		1, 0, 0, 0, 69, 70, 1, 0, 0, 0, 70, 71, 1, 0, 0, 0, 71, 72, 5, 26, 0, 0,
		72, 7, 1, 0, 0, 0, 73, 74, 5, 13, 0, 0, 74, 76, 3, 36, 18, 0, 75, 77, 5,
		1, 0, 0, 76, 75, 1, 0, 0, 0, 76, 77, 1, 0, 0, 0, 77, 79, 1, 0, 0, 0, 78,
		80, 3, 10, 5, 0, 79, 78, 1, 0, 0, 0, 79, 80, 1, 0, 0, 0, 80, 84, 1, 0,
		0, 0, 81, 83, 3, 12, 6, 0, 82, 81, 1, 0, 0, 0, 83, 86, 1, 0, 0, 0, 84,
		82, 1, 0, 0, 0, 84, 85, 1, 0, 0, 0, 85, 88, 1, 0, 0, 0, 86, 84, 1, 0, 0,
		0, 87, 89, 3, 14, 7, 0, 88, 87, 1, 0, 0, 0, 88, 89, 1, 0, 0, 0, 89, 90,
		1, 0, 0, 0, 90, 91, 5, 2, 0, 0, 91, 9, 1, 0, 0, 0, 92, 93, 5, 3, 0, 0,
		93, 94, 3, 38, 19, 0, 94, 11, 1, 0, 0, 0, 95, 96, 5, 14, 0, 0, 96, 102,
		5, 19, 0, 0, 97, 98, 3, 32, 16, 0, 98, 99, 5, 25, 0, 0, 99, 101, 1, 0,
		0, 0, 100, 97, 1, 0, 0, 0, 101, 104, 1, 0, 0, 0, 102, 100, 1, 0, 0, 0,
		102, 103, 1, 0, 0, 0, 103, 106, 1, 0, 0, 0, 104, 102, 1, 0, 0, 0, 105,
		107, 3, 32, 16, 0, 106, 105, 1, 0, 0, 0, 106, 107, 1, 0, 0, 0, 107, 108,
		1, 0, 0, 0, 108, 109, 5, 20, 0, 0, 109, 13, 1, 0, 0, 0, 110, 111, 5, 4,
		0, 0, 111, 112, 3, 38, 19, 0, 112, 15, 1, 0, 0, 0, 113, 115, 3, 36, 18,
		0, 114, 116, 3, 18, 9, 0, 115, 114, 1, 0, 0, 0, 115, 116, 1, 0, 0, 0, 116,
		119, 1, 0, 0, 0, 117, 118, 7, 0, 0, 0, 118, 120, 3, 22, 11, 0, 119, 117,
		1, 0, 0, 0, 119, 120, 1, 0, 0, 0, 120, 121, 1, 0, 0, 0, 121, 122, 5, 2,
		0, 0, 122, 17, 1, 0, 0, 0, 123, 124, 5, 34, 0, 0, 124, 125, 5, 19, 0, 0,
		125, 128, 3, 20, 10, 0, 126, 127, 5, 25, 0, 0, 127, 129, 3, 20, 10, 0,
		128, 126, 1, 0, 0, 0, 128, 129, 1, 0, 0, 0, 129, 130, 1, 0, 0, 0, 130,
		131, 5, 20, 0, 0, 131, 19, 1, 0, 0, 0, 132, 133, 7, 1, 0, 0, 133, 21, 1,
		0, 0, 0, 134, 139, 3, 28, 14, 0, 135, 136, 5, 25, 0, 0, 136, 138, 3, 28,
		14, 0, 137, 135, 1, 0, 0, 0, 138, 141, 1, 0, 0, 0, 139, 137, 1, 0, 0, 0,
		139, 140, 1, 0, 0, 0, 140, 143, 1, 0, 0, 0, 141, 139, 1, 0, 0, 0, 142,
		144, 5, 25, 0, 0, 143, 142, 1, 0, 0, 0, 143, 144, 1, 0, 0, 0, 144, 149,
		1, 0, 0, 0, 145, 146, 5, 33, 0, 0, 146, 148, 3, 24, 12, 0, 147, 145, 1,
		0, 0, 0, 148, 151, 1, 0, 0, 0, 149, 147, 1, 0, 0, 0, 149, 150, 1, 0, 0,
		0, 150, 23, 1, 0, 0, 0, 151, 149, 1, 0, 0, 0, 152, 153, 5, 16, 0, 0, 153,
		163, 3, 32, 16, 0, 154, 155, 5, 25, 0, 0, 155, 160, 3, 26, 13, 0, 156,
		157, 5, 25, 0, 0, 157, 159, 3, 26, 13, 0, 158, 156, 1, 0, 0, 0, 159, 162,
		1, 0, 0, 0, 160, 158, 1, 0, 0, 0, 160, 161, 1, 0, 0, 0, 161, 164, 1, 0,
		0, 0, 162, 160, 1, 0, 0, 0, 163, 154, 1, 0, 0, 0, 163, 164, 1, 0, 0, 0,
		164, 174, 1, 0, 0, 0, 165, 170, 3, 26, 13, 0, 166, 167, 5, 25, 0, 0, 167,
		169, 3, 26, 13, 0, 168, 166, 1, 0, 0, 0, 169, 172, 1, 0, 0, 0, 170, 168,
		1, 0, 0, 0, 170, 171, 1, 0, 0, 0, 171, 174, 1, 0, 0, 0, 172, 170, 1, 0,
		0, 0, 173, 152, 1, 0, 0, 0, 173, 165, 1, 0, 0, 0, 174, 25, 1, 0, 0, 0,
		175, 176, 5, 15, 0, 0, 176, 177, 5, 41, 0, 0, 177, 178, 5, 23, 0, 0, 178,
		179, 3, 32, 16, 0, 179, 27, 1, 0, 0, 0, 180, 182, 3, 30, 15, 0, 181, 180,
		1, 0, 0, 0, 181, 182, 1, 0, 0, 0, 182, 183, 1, 0, 0, 0, 183, 185, 3, 32,
		16, 0, 184, 186, 3, 18, 9, 0, 185, 184, 1, 0, 0, 0, 185, 186, 1, 0, 0,
		0, 186, 189, 1, 0, 0, 0, 187, 188, 7, 2, 0, 0, 188, 190, 3, 32, 16, 0,
		189, 187, 1, 0, 0, 0, 189, 190, 1, 0, 0, 0, 190, 194, 1, 0, 0, 0, 191,
		192, 5, 26, 0, 0, 192, 194, 3, 32, 16, 0, 193, 181, 1, 0, 0, 0, 193, 191,
		1, 0, 0, 0, 194, 29, 1, 0, 0, 0, 195, 196, 5, 35, 0, 0, 196, 197, 5, 19,
		0, 0, 197, 198, 3, 20, 10, 0, 198, 199, 5, 25, 0, 0, 199, 200, 3, 20, 10,
		0, 200, 201, 5, 20, 0, 0, 201, 210, 1, 0, 0, 0, 202, 203, 5, 36, 0, 0,
		203, 204, 5, 19, 0, 0, 204, 205, 3, 20, 10, 0, 205, 206, 5, 25, 0, 0, 206,
		207, 3, 20, 10, 0, 207, 208, 5, 20, 0, 0, 208, 210, 1, 0, 0, 0, 209, 195,
		1, 0, 0, 0, 209, 202, 1, 0, 0, 0, 210, 31, 1, 0, 0, 0, 211, 298, 5, 41,
		0, 0, 212, 298, 5, 45, 0, 0, 213, 298, 5, 39, 0, 0, 214, 298, 5, 40, 0,
		0, 215, 298, 5, 46, 0, 0, 216, 298, 5, 47, 0, 0, 217, 223, 5, 19, 0, 0,
		218, 219, 3, 32, 16, 0, 219, 220, 5, 25, 0, 0, 220, 222, 1, 0, 0, 0, 221,
		218, 1, 0, 0, 0, 222, 225, 1, 0, 0, 0, 223, 221, 1, 0, 0, 0, 223, 224,
		1, 0, 0, 0, 224, 227, 1, 0, 0, 0, 225, 223, 1, 0, 0, 0, 226, 228, 3, 32,
		16, 0, 227, 226, 1, 0, 0, 0, 227, 228, 1, 0, 0, 0, 228, 229, 1, 0, 0, 0,
		229, 298, 5, 20, 0, 0, 230, 238, 5, 19, 0, 0, 231, 232, 3, 32, 16, 0, 232,
		233, 5, 6, 0, 0, 233, 234, 3, 32, 16, 0, 234, 235, 5, 25, 0, 0, 235, 237,
		1, 0, 0, 0, 236, 231, 1, 0, 0, 0, 237, 240, 1, 0, 0, 0, 238, 236, 1, 0,
		0, 0, 238, 239, 1, 0, 0, 0, 239, 245, 1, 0, 0, 0, 240, 238, 1, 0, 0, 0,
		241, 242, 3, 32, 16, 0, 242, 243, 5, 6, 0, 0, 243, 244, 3, 32, 16, 0, 244,
		246, 1, 0, 0, 0, 245, 241, 1, 0, 0, 0, 245, 246, 1, 0, 0, 0, 246, 247,
		1, 0, 0, 0, 247, 298, 5, 20, 0, 0, 248, 256, 5, 21, 0, 0, 249, 250, 3,
		32, 16, 0, 250, 251, 5, 6, 0, 0, 251, 252, 3, 32, 16, 0, 252, 253, 5, 25,
		0, 0, 253, 255, 1, 0, 0, 0, 254, 249, 1, 0, 0, 0, 255, 258, 1, 0, 0, 0,
		256, 254, 1, 0, 0, 0, 256, 257, 1, 0, 0, 0, 257, 263, 1, 0, 0, 0, 258,
		256, 1, 0, 0, 0, 259, 260, 3, 32, 16, 0, 260, 261, 5, 6, 0, 0, 261, 262,
		3, 32, 16, 0, 262, 264, 1, 0, 0, 0, 263, 259, 1, 0, 0, 0, 263, 264, 1,
		0, 0, 0, 264, 265, 1, 0, 0, 0, 265, 298, 5, 22, 0, 0, 266, 267, 5, 44,
		0, 0, 267, 273, 5, 28, 0, 0, 268, 269, 3, 34, 17, 0, 269, 270, 5, 25, 0,
		0, 270, 272, 1, 0, 0, 0, 271, 268, 1, 0, 0, 0, 272, 275, 1, 0, 0, 0, 273,
		271, 1, 0, 0, 0, 273, 274, 1, 0, 0, 0, 274, 280, 1, 0, 0, 0, 275, 273,
		1, 0, 0, 0, 276, 278, 3, 34, 17, 0, 277, 279, 5, 25, 0, 0, 278, 277, 1,
		0, 0, 0, 278, 279, 1, 0, 0, 0, 279, 281, 1, 0, 0, 0, 280, 276, 1, 0, 0,
		0, 280, 281, 1, 0, 0, 0, 281, 282, 1, 0, 0, 0, 282, 298, 5, 30, 0, 0, 283,
		284, 5, 42, 0, 0, 284, 290, 5, 17, 0, 0, 285, 286, 3, 32, 16, 0, 286, 287,
		5, 25, 0, 0, 287, 289, 1, 0, 0, 0, 288, 285, 1, 0, 0, 0, 289, 292, 1, 0,
		0, 0, 290, 288, 1, 0, 0, 0, 290, 291, 1, 0, 0, 0, 291, 294, 1, 0, 0, 0,
		292, 290, 1, 0, 0, 0, 293, 295, 3, 32, 16, 0, 294, 293, 1, 0, 0, 0, 294,
		295, 1, 0, 0, 0, 295, 296, 1, 0, 0, 0, 296, 298, 5, 18, 0, 0, 297, 211,
		1, 0, 0, 0, 297, 212, 1, 0, 0, 0, 297, 213, 1, 0, 0, 0, 297, 214, 1, 0,
		0, 0, 297, 215, 1, 0, 0, 0, 297, 216, 1, 0, 0, 0, 297, 217, 1, 0, 0, 0,
		297, 230, 1, 0, 0, 0, 297, 248, 1, 0, 0, 0, 297, 266, 1, 0, 0, 0, 297,
		283, 1, 0, 0, 0, 298, 33, 1, 0, 0, 0, 299, 302, 3, 32, 16, 0, 300, 301,
		5, 6, 0, 0, 301, 303, 3, 32, 16, 0, 302, 300, 1, 0, 0, 0, 302, 303, 1,
		0, 0, 0, 303, 310, 1, 0, 0, 0, 304, 305, 5, 7, 0, 0, 305, 306, 3, 32, 16,
		0, 306, 307, 5, 6, 0, 0, 307, 308, 3, 32, 16, 0, 308, 310, 1, 0, 0, 0,
		309, 299, 1, 0, 0, 0, 309, 304, 1, 0, 0, 0, 310, 35, 1, 0, 0, 0, 311, 312,
		3, 32, 16, 0, 312, 37, 1, 0, 0, 0, 313, 319, 5, 19, 0, 0, 314, 315, 3,
		36, 18, 0, 315, 316, 5, 25, 0, 0, 316, 318, 1, 0, 0, 0, 317, 314, 1, 0,
		0, 0, 318, 321, 1, 0, 0, 0, 319, 317, 1, 0, 0, 0, 319, 320, 1, 0, 0, 0,
		320, 323, 1, 0, 0, 0, 321, 319, 1, 0, 0, 0, 322, 324, 3, 36, 18, 0, 323,
		322, 1, 0, 0, 0, 323, 324, 1, 0, 0, 0, 324, 325, 1, 0, 0, 0, 325, 326,
		5, 20, 0, 0, 326, 39, 1, 0, 0, 0, 43, 44, 49, 54, 56, 62, 69, 76, 79, 84,
		88, 102, 106, 115, 119, 128, 139, 143, 149, 160, 163, 170, 173, 181, 185,
		189, 193, 209, 223, 227, 238, 245, 256, 263, 273, 278, 280, 290, 294, 297,
		302, 309, 319, 323,
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
	MangleParserLBRACE              = 21
	MangleParserRBRACE              = 22
	MangleParserEQ                  = 23
	MangleParserBANGEQ              = 24
	MangleParserCOMMA               = 25
	MangleParserBANG                = 26
	MangleParserLESSEQ              = 27
	MangleParserLESS                = 28
	MangleParserGREATEREQ           = 29
	MangleParserGREATER             = 30
	MangleParserCOLONDASH           = 31
	MangleParserNEWLINE             = 32
	MangleParserPIPEGREATER         = 33
	MangleParserAT                  = 34
	MangleParserDIAMONDMINUS        = 35
	MangleParserBOXMINUS            = 36
	MangleParserTIMESTAMP           = 37
	MangleParserDURATION            = 38
	MangleParserNUMBER              = 39
	MangleParserFLOAT               = 40
	MangleParserVARIABLE            = 41
	MangleParserNAME                = 42
	MangleParserTYPENAME            = 43
	MangleParserDOT_TYPE            = 44
	MangleParserCONSTANT            = 45
	MangleParserSTRING              = 46
	MangleParserBYTESTRING          = 47
)

// MangleParser rules.
const (
	MangleParserRULE_start              = 0
	MangleParserRULE_program            = 1
	MangleParserRULE_packageDecl        = 2
	MangleParserRULE_useDecl            = 3
	MangleParserRULE_decl               = 4
	MangleParserRULE_descrBlock         = 5
	MangleParserRULE_boundsBlock        = 6
	MangleParserRULE_constraintsBlock   = 7
	MangleParserRULE_clause             = 8
	MangleParserRULE_temporalAnnotation = 9
	MangleParserRULE_temporalBound      = 10
	MangleParserRULE_clauseBody         = 11
	MangleParserRULE_transform          = 12
	MangleParserRULE_letStmt            = 13
	MangleParserRULE_literalOrFml       = 14
	MangleParserRULE_temporalOperator   = 15
	MangleParserRULE_term               = 16
	MangleParserRULE_member             = 17
	MangleParserRULE_atom               = 18
	MangleParserRULE_atoms              = 19
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
		p.SetState(40)
		p.Program()
	}
	{
		p.SetState(41)
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
	p.SetState(44)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MangleParserPACKAGE {
		{
			p.SetState(43)
			p.PackageDecl()
		}

	}
	p.SetState(49)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == MangleParserUSE {
		{
			p.SetState(46)
			p.UseDecl()
		}

		p.SetState(51)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(56)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&272129130504192) != 0 {
		p.SetState(54)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case MangleParserDECL:
			{
				p.SetState(52)
				p.Decl()
			}

		case MangleParserLBRACKET, MangleParserLBRACE, MangleParserNUMBER, MangleParserFLOAT, MangleParserVARIABLE, MangleParserNAME, MangleParserDOT_TYPE, MangleParserCONSTANT, MangleParserSTRING, MangleParserBYTESTRING:
			{
				p.SetState(53)
				p.Clause()
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(58)
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
		p.SetState(59)
		p.Match(MangleParserPACKAGE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(60)
		p.Match(MangleParserNAME)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(62)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MangleParserLBRACKET {
		{
			p.SetState(61)
			p.Atoms()
		}

	}
	{
		p.SetState(64)
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
		p.SetState(66)
		p.Match(MangleParserUSE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(67)
		p.Match(MangleParserNAME)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(69)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MangleParserLBRACKET {
		{
			p.SetState(68)
			p.Atoms()
		}

	}
	{
		p.SetState(71)
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
		p.SetState(73)
		p.Match(MangleParserDECL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(74)
		p.Atom()
	}
	p.SetState(76)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MangleParserT__0 {
		{
			p.SetState(75)
			p.Match(MangleParserT__0)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

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
			p.DescrBlock()
		}

	}
	p.SetState(84)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == MangleParserBOUND {
		{
			p.SetState(81)
			p.BoundsBlock()
		}

		p.SetState(86)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(88)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MangleParserT__3 {
		{
			p.SetState(87)
			p.ConstraintsBlock()
		}

	}
	{
		p.SetState(90)
		p.Match(MangleParserT__1)
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
		p.SetState(92)
		p.Match(MangleParserT__2)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(93)
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
		p.SetState(95)
		p.Match(MangleParserBOUND)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(96)
		p.Match(MangleParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(102)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 10, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(97)
				p.Term()
			}
			{
				p.SetState(98)
				p.Match(MangleParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		p.SetState(104)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 10, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	p.SetState(106)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&272129130496000) != 0 {
		{
			p.SetState(105)
			p.Term()
		}

	}
	{
		p.SetState(108)
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
		p.SetState(110)
		p.Match(MangleParserT__3)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(111)
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
	TemporalAnnotation() ITemporalAnnotationContext
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

func (s *ClauseContext) TemporalAnnotation() ITemporalAnnotationContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITemporalAnnotationContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITemporalAnnotationContext)
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
		p.SetState(113)
		p.Atom()
	}
	p.SetState(115)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MangleParserAT {
		{
			p.SetState(114)
			p.TemporalAnnotation()
		}

	}
	p.SetState(119)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MangleParserLONGLEFTDOUBLEARROW || _la == MangleParserCOLONDASH {
		{
			p.SetState(117)
			_la = p.GetTokenStream().LA(1)

			if !(_la == MangleParserLONGLEFTDOUBLEARROW || _la == MangleParserCOLONDASH) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(118)
			p.ClauseBody()
		}

	}
	{
		p.SetState(121)
		p.Match(MangleParserT__1)
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

// ITemporalAnnotationContext is an interface to support dynamic dispatch.
type ITemporalAnnotationContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AT() antlr.TerminalNode
	LBRACKET() antlr.TerminalNode
	AllTemporalBound() []ITemporalBoundContext
	TemporalBound(i int) ITemporalBoundContext
	RBRACKET() antlr.TerminalNode
	COMMA() antlr.TerminalNode

	// IsTemporalAnnotationContext differentiates from other interfaces.
	IsTemporalAnnotationContext()
}

type TemporalAnnotationContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTemporalAnnotationContext() *TemporalAnnotationContext {
	var p = new(TemporalAnnotationContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_temporalAnnotation
	return p
}

func InitEmptyTemporalAnnotationContext(p *TemporalAnnotationContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_temporalAnnotation
}

func (*TemporalAnnotationContext) IsTemporalAnnotationContext() {}

func NewTemporalAnnotationContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TemporalAnnotationContext {
	var p = new(TemporalAnnotationContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_temporalAnnotation

	return p
}

func (s *TemporalAnnotationContext) GetParser() antlr.Parser { return s.parser }

func (s *TemporalAnnotationContext) AT() antlr.TerminalNode {
	return s.GetToken(MangleParserAT, 0)
}

func (s *TemporalAnnotationContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(MangleParserLBRACKET, 0)
}

func (s *TemporalAnnotationContext) AllTemporalBound() []ITemporalBoundContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITemporalBoundContext); ok {
			len++
		}
	}

	tst := make([]ITemporalBoundContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITemporalBoundContext); ok {
			tst[i] = t.(ITemporalBoundContext)
			i++
		}
	}

	return tst
}

func (s *TemporalAnnotationContext) TemporalBound(i int) ITemporalBoundContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITemporalBoundContext); ok {
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

	return t.(ITemporalBoundContext)
}

func (s *TemporalAnnotationContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(MangleParserRBRACKET, 0)
}

func (s *TemporalAnnotationContext) COMMA() antlr.TerminalNode {
	return s.GetToken(MangleParserCOMMA, 0)
}

func (s *TemporalAnnotationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TemporalAnnotationContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TemporalAnnotationContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterTemporalAnnotation(s)
	}
}

func (s *TemporalAnnotationContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitTemporalAnnotation(s)
	}
}

func (s *TemporalAnnotationContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitTemporalAnnotation(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) TemporalAnnotation() (localctx ITemporalAnnotationContext) {
	localctx = NewTemporalAnnotationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, MangleParserRULE_temporalAnnotation)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(123)
		p.Match(MangleParserAT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(124)
		p.Match(MangleParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(125)
		p.TemporalBound()
	}
	p.SetState(128)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MangleParserCOMMA {
		{
			p.SetState(126)
			p.Match(MangleParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(127)
			p.TemporalBound()
		}

	}
	{
		p.SetState(130)
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

// ITemporalBoundContext is an interface to support dynamic dispatch.
type ITemporalBoundContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TIMESTAMP() antlr.TerminalNode
	DURATION() antlr.TerminalNode
	VARIABLE() antlr.TerminalNode

	// IsTemporalBoundContext differentiates from other interfaces.
	IsTemporalBoundContext()
}

type TemporalBoundContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTemporalBoundContext() *TemporalBoundContext {
	var p = new(TemporalBoundContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_temporalBound
	return p
}

func InitEmptyTemporalBoundContext(p *TemporalBoundContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_temporalBound
}

func (*TemporalBoundContext) IsTemporalBoundContext() {}

func NewTemporalBoundContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TemporalBoundContext {
	var p = new(TemporalBoundContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_temporalBound

	return p
}

func (s *TemporalBoundContext) GetParser() antlr.Parser { return s.parser }

func (s *TemporalBoundContext) TIMESTAMP() antlr.TerminalNode {
	return s.GetToken(MangleParserTIMESTAMP, 0)
}

func (s *TemporalBoundContext) DURATION() antlr.TerminalNode {
	return s.GetToken(MangleParserDURATION, 0)
}

func (s *TemporalBoundContext) VARIABLE() antlr.TerminalNode {
	return s.GetToken(MangleParserVARIABLE, 0)
}

func (s *TemporalBoundContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TemporalBoundContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TemporalBoundContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterTemporalBound(s)
	}
}

func (s *TemporalBoundContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitTemporalBound(s)
	}
}

func (s *TemporalBoundContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitTemporalBound(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) TemporalBound() (localctx ITemporalBoundContext) {
	localctx = NewTemporalBoundContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, MangleParserRULE_temporalBound)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(132)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2611340116000) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
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
	p.EnterRule(localctx, 22, MangleParserRULE_clauseBody)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(134)
		p.LiteralOrFml()
	}
	p.SetState(139)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 15, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(135)
				p.Match(MangleParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(136)
				p.LiteralOrFml()
			}

		}
		p.SetState(141)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 15, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	p.SetState(143)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MangleParserCOMMA {
		{
			p.SetState(142)
			p.Match(MangleParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	p.SetState(149)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == MangleParserPIPEGREATER {
		{
			p.SetState(145)
			p.Match(MangleParserPIPEGREATER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(146)
			p.Transform()
		}

		p.SetState(151)
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
	p.EnterRule(localctx, 24, MangleParserRULE_transform)
	var _la int

	p.SetState(173)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MangleParserDO:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(152)
			p.Match(MangleParserDO)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(153)
			p.Term()
		}
		p.SetState(163)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == MangleParserCOMMA {
			{
				p.SetState(154)
				p.Match(MangleParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(155)
				p.LetStmt()
			}
			p.SetState(160)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == MangleParserCOMMA {
				{
					p.SetState(156)
					p.Match(MangleParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(157)
					p.LetStmt()
				}

				p.SetState(162)
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
			p.SetState(165)
			p.LetStmt()
		}
		p.SetState(170)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == MangleParserCOMMA {
			{
				p.SetState(166)
				p.Match(MangleParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(167)
				p.LetStmt()
			}

			p.SetState(172)
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
	p.EnterRule(localctx, 26, MangleParserRULE_letStmt)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(175)
		p.Match(MangleParserLET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(176)
		p.Match(MangleParserVARIABLE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(177)
		p.Match(MangleParserEQ)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(178)
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
	TemporalOperator() ITemporalOperatorContext
	TemporalAnnotation() ITemporalAnnotationContext
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

func (s *LiteralOrFmlContext) TemporalOperator() ITemporalOperatorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITemporalOperatorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITemporalOperatorContext)
}

func (s *LiteralOrFmlContext) TemporalAnnotation() ITemporalAnnotationContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITemporalAnnotationContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITemporalAnnotationContext)
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
	p.EnterRule(localctx, 28, MangleParserRULE_literalOrFml)
	var _la int

	p.SetState(193)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MangleParserLBRACKET, MangleParserLBRACE, MangleParserDIAMONDMINUS, MangleParserBOXMINUS, MangleParserNUMBER, MangleParserFLOAT, MangleParserVARIABLE, MangleParserNAME, MangleParserDOT_TYPE, MangleParserCONSTANT, MangleParserSTRING, MangleParserBYTESTRING:
		p.EnterOuterAlt(localctx, 1)
		p.SetState(181)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == MangleParserDIAMONDMINUS || _la == MangleParserBOXMINUS {
			{
				p.SetState(180)
				p.TemporalOperator()
			}

		}
		{
			p.SetState(183)
			p.Term()
		}
		p.SetState(185)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == MangleParserAT {
			{
				p.SetState(184)
				p.TemporalAnnotation()
			}

		}
		p.SetState(189)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2038431744) != 0 {
			{
				p.SetState(187)
				_la = p.GetTokenStream().LA(1)

				if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2038431744) != 0) {
					p.GetErrorHandler().RecoverInline(p)
				} else {
					p.GetErrorHandler().ReportMatch(p)
					p.Consume()
				}
			}
			{
				p.SetState(188)
				p.Term()
			}

		}

	case MangleParserBANG:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(191)
			p.Match(MangleParserBANG)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(192)
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

// ITemporalOperatorContext is an interface to support dynamic dispatch.
type ITemporalOperatorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	DIAMONDMINUS() antlr.TerminalNode
	LBRACKET() antlr.TerminalNode
	AllTemporalBound() []ITemporalBoundContext
	TemporalBound(i int) ITemporalBoundContext
	COMMA() antlr.TerminalNode
	RBRACKET() antlr.TerminalNode
	BOXMINUS() antlr.TerminalNode

	// IsTemporalOperatorContext differentiates from other interfaces.
	IsTemporalOperatorContext()
}

type TemporalOperatorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTemporalOperatorContext() *TemporalOperatorContext {
	var p = new(TemporalOperatorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_temporalOperator
	return p
}

func InitEmptyTemporalOperatorContext(p *TemporalOperatorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MangleParserRULE_temporalOperator
}

func (*TemporalOperatorContext) IsTemporalOperatorContext() {}

func NewTemporalOperatorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TemporalOperatorContext {
	var p = new(TemporalOperatorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MangleParserRULE_temporalOperator

	return p
}

func (s *TemporalOperatorContext) GetParser() antlr.Parser { return s.parser }

func (s *TemporalOperatorContext) DIAMONDMINUS() antlr.TerminalNode {
	return s.GetToken(MangleParserDIAMONDMINUS, 0)
}

func (s *TemporalOperatorContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(MangleParserLBRACKET, 0)
}

func (s *TemporalOperatorContext) AllTemporalBound() []ITemporalBoundContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITemporalBoundContext); ok {
			len++
		}
	}

	tst := make([]ITemporalBoundContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITemporalBoundContext); ok {
			tst[i] = t.(ITemporalBoundContext)
			i++
		}
	}

	return tst
}

func (s *TemporalOperatorContext) TemporalBound(i int) ITemporalBoundContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITemporalBoundContext); ok {
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

	return t.(ITemporalBoundContext)
}

func (s *TemporalOperatorContext) COMMA() antlr.TerminalNode {
	return s.GetToken(MangleParserCOMMA, 0)
}

func (s *TemporalOperatorContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(MangleParserRBRACKET, 0)
}

func (s *TemporalOperatorContext) BOXMINUS() antlr.TerminalNode {
	return s.GetToken(MangleParserBOXMINUS, 0)
}

func (s *TemporalOperatorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TemporalOperatorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TemporalOperatorContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.EnterTemporalOperator(s)
	}
}

func (s *TemporalOperatorContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MangleListener); ok {
		listenerT.ExitTemporalOperator(s)
	}
}

func (s *TemporalOperatorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MangleVisitor:
		return t.VisitTemporalOperator(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MangleParser) TemporalOperator() (localctx ITemporalOperatorContext) {
	localctx = NewTemporalOperatorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, MangleParserRULE_temporalOperator)
	p.SetState(209)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MangleParserDIAMONDMINUS:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(195)
			p.Match(MangleParserDIAMONDMINUS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(196)
			p.Match(MangleParserLBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(197)
			p.TemporalBound()
		}
		{
			p.SetState(198)
			p.Match(MangleParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(199)
			p.TemporalBound()
		}
		{
			p.SetState(200)
			p.Match(MangleParserRBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case MangleParserBOXMINUS:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(202)
			p.Match(MangleParserBOXMINUS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(203)
			p.Match(MangleParserLBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(204)
			p.TemporalBound()
		}
		{
			p.SetState(205)
			p.Match(MangleParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(206)
			p.TemporalBound()
		}
		{
			p.SetState(207)
			p.Match(MangleParserRBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
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

func (s *StructContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(MangleParserLBRACE, 0)
}

func (s *StructContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(MangleParserRBRACE, 0)
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
	p.EnterRule(localctx, 32, MangleParserRULE_term)
	var _la int

	var _alt int

	p.SetState(297)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 38, p.GetParserRuleContext()) {
	case 1:
		localctx = NewVarContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(211)
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
			p.SetState(212)
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
			p.SetState(213)
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
			p.SetState(214)
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
			p.SetState(215)
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
			p.SetState(216)
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
			p.SetState(217)
			p.Match(MangleParserLBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(223)
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
					p.SetState(218)
					p.Term()
				}
				{
					p.SetState(219)
					p.Match(MangleParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(225)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 27, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(227)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&272129130496000) != 0 {
			{
				p.SetState(226)
				p.Term()
			}

		}
		{
			p.SetState(229)
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
			p.SetState(230)
			p.Match(MangleParserLBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(238)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 29, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(231)
					p.Term()
				}
				{
					p.SetState(232)
					p.Match(MangleParserT__5)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(233)
					p.Term()
				}
				{
					p.SetState(234)
					p.Match(MangleParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(240)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 29, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(245)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&272129130496000) != 0 {
			{
				p.SetState(241)
				p.Term()
			}
			{
				p.SetState(242)
				p.Match(MangleParserT__5)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(243)
				p.Term()
			}

		}
		{
			p.SetState(247)
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
			p.SetState(248)
			p.Match(MangleParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(256)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 31, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(249)
					p.Term()
				}
				{
					p.SetState(250)
					p.Match(MangleParserT__5)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(251)
					p.Term()
				}
				{
					p.SetState(252)
					p.Match(MangleParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(258)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 31, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(263)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&272129130496000) != 0 {
			{
				p.SetState(259)
				p.Term()
			}
			{
				p.SetState(260)
				p.Match(MangleParserT__5)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(261)
				p.Term()
			}

		}
		{
			p.SetState(265)
			p.Match(MangleParserRBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 10:
		localctx = NewDotTypeContext(p, localctx)
		p.EnterOuterAlt(localctx, 10)
		{
			p.SetState(266)
			p.Match(MangleParserDOT_TYPE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(267)
			p.Match(MangleParserLESS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(273)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 33, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(268)
					p.Member()
				}
				{
					p.SetState(269)
					p.Match(MangleParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(275)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 33, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(280)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&272129130496128) != 0 {
			{
				p.SetState(276)
				p.Member()
			}
			p.SetState(278)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			if _la == MangleParserCOMMA {
				{
					p.SetState(277)
					p.Match(MangleParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}

		}
		{
			p.SetState(282)
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
			p.SetState(283)
			p.Match(MangleParserNAME)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(284)
			p.Match(MangleParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(290)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 36, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(285)
					p.Term()
				}
				{
					p.SetState(286)
					p.Match(MangleParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(292)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 36, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(294)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&272129130496000) != 0 {
			{
				p.SetState(293)
				p.Term()
			}

		}
		{
			p.SetState(296)
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
	p.EnterRule(localctx, 34, MangleParserRULE_member)
	var _la int

	p.SetState(309)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MangleParserLBRACKET, MangleParserLBRACE, MangleParserNUMBER, MangleParserFLOAT, MangleParserVARIABLE, MangleParserNAME, MangleParserDOT_TYPE, MangleParserCONSTANT, MangleParserSTRING, MangleParserBYTESTRING:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(299)
			p.Term()
		}
		p.SetState(302)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == MangleParserT__5 {
			{
				p.SetState(300)
				p.Match(MangleParserT__5)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(301)
				p.Term()
			}

		}

	case MangleParserT__6:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(304)
			p.Match(MangleParserT__6)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(305)
			p.Term()
		}
		{
			p.SetState(306)
			p.Match(MangleParserT__5)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(307)
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
	p.EnterRule(localctx, 36, MangleParserRULE_atom)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(311)
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
	p.EnterRule(localctx, 38, MangleParserRULE_atoms)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(313)
		p.Match(MangleParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(319)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 41, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(314)
				p.Atom()
			}
			{
				p.SetState(315)
				p.Match(MangleParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		p.SetState(321)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 41, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	p.SetState(323)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&272129130496000) != 0 {
		{
			p.SetState(322)
			p.Atom()
		}

	}
	{
		p.SetState(325)
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
