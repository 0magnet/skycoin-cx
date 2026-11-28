package ast

import "fmt"

// serializeFailure is one way a serialized program fails to load. A decoder
// records it with its operands and, for the failures whose format starts with
// %s, a serializeName, and the text is built only when it is printed (#412).
type serializeFailure uint8

const (
	serializeFailureRange serializeFailure = iota
	serializeFailureUint32
	serializeFailureUint8
	serializeFailureBool
	serializeFailurePointer
	serializeFailureLength
	serializeFailureSectionLeftOver
	serializeFailureTrailing
	serializeFailureMngCounts
	serializeFailureMngFreeSlot
	serializeFailureMngLive
	serializeFailureCallStack
	serializeFailureCallFunction
	serializeFailureStringRepeat
	serializeFailureStringGid
	serializeFailureLibraryFunction
	serializeFailureEscape
	serializeFailureExpressionFunction
	serializeFailureMemorySize
	serializeFailureMemoryFit
	serializeFailureLibraries
	serializeFailureOpcodeSet
	serializeFailureOutputMode
	serializeFailureHeaderMagic
	serializeFailureHeaderFormat
	serializeFailureHeaderPointerSize
	serializeFailureHeaderLanguageVersion
)

var serializeFailureFormats = [...]string{
	serializeFailureRange:                 "%s %d is outside %d to %d",
	serializeFailureUint32:                "value %d does not fit 32 bits",
	serializeFailureUint8:                 "value %d does not fit 8 bits",
	serializeFailureBool:                  "boolean byte %d",
	serializeFailurePointer:               "pointer %d does not fit %d bytes",
	serializeFailureLength:                "length %d with %d bytes left",
	serializeFailureSectionLeftOver:       "%s section has %d bytes left over",
	serializeFailureTrailing:              "%d bytes after the last section",
	serializeFailureMngCounts:             "%s has Cap %d, Next %d, Len %d, %d free and %d retired slots",
	serializeFailureMngFreeSlot:           "%s free slot %d is not below Next %d",
	serializeFailureMngLive:               "%s has %d live slots, Len %d",
	serializeFailureCallStack:             "call stack of %d entries with %d in use",
	serializeFailureCallFunction:          "call %d names function %#x, which is not live",
	serializeFailureStringRepeat:          "string slot %d repeats an earlier string or is empty",
	serializeFailureStringGid:             "string Gid %#x",
	serializeFailureLibraryFunction:       "function %d calls library %d function %d, which does not exist",
	serializeFailureEscape:                "escape %d",
	serializeFailureExpressionFunction:    "expression %d names function %#x or %#x, which is not live",
	serializeFailureMemorySize:            "memory of %d bytes for a stack of %d, data of %d and a heap of %d",
	serializeFailureMemoryFit:             "%d memory bytes at %d do not fit %d",
	serializeFailureLibraries:             "%d system libraries, this cx has %d",
	serializeFailureOpcodeSet:             "an opcode set of %d words, this cx has %d",
	serializeFailureOutputMode:            "output mode %d, this cx has %d",
	serializeFailureHeaderMagic:           "not a serialized CX program",
	serializeFailureHeaderFormat:          "serialized program has format %d, this cx reads format %d, recompile it from source",
	serializeFailureHeaderPointerSize:     "serialized program was compiled with %d byte pointers, this cx uses %d byte pointers, load it with a cx built to match (the ptr64 build tag gives 8)",
	serializeFailureHeaderLanguageVersion: "serialized program is language version %d, this CX implements version %d, recompile it from source",
}

// serializeName names a section, a manager or a value in the text of a
// serializeFailure.
type serializeName uint8

const (
	serializeNameNone serializeName = iota
	serializeNameInteger
	serializeNameStructReference
	serializeNameSectionProgram
	serializeNameSectionStrings
	serializeNameSectionPackages
	serializeNameSectionStructs
	serializeNameSectionFunctions
	serializeNameSectionExpressions
	serializeNameSectionArguments
	serializeNameSectionTypeSignatures
	serializeNameSectionMemory
	serializeNameCXFunctionMng
	serializeNameCXLineMng
	serializeNameCXArgumentMng
	serializeNameCXArgumentDebugMng
	serializeNameCXTypeSignatureMng
	serializeNameCXTypeSignatureArrayMng
	serializeNameCXTypeSignatureStructMng
)

var serializeNameTexts = [...]string{
	serializeNameNone:                     "",
	serializeNameInteger:                  "integer",
	serializeNameStructReference:          "struct reference",
	serializeNameSectionProgram:           "program",
	serializeNameSectionStrings:           "strings",
	serializeNameSectionPackages:          "packages",
	serializeNameSectionStructs:           "structs",
	serializeNameSectionFunctions:         "functions",
	serializeNameSectionExpressions:       "expressions",
	serializeNameSectionArguments:         "arguments",
	serializeNameSectionTypeSignatures:    "type signatures",
	serializeNameSectionMemory:            "memory",
	serializeNameCXFunctionMng:            "CXFunctionMng",
	serializeNameCXLineMng:                "CXLineMng",
	serializeNameCXArgumentMng:            "CXArgumentMng",
	serializeNameCXArgumentDebugMng:       "CXArgumentDebugMng",
	serializeNameCXTypeSignatureMng:       "CXTypeSignatureMng",
	serializeNameCXTypeSignatureArrayMng:  "CXTypeSignature_ArrayMng",
	serializeNameCXTypeSignatureStructMng: "CXTypeSignature_StructMng",
}

// serializeMalformedError is a malformed serialized program, a failure with
// its name and operands. Error builds the text only when it is printed.
type serializeMalformedError struct {
	serializeMalformedErrorFailure  serializeFailure
	serializeMalformedErrorName     serializeName
	serializeMalformedErrorOperands []interface{}
}

func (malformed serializeMalformedError) Error() string {
	return "serialized program is malformed: " + serializeFailureText(malformed.serializeMalformedErrorFailure, malformed.serializeMalformedErrorName, malformed.serializeMalformedErrorOperands)
}

// serializeHeaderError is a header this cx cannot load.
type serializeHeaderError struct {
	serializeHeaderErrorFailure  serializeFailure
	serializeHeaderErrorOperands []interface{}
}

func (header serializeHeaderError) Error() string {
	return serializeFailureText(header.serializeHeaderErrorFailure, serializeNameNone, header.serializeHeaderErrorOperands)
}

// serializeSectionError is the failure err inside a section.
type serializeSectionError struct {
	serializeSectionErrorName  serializeName
	serializeSectionErrorCause error
}

func (section serializeSectionError) Error() string {
	return serializeNameTexts[section.serializeSectionErrorName] + " section: " + section.serializeSectionErrorCause.Error()
}

func (section serializeSectionError) Unwrap() error {
	return section.serializeSectionErrorCause
}

func serializeFailureText(failure serializeFailure, name serializeName, operands []interface{}) string {
	if name != serializeNameNone {
		operands = append([]interface{}{serializeNameTexts[name]}, operands...)
	}
	return fmt.Sprintf(serializeFailureFormats[failure], operands...)
}

var serializeMagicBytes = [len(serializeMagic)]byte{'C', 'X', 'P', 'R', 'O', 'G', 'R', 'M'}
