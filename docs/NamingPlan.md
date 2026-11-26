# Naming plan

This plan renames the Go names that break the Naming section of
[AGENTS.md](../AGENTS.md), and lists the names that stay with their reason.
It is tracked in [#569](https://github.com/skycoin/skycoin-cx/issues/569).
`make check-names` enforces the rules below, and `names-allowlist.txt` lists
every finding that is left, so it only shrinks as each phase lands.

## Rules

A name reads from the general to the specific, in Java type name style, so
related names share a prefix and sort together: `CXFunctionMng`,
`CXFunctionGet`, `LibraryFunctionDefinitionTable`, `OpcodeGroupFile`. The
check reports a Go name outside tests when it

1. has a vague word, `Data`, `Info`, `Tmp`, `Temp`, `Item`, `Thing`, `Stuff`,
   `Misc`, `Do`, `Handle`, `Handler`, `Process`, `Manager`, `Helper` or
   `Util`, as any word of the name, for every kind of name;
2. starts with `is` or `has`, for every kind of name;
3. is a single letter at package level;
4. has an underscore at package level, where Java type name style has none;
5. is exported and starts with its own package name, as `gc.GCState` does;
6. is an exported function or method that starts with a verb, as
   `GetCXAtomicOp` does. The subject leads and the verb ends the name,
   `CXAtomicOpGet`. Methods that a standard interface calls keep their name,
   `String`, `Error`, `Unwrap`, `Is`, `As`, `Read`, `Write`, `Close`, `Seek`,
   `Format`, `Len`, `Less`, `Swap`, `Lex`, `Reset` and the `Marshal` and
   `Unmarshal` methods. A constructor is `<Type>New`, not `New<Type>`;
7. has an abbreviation that is not the usual name at package level, such as
   `prgrm`, `expr`, `fn`, `glbl`, `inp`, `idx`, `sym`, `decl`, `obj`,
   `sig` or `typ`. The CX type and opcode words (`str`, `ptr`, `i32`, `jmp`,
   `eq`) and the usual compiler words `arg`, `instr` and `spec` are allowed.

Locals, parameters and receivers are checked by rules 1 and 2 only, so a
receiver or a loop index keeps its short name.

Nothing that leaves the process is renamed: CX source names, opcode and
library function names, CX constants, serialized fields, flags and JSON.

## What stays

| Reason | Names |
|---|---|
| `handle` is the i32 handle of the file, net, cipher and cxfx libraries | 70 |
| The object model manager, docs/ObjectModel.md, and the `manager` receivers that match the generated `_mng.go` files | 18 |
| The data segment that `mem("data")` selects, beside the stack and the heap | 12 |
| GL, AL and gltext wrappers named after the library function they wrap, and the `cxgl` GL constants spelled as in GL | 21 |
| The sys/process library, and `NetTriggerData` for the CX constant `net.TriggerData` | 4 |
| `CONST_<library>_<NAME>`, the index of a CX constant, spelled as the constant | 252 |
| `CXTypeSignature_Array` and `CXTypeSignature_Struct`, object model types whose name the generator derives the manager names from | 4 |
| The copy of Go strconv in t0-cx-core-floatparse | 3 |
| `x` and `y` coordinates and the `R`, `G`, `B` and `E` channels of RGBE | 8 |

## Pull requests

Each step is one pull request, or a few for a large package, merged one at a
time with main CI green in between. The names below are the proposal, and a
pull request may refine one where the mechanical form reads badly, saying so.

1. `make check-names` with all rules, the full allowlist and this plan.
2. `t0-util` is split by subject. The sandbox paths and file errors go to
   `t1-cx-runtime-sandbox` (package `sandbox`), the source arguments to
   `t0-cx-compiler-sourceargs` (`sourceargs.Parse`), the logged file calls to
   `t0-cx-runtime-filelog` (`filelog.Open`, `Create`, `Stat`, `Enable`), and
   `ProgramTextStringView` to `stringpool.StringView`.
3. Vague words and `is` or `has`, rules 1 and 2: t0 to t2, then t3-ast, then
   t4-actions, then the rest (253 names).
4. Verbs first, rule 6: t3-ast in two parts, then t4-actions, then the rest
   (252).
5. Package name repeated, rule 5: gc, core, printer, assert, cxfx, cipher and
   repl (67).
6. Underscores, rule 4: t0-types, then t1-constants, then t3-ast, then the
   rest (357).
7. Abbreviations and single letters, rules 3 and 7 (36).
8. Package directories, see below (46).

Exported renames get a CHANGELOG line in their pull request.


## Package directories

Every package directory reads `t<N>-cx-<layer>-<subject>`, and its Go
package is the subject, or `main` for a command. The tier N does not
change. The layers are

- `core`, the language model that the compiler and the runtime share;
- `compiler`, from source to the IR;
- `runtime`, running a program;
- `library`, the system libraries;
- `tools`, the developer commands and test harnesses;
- `command`, the binaries that ship, cx, cx-core and mobile.

`make check-names` reports a directory outside this form, or a package
clause that is not its subject. These renames are step 8, a few
directories per pull request. They change import paths, so
skycoin-cx-playground and skycoin-cx-playground-android update their
imports at their next bump of skycoin-cx. skycoin-cx-forth imports none.

| Old | New | Package | Holds |
|---|---|---|---|
| `t0-benchmarks` | `t0-cx-tools-benchmarks` | `benchmarks` | the Go benchmarks and the language comparison |
| `t0-benchmarks-structser` | `t0-cx-tools-benchmarkserialize` | `main` | the struct serialization benchmark |
| `t0-cxfiles` | `t0-cx-library-filetable` | `filetable` | the open files of the file library by i32 handle |
| `t0-cxhash` | `t0-cx-core-digest` | `digest` | the BLAKE3-256 digest of CX |
| `t0-cxtestrunner` | `t0-cx-tools-testrunner` | `testrunner` | the runner of the CX program tests |
| `t0-ifacecheck` | `t0-cx-tools-interfacecheck` | `main` | the interface embedding check |
| `t0-lz4codec` | `t0-cx-core-compression` | `compression` | the LZ4 block codec of serialized programs |
| `t0-packages` | `t0-cx-core-syspackages` | `syspackages` | the built-in sys/ packages |
| `t0-plbolt` | `t0-cx-tools-packagestore` | `packagestore` | the bolt store of the package loader |
| `t0-textsections` | `t0-cx-core-textsections` | `textsections` | the CX text form of stored files |
| `t0-types` | `t0-cx-core-valuetypes` | `valuetypes` | the CX value types, their sizes and memory form |
| `t1-cipherstate` | `t1-cx-library-cipherstate` | `cipherstate` | the hash states of the cipher library |
| `t1-constants` | `t1-cx-core-opcodetable` | `opcodetable` | the opcode and library tables and their constants |
| `t1-cxtest` | `t1-cx-tools-programtest` | `main` | the command that runs the CX program tests |
| `t1-netbatch` | `t1-cx-library-networkbatch` | `networkbatch` | the batched runtime of the net library |
| `t1-plloader` | `t1-cx-tools-packageloader` | `packageloader` | the package loader |
| `t1-profiling` | `t1-cx-runtime-profiling` | `profiling` | the CPU and memory profiles |
| `t2-globals` | `t2-cx-core-environment` | `environment` | CXPATH, the debug switches and the error exit |
| `t2-plencoder` | `t2-cx-tools-packageencoder` | `packageencoder` | writes stored packages to disk |
| `t2-plgraph` | `t2-cx-tools-packagegraph` | `packagegraph` | the import graph of stored packages |
| `t2-pltree` | `t2-cx-tools-packagetree` | `packagetree` | the import tree of stored packages |
| `t3-ast` | `t3-cx-core-ast` | `ast` | the compiler IR of a CX program |
| `t3-packageloader` | `t3-cx-tools-packagecli` | `main` | the package loader command |
| `t4-actions` | `t4-cx-compiler-astbuild` | `astbuild` | the parser actions that build the AST |
| `t4-cipher` | `t4-cx-library-cipher` | `cipher` | the cipher library opcodes |
| `t4-cxnet` | `t4-cx-library-netopcodes` | `netopcodes` | the net library opcodes |
| `t4-cxos-core` | `t4-cx-library-osopcodes` | `osopcodes` | the time and os opcodes without a file system |
| `t4-execute` | `t4-cx-runtime-interpreter` | `interpreter` | the old loop that runs a CXProgram |
| `t4-opcodes` | `t4-cx-runtime-opcodes` | `opcodes` | the core opcodes |
| `t4-regexp` | `t4-cx-library-regexp` | `regexp` | the regexp library opcodes |
| `t5-cxos` | `t5-cx-library-fileopcodes` | `fileopcodes` | the file library opcodes |
| `t5-cxpartialparsing` | `t5-cx-compiler-partialparsing` | `partialparsing` | the declaration pass parser |
| `t5-plfileoutput` | `t5-cx-tools-packagefiles` | `packagefiles` | the files of stored packages |
| `t6-cx-core-run` | `t6-cx-command-core` | `main` | the cx-core command of cx build |
| `t6-cxfx` | `t6-cx-library-graphics` | `graphics` | the GL, GLFW, AL and gltext libraries |
| `t6-declextractor` | `t6-cx-compiler-declarationextractor` | `declarationextractor` | the declaration extractor |
| `t7-cxmobile` | `t7-cx-command-mobile` | `main` | the mobile command |
| `t7-init` | `t7-cx-runtime-libraries` | `libraries` | registers the libraries of a program |
| `t8-cxparsingcompletor` | `t8-cx-compiler-fullparsing` | `fullparsing` | the parser of function bodies |
| `t8-typechecker` | `t8-cx-compiler-typechecker` | `typechecker` | the type checker |
| `t9-astapi` | `t9-cx-core-astapi` | `astapi` | the AST editing API |
| `t9-cxparsing` | `t9-cx-compiler-parsing` | `parsing` | the parse driver |
| `t9-token-fuzzer` | `t9-cx-tools-tokenfuzzer` | `main` | the token fuzzer |
| `t10-cxrepl` | `t10-cx-tools-repl` | `repl` | the REPL |
| `t10-generator` | `t10-cx-tools-programgenerator` | `programgenerator` | the random program generator |
| `t11-cx` | `t11-cx-command-cx` | `main` | the cx command |

## Renames

### t0-cx-core-assert

| Step | Kind | Old | New |
|---|---|---|---|
| 5 | constant | `AssertExitCode` | `ExitCode` |
| 5 | function | `AssertInvariantTrue` | `InvariantTrue` |
| 5 | function | `AssertPostconditionTrue` | `PostconditionTrue` |
| 5 | function | `AssertPreconditionTrue` | `PreconditionTrue` |
| 5 | function | `AssertUnreachable` | `Unreachable` |

### t0-cx-core-objmodelgen

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | local_variable | `objmodelgenCheck data` | `fileBytes` |
| 3 | local_variable | `objmodelgenRender data` | `templateFields` |
| 3 | local_variable | `objmodelgenSitesCodeCheck data` | `fileBytes` |
| 3 | local_variable | `objmodelgenSitesDoc data` | `fileBytes` |
| 7 | field | `objmodelgenObject.Obj` | `Object` |

### t0-cxhash

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | parameter | `CxHashOfBytes data` | `input` |

### t0-cxtestrunner

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | function | `HasAll` | `BitsAllSet` |
| 4 | function | `NewTestRunner` | `TestRunnerNew` |
| 4 | method | `TestRunner.RunEx` | `ExRun` |
| 4 | method | `TestRunner.RunMissing` | `MissingRun` |
| 4 | method | `TestRunner.RunWant` | `WantRun` |
| 4 | method | `TestRunner.RunWantEx` | `WantExRun` |
| 4 | method | `TestRunner.RunWantInput` | `WantInputRun` |

### t0-cx-tools-names

| Step | Kind | Old | New |
|---|---|---|---|
| 4 | method | `GoCollector.AddAt` | `AtAdd` |
| 7 | function | `ExprUsesIota` | `ExpressionUsesIota` |
| 7 | method | `GoCollector.FuncDeclNames` | `FuncDeclarationNames` |
| 7 | method | `GoCollector.GenDeclNames` | `GenDeclarationNames` |
| 7 | method | `GoCollector.TypeExprNames` | `TypeExpressionNames` |

### t0-cx-tools-strhot

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | parameter | `calleeOf info` | `typeFacts` |
| 3 | parameter | `constant info` | `typeFacts` |
| 3 | parameter | `failing info` | `typeFacts` |
| 3 | field | `function.handler` | `opcodeGoFunction` |
| 3 | function | `handlerSignature` | `opcodeGoFunctionSignature` |
| 3 | function | `isString` | `stringType` |
| 3 | local_variable | `load info` | `typeFacts` |
| 3 | parameter | `scanFile info` | `typeFacts` |
| 3 | parameter | `stringOperation info` | `typeFacts` |
| 3 | parameter | `suggestClass info` | `typeFacts` |

### t0-packages

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | function | `IsBuiltinImport` | `ImportBuiltin` |
| 3 | function | `IsDefaultPackage` | `PackageDefault` |
| 4 | function | `CheckImport` | `ImportCheck` |

### t0-types

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | method | `Code.IsPrimitive` | `Primitive` |
| 3 | parameter | `Compute_obj_size data` | `payload` |
| 3 | function | `Get_obj_data` | `Get_obj_payload` |
| 3 | parameter | `Make_obj data` | `payload` |
| 3 | method | `Pointer.IsValid` | `Valid` |
| 3 | function | `Read_obj_data` | `Read_obj_payload` |
| 3 | field | `Type.isPrimitive` | `primitive` |
| 3 | function | `Write_obj_data` | `Write_obj_payload` |
| 3 | function | `Write_str_data` | `Write_str_payload` |
| 6 | constant | `AFF_SIZE` | `AffSize` |
| 6 | constant | `ARRAY_SIZE` | `ArraySize` |
| 6 | constant | `BOOL_SIZE` | `BoolSize` |
| 6 | function | `Cast_f32_to_ptr` | `F32ToPtrCast` |
| 6 | function | `Cast_f32_to_ptr` | `F32ToPtrCast` |
| 6 | function | `Cast_f64_to_ptr` | `F64ToPtrCast` |
| 6 | function | `Cast_f64_to_ptr` | `F64ToPtrCast` |
| 6 | function | `Cast_i32_to_ptr` | `I32ToPtrCast` |
| 6 | function | `Cast_i32_to_ptr` | `I32ToPtrCast` |
| 6 | function | `Cast_i64_to_ptr` | `I64ToPtrCast` |
| 6 | function | `Cast_i64_to_ptr` | `I64ToPtrCast` |
| 6 | function | `Cast_int_to_ptr` | `IntToPtrCast` |
| 6 | function | `Cast_int_to_ptr` | `IntToPtrCast` |
| 6 | function | `Cast_ptr_to_i32` | `PtrToI32Cast` |
| 6 | function | `Cast_ptr_to_i32` | `PtrToI32Cast` |
| 6 | function | `Cast_ptr_to_int` | `PtrToIntCast` |
| 6 | function | `Cast_ptr_to_int` | `PtrToIntCast` |
| 6 | function | `Cast_ptr_to_ui32` | `PtrToUI32Cast` |
| 6 | function | `Cast_ptr_to_ui32` | `PtrToUI32Cast` |
| 6 | function | `Cast_ptr_to_ui64` | `PtrToUI64Cast` |
| 6 | function | `Cast_ptr_to_ui64` | `PtrToUI64Cast` |
| 6 | function | `Cast_sint_to_sptr` | `SintToSptrCast` |
| 6 | function | `Cast_ui64_to_ptr` | `UI64ToPtrCast` |
| 6 | function | `Cast_ui64_to_ptr` | `UI64ToPtrCast` |
| 6 | function | `Compute_obj_size` | `ObjectSizeCompute` |
| 6 | enum_value | `CX_ASSERT` | `CXAssert` |
| 6 | enum_value | `CX_COMPILATION_ERROR` | `CXCompilationError` |
| 6 | enum_value | `CX_INTERNAL_ERROR` | `CXInternalError` |
| 6 | enum_value | `CX_PANIC` | `CXPanic` |
| 6 | enum_value | `CX_RUNTIME_ERROR` | `CXRuntimeError` |
| 6 | enum_value | `CX_RUNTIME_HEAP_EXHAUSTED_ERROR` | `CXRuntimeHeapExhaustedError` |
| 6 | enum_value | `CX_RUNTIME_INVALID_ARGUMENT` | `CXRuntimeInvalidArgument` |
| 6 | enum_value | `CX_RUNTIME_INVALID_CAST` | `CXRuntimeInvalidCast` |
| 6 | enum_value | `CX_RUNTIME_LIBRARY_DISABLED` | `CXRuntimeLibraryDisabled` |
| 6 | enum_value | `CX_RUNTIME_NOT_IMPLEMENTED` | `CXRuntimeNotImplemented` |
| 6 | enum_value | `CX_RUNTIME_SLICE_INDEX_OUT_OF_RANGE` | `CXRuntimeSliceIndexOutOfRange` |
| 6 | enum_value | `CX_RUNTIME_STACK_OVERFLOW_ERROR` | `CXRuntimeStackOverflowError` |
| 6 | enum_value | `CX_SUCCESS` | `CXSuccess` |
| 6 | constant | `F32_SIZE` | `F32Size` |
| 6 | constant | `F64_SIZE` | `F64Size` |
| 6 | constant | `FORWARDING_ADDRESS_SIZE` | `ForwardingAddressSize` |
| 6 | function | `Get_obj_header` | `ObjectHeaderGet` |
| 6 | function | `GetSlice_byte` | `SliceByteGet` |
| 6 | constant | `I16_SIZE` | `I16Size` |
| 6 | constant | `I32_SIZE` | `I32Size` |
| 6 | constant | `I64_SIZE` | `I64Size` |
| 6 | constant | `I8_SIZE` | `I8Size` |
| 6 | function | `Make_obj` | `ObjectMake` |
| 6 | constant | `MARK_SIZE` | `MarkSize` |
| 6 | constant | `MAX_INT` | `MaxInt` |
| 6 | constant | `MAX_INT32` | `MaxInt32` |
| 6 | constant | `MAX_UINT` | `MaxUint` |
| 6 | constant | `MAX_UINT16` | `MaxUint16` |
| 6 | constant | `MAX_UINT32` | `MaxUint32` |
| 6 | constant | `MAX_UINT64` | `MaxUint64` |
| 6 | constant | `MAX_UINT8` | `MaxUint8` |
| 6 | constant | `MIN_INT32` | `MinInt32` |
| 6 | constant | `OBJECT_GC_HEADER_SIZE` | `ObjectGCHeaderSize` |
| 6 | constant | `OBJECT_HEADER_SIZE` | `ObjectHeaderSize` |
| 6 | constant | `OBJECT_SIZE` | `ObjectSize` |
| 6 | constant | `POINTER_SIZE` | `PointerSize` |
| 6 | constant | `POINTER_SIZE` | `PointerSize` |
| 6 | function | `Read_bool` | `BoolRead` |
| 6 | function | `Read_f32` | `F32Read` |
| 6 | function | `Read_f64` | `F64Read` |
| 6 | function | `Read_i16` | `I16Read` |
| 6 | function | `Read_i32` | `I32Read` |
| 6 | function | `Read_i64` | `I64Read` |
| 6 | function | `Read_i8` | `I8Read` |
| 6 | function | `Read_obj` | `ObjectRead` |
| 6 | function | `Read_obj_forwarding_address` | `ObjectForwardingAddressRead` |
| 6 | function | `Read_obj_size` | `ObjectSizeRead` |
| 6 | function | `Read_ptr` | `PtrRead` |
| 6 | function | `Read_ptr` | `PtrRead` |
| 6 | function | `ReadSlice_f32` | `SliceF32Read` |
| 6 | function | `ReadSlice_f64` | `SliceF64Read` |
| 6 | function | `ReadSlice_i16` | `SliceI16Read` |
| 6 | function | `ReadSlice_i32` | `SliceI32Read` |
| 6 | function | `ReadSlice_i64` | `SliceI64Read` |
| 6 | function | `ReadSlice_i8` | `SliceI8Read` |
| 6 | function | `ReadSlice_ui16` | `SliceUI16Read` |
| 6 | function | `ReadSlice_ui32` | `SliceUI32Read` |
| 6 | function | `ReadSlice_ui64` | `SliceUI64Read` |
| 6 | function | `ReadSlice_ui8` | `SliceUI8Read` |
| 6 | function | `Read_str` | `StrRead` |
| 6 | function | `Read_str_bytes` | `StrBytesRead` |
| 6 | function | `Read_str_size` | `StrSizeRead` |
| 6 | function | `Read_ui16` | `UI16Read` |
| 6 | function | `Read_ui32` | `UI32Read` |
| 6 | function | `Read_ui64` | `UI64Read` |
| 6 | function | `Read_ui8` | `UI8Read` |
| 6 | constant | `SLICE_SIZE` | `SliceSize` |
| 6 | constant | `STR_SIZE` | `StrSize` |
| 6 | constant | `UI16_SIZE` | `UI16Size` |
| 6 | constant | `UI32_SIZE` | `UI32Size` |
| 6 | constant | `UI64_SIZE` | `UI64Size` |
| 6 | constant | `UI8_SIZE` | `UI8Size` |
| 6 | function | `Write_bool` | `BoolWrite` |
| 6 | function | `Write_f32` | `F32Write` |
| 6 | function | `Write_f64` | `F64Write` |
| 6 | function | `Write_i16` | `I16Write` |
| 6 | function | `Write_i32` | `I32Write` |
| 6 | function | `Write_i64` | `I64Write` |
| 6 | function | `Write_i8` | `I8Write` |
| 6 | function | `Write_obj_forwarding_address` | `ObjectForwardingAddressWrite` |
| 6 | function | `Write_obj_mark` | `ObjectMarkWrite` |
| 6 | function | `Write_obj_size` | `ObjectSizeWrite` |
| 6 | function | `Write_ptr` | `PtrWrite` |
| 6 | function | `Write_ptr` | `PtrWrite` |
| 6 | function | `WriteSlice_byte` | `SliceByteWrite` |
| 6 | function | `Write_ui16` | `UI16Write` |
| 6 | function | `Write_ui32` | `UI32Write` |
| 6 | function | `Write_ui64` | `UI64Write` |
| 6 | function | `Write_ui8` | `UI8Write` |

### t0-util

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | local_variable | `CXStatFile fileInfo` | `fileStat` |
| 3 | local_variable | `listDir fileInfo` | `entries` |
| 3 | parameter | `listDirRecursive info` | `entryStat` |
| 2 | package | `util` | `sandbox`, `sourceargs`, `filelog`, `stringpool` |
| 2 | function | `ParseArgsForCX` | `sourceargs.Parse` |
| 2 | function | `CXOpenFile`, `CXCreateFile`, `CXStatFile`, `CXLogFile` | `filelog.Open`, `Create`, `Stat`, `Enable` |
| 2 | function | `ProgramSandbox*`, `ProgramFileSystemPath*` | `sandbox.RootOpen`, `FileOpen`, `FileCreate`, `FileOpenReadWrite`, `FileRemove`, `FileRead`, `Path`, `PathFromBytes`, `PathLengthMaximum` |
| 2 | function | `ProgramTextStringView` | `stringpool.StringView` |

### t10-cxrepl

| Step | Kind | Old | New |
|---|---|---|---|
| 5 | global_variable | `ReplTargetFn` | `TargetFunction` |
| 5 | global_variable | `ReplTargetMod` | `TargetModule` |
| 5 | global_variable | `ReplTargetStrct` | `TargetStruct` |

### t10-generator

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | function | `IsJumpOperator` | `OpcodeOperatorJump` |
| 4 | function | `GenerateRandomExpressions` | `RandomExpressionsGenerate` |
| 4 | function | `GenerateSampleProgram` | `SampleProgramGenerate` |
| 4 | function | `GetFunctionSet` | `FunctionSetGet` |
| 7 | function | `calcFnSize` | `calculateFunctionSize` |
| 7 | function | `copySig` | `copySignature` |
| 7 | function | `getRandFn` | `getRandomFunction` |
| 7 | function | `getRandInp` | `getRandomInput` |
| 7 | function | `sigSize` | `signatureSize` |
| 7 | function | `sigType` | `signatureType` |

### t11-cx

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | local_variable | `buildCorePath info` | `pathStat` |
| 3 | local_variable | `buildOptionsParse hasValue` | `valueGiven` |
| 3 | local_variable | `buildOutputDefault info` | `pathStat` |
| 3 | local_variable | `buildPathSame firstInfo` | `firstStat` |
| 3 | local_variable | `buildPathSame secondInfo` | `secondStat` |
| 4 | function | `GetCXPath` | `CXPathGet` |

### t1-constants

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | field | `LibraryFunctionDefinition.Handler` | `GoFunction` |
| 3 | field | `OpcodeDefinition.Handler` | `GoFunction` |
| 6 | constant | `ARRAY_ELEMENT_PREFIX` | `ArrayElementPrefix` |
| 6 | constant | `CALLSTACK_SIZE` | `CallstackSize` |
| 6 | enum_value | `DECL_ARRAY` | `DeclarationArray` |
| 6 | enum_value | `DECL_BASIC` | `DeclarationBasic` |
| 6 | enum_value | `DECL_DEREF` | `DeclarationDeref` |
| 6 | enum_value | `DECL_FUNC` | `DeclarationFunc` |
| 6 | enum_value | `DECL_INDEXING` | `DeclarationIndexing` |
| 6 | enum_value | `DECL_POINTER` | `DeclarationPointer` |
| 6 | enum_value | `DECL_SLICE` | `DeclarationSlice` |
| 6 | enum_value | `DECL_STRUCT` | `DeclarationStruct` |
| 6 | enum_value | `DECL_UNUSED` | `DeclarationUnused` |
| 6 | enum_value | `DEREF_ARRAY` | `DerefArray` |
| 6 | enum_value | `DEREF_DEREF` | `DerefDeref` |
| 6 | enum_value | `DEREF_FIELD` | `DerefField` |
| 6 | enum_value | `DEREF_POINTER` | `DerefPointer` |
| 6 | enum_value | `DEREF_SLICE` | `DerefSlice` |
| 6 | enum_value | `DEREF_UNUSED` | `DerefUnused` |
| 6 | constant | `ELEMENT_PREFIX` | `ElementPrefix` |
| 6 | constant | `END_OF_ARITHMETIC_OPERATORS` | `EndOfArithmeticOperators` |
| 6 | constant | `END_OF_COMPARISON_OPERATORS` | `EndOfComparisonOperators` |
| 6 | constant | `END_OF_NAMED_OPCODES` | `EndOfNamedOpcodes` |
| 6 | constant | `END_OF_OPERATORS` | `EndOfOperators` |
| 6 | constant | `END_PARSE_OPS` | `EndParseOps` |
| 6 | constant | `ESCAPE_OFFSET_FLAG` | `EscapeOffsetFlag` |
| 6 | constant | `ID_FN` | `IDFunction` |
| 6 | constant | `INIT_FN` | `InitFunction` |
| 6 | global_variable | `INIT_HEAP_SIZE` | `HeapSizeInit` |
| 6 | constant | `LABEL_PREFIX` | `LabelPrefix` |
| 6 | constant | `LITERAL_PLACEHOLDER` | `LiteralPlaceholder` |
| 6 | constant | `LOCAL_PREFIX` | `LocalPrefix` |
| 6 | constant | `MAIN_FUNC` | `MainFunc` |
| 6 | constant | `MAIN_PKG` | `MainPkg` |
| 6 | global_variable | `MAX_HEAP_FREE_RATIO` | `MaxHeapFreeRatio` |
| 6 | global_variable | `MAX_HEAP_SIZE` | `MaxHeapSize` |
| 6 | constant | `METHOD_RESULT_PREFIX` | `MethodResultPrefix` |
| 6 | global_variable | `MIN_HEAP_FREE_RATIO` | `MinHeapFreeRatio` |
| 6 | constant | `MULTIPLE_VALUE_PREFIX` | `MultipleValuePrefix` |
| 6 | constant | `NON_ASSIGN_PREFIX` | `NonAssignPrefix` |
| 6 | constant | `NULL_HEAP_ADDRESS` | `NullHeapAddress` |
| 6 | constant | `NULL_HEAP_ADDRESS_OFFSET` | `NullHeapAddressOffset` |
| 6 | constant | `OP_ABS_JMP` | `OpAbsJmp` |
| 6 | constant | `OP_ADD` | `OpAdd` |
| 6 | constant | `OP_APPEND` | `OpAppend` |
| 6 | constant | `OP_ASSERT` | `OpAssert` |
| 6 | constant | `OP_BITAND` | `OpBitand` |
| 6 | constant | `OP_BITCLEAR` | `OpBitclear` |
| 6 | constant | `OP_BITOR` | `OpBitor` |
| 6 | constant | `OP_BITSHL` | `OpBitshl` |
| 6 | constant | `OP_BITSHR` | `OpBitshr` |
| 6 | constant | `OP_BITXOR` | `OpBitxor` |
| 6 | constant | `OP_BOOL_AND` | `OpBoolAnd` |
| 6 | constant | `OP_BOOL_NOT` | `OpBoolNot` |
| 6 | constant | `OP_BOOL_OR` | `OpBoolOr` |
| 6 | constant | `OP_BREAK` | `OpBreak` |
| 6 | constant | `OP_CONTINUE` | `OpContinue` |
| 6 | constant | `OP_COPY` | `OpCopy` |
| 6 | constant | `OP_DEBUG_PRINT_STACK` | `OpDebugPrintStack` |
| 6 | constant | `OP_DESERIALIZE` | `OpDeserialize` |
| 6 | constant | `OP_DIV` | `OpDiv` |
| 6 | constant | `OP_EQUAL` | `OpEqual` |
| 6 | constant | `OP_GOTO` | `OpGoto` |
| 6 | constant | `OP_GT` | `OpGt` |
| 6 | constant | `OP_GTEQ` | `OpGteq` |
| 6 | constant | `OP_IDENTITY` | `OpIdentity` |
| 6 | constant | `OP_INSERT` | `OpInsert` |
| 6 | constant | `OP_JMP` | `OpJmp` |
| 6 | constant | `OP_JMP_EQ` | `OpJmpEq` |
| 6 | constant | `OP_JMP_GT` | `OpJmpGt` |
| 6 | constant | `OP_JMP_GTEQ` | `OpJmpGteq` |
| 6 | constant | `OP_JMP_LT` | `OpJmpLt` |
| 6 | constant | `OP_JMP_LTEQ` | `OpJmpLteq` |
| 6 | constant | `OP_JMP_NOT_ZERO` | `OpJmpNotZero` |
| 6 | constant | `OP_JMP_UNEQ` | `OpJmpUneq` |
| 6 | constant | `OP_JMP_ZERO` | `OpJmpZero` |
| 6 | constant | `OP_LIBRARY_CALL` | `OpLibraryCall` |
| 6 | constant | `OP_LT` | `OpLt` |
| 6 | constant | `OP_LTEQ` | `OpLteq` |
| 6 | constant | `OP_MOD` | `OpModule` |
| 6 | constant | `OP_MUL` | `OpMul` |
| 6 | constant | `OP_NEG` | `OpNeg` |
| 6 | constant | `OP_NOP` | `OpNop` |
| 6 | constant | `OP_PANIC` | `OpPanic` |
| 6 | constant | `OP_REMOVE` | `OpRemove` |
| 6 | constant | `OP_RESIZE` | `OpResize` |
| 6 | constant | `OP_SERIALIZE` | `OpSerialize` |
| 6 | constant | `OP_SUB` | `OpSub` |
| 6 | constant | `OP_TEST` | `OpTest` |
| 6 | constant | `OP_UNEQUAL` | `OpUnequal` |
| 6 | constant | `OS_ARGS` | `OsArgs` |
| 6 | constant | `OS_PKG` | `OsPkg` |
| 6 | enum_value | `PARAM_DEFAULT` | `ParamDefault` |
| 6 | enum_value | `PARAM_SLICE` | `ParamSlice` |
| 6 | enum_value | `PARAM_STRUCT` | `ParamStruct` |
| 6 | enum_value | `PASSBY_REFERENCE` | `PassbyReference` |
| 6 | enum_value | `PASSBY_VALUE` | `PassbyValue` |
| 6 | constant | `RESULT_NAME` | `ResultName` |
| 6 | constant | `SLICE_HEADER_SIZE` | `SliceHeaderSize` |
| 6 | constant | `SLICE_LITERAL_PREFIX` | `SliceLiteralPrefix` |
| 6 | constant | `STACK_OVERFLOW_ERROR` | `StackOverflowError` |
| 6 | global_variable | `STACK_SIZE` | `StackSize` |
| 6 | constant | `START_OF_ARITHMETIC_OPERATORS` | `StartOfArithmeticOperators` |
| 6 | constant | `START_OF_COMPARISON_OPERATORS` | `StartOfComparisonOperators` |
| 6 | constant | `START_OF_OPERATORS` | `StartOfOperators` |
| 6 | constant | `START_PARSE_OPS` | `StartParseOps` |
| 6 | constant | `STDLIB_PKG` | `StdlibPkg` |
| 6 | constant | `SYS_INIT_FUNC` | `SysInitFunc` |

### t1-netbatch

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | function | `netDataEntry` | `netPayloadEntry` |

### t1-plloader

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | parameter | `File.UnmarshalBinary data` | `encoded` |
| 3 | parameter | `PackageList.UnmarshalBinary data` | `encoded` |
| 3 | parameter | `Package.UnmarshalBinary data` | `encoded` |
| 4 | function | `LoadCXProgram` | `CXProgramLoad` |
| 4 | function | `ReadProgram` | `ProgramRead` |
| 4 | function | `RemoveDuplicates` | `DuplicatesRemove` |

### t1-profiling

| Step | Kind | Old | New |
|---|---|---|---|
| 4 | function | `StartCPUProfile` | `CPUProfileStart` |
| 4 | function | `StartCPUProfileFile` | `CPUProfileFileStart` |
| 4 | function | `StartCPUProfileOnFile` | `CPUProfileOnFileStart` |
| 4 | function | `StartProfile` | `ProfileStart` |
| 4 | function | `StopProfile` | `ProfileStop` |

### t2-cx-core-opcodegen

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | function | `handlerInline` | `goFunctionInline` |
| 3 | global_variable | `handlerPackages` | `goFunctionPackages` |
| 3 | local_variable | `librariesGenerated.dispatchFiles handler` | `goFunction` |

### t2-cx-runtime-gc

| Step | Kind | Old | New |
|---|---|---|---|
| 5 | enum | `GCKind` | `Kind` |
| 5 | enum_value | `GCKindArray` | `KindArray` |
| 5 | enum_value | `GCKindPointer` | `KindPointer` |
| 5 | enum_value | `GCKindSlice` | `KindSlice` |
| 5 | enum_value | `GCKindStr` | `KindStr` |
| 5 | enum_value | `GCKindStruct` | `KindStruct` |
| 5 | struct | `GCRoot` | `Root` |
| 5 | struct | `GCRootFrame` | `RootFrame` |
| 5 | struct | `GCRootTable` | `RootTable` |
| 5 | struct | `GCState` | `State` |
| 5 | function | `GCStatePrepare` | `StatePrepare` |
| 5 | function | `GCStateTypeAdd` | `StateTypeAdd` |
| 5 | function | `GCStateViewsSize` | `StateViewsSize` |
| 5 | function | `GCStateWorkSize` | `StateWorkSize` |
| 5 | struct | `GCType` | `Type` |
| 5 | struct | `GCTypeField` | `TypeField` |
| 5 | struct | `GCVisit` | `Visit` |
| 7 | field | `GCType.GCTypeElem` | `GCTypeElement` |
| 7 | field | `GCType.GCTypeElemSize` | `GCTypeElementSize` |

### t2-globals

| Step | Kind | Old | New |
|---|---|---|---|
| 6 | global_variable | `DBG_GOLANG_STACK_TRACE` | `DebugGolangStackTrace` |

### t2-plencoder

| Step | Kind | Old | New |
|---|---|---|---|
| 4 | function | `SaveFilesToDisk` | `FilesToDiskSave` |
| 4 | function | `SavePackagesToDisk` | `PackagesToDiskSave` |

### t2-plgraph

| Step | Kind | Old | New |
|---|---|---|---|
| 4 | function | `GetImportGraph` | `ImportGraphGet` |

### t2-pltree

| Step | Kind | Old | New |
|---|---|---|---|
| 4 | function | `AddDependenciesRecur` | `DependenciesRecursiveAdd` |
| 4 | function | `GetImportTree` | `ImportTreeGet` |

### t3-ast

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | field | `AffordanceValue.AffordanceValueItem` | `AffordanceValueTarget` |
| 3 | parameter | `AllocateStrAt data` | `text` |
| 3 | local_variable | `AllocateStrBytesAt data` | `payload` |
| 3 | parameter | `AllocateStr data` | `text` |
| 3 | local_variable | `argumentDebugAlloc manager` | `argumentDebugMng` |
| 3 | method | `CXArgument.IsPointer` | `PointerTyped` |
| 3 | method | `CXArgument.IsSlice` | `SliceTyped` |
| 3 | method | `CXArgument.IsString` | `StringTyped` |
| 3 | method | `CXArgument.IsStruct` | `StructTyped` |
| 3 | method | `CXExpression.IsArrayLiteral` | `TypeArrayLiteral` |
| 3 | method | `CXExpression.IsBreak` | `BreakOperator` |
| 3 | method | `CXExpression.IsContinue` | `ContinueOperator` |
| 3 | method | `CXExpression.IsMethodCall` | `TypeMethodCall` |
| 3 | method | `CXExpression.IsStructLiteral` | `TypeStructLiteral` |
| 3 | method | `CXExpression.IsUndType` | `UndTyped` |
| 3 | method | `CXFunction.IsAtomic` | `Atomic` |
| 3 | method | `CXFunction.IsBuiltIn` | `BuiltIn` |
| 3 | method | `CXFunction.IsLocalVariable` | `LocalVariableDeclared` |
| 3 | local_variable | `CxIrState.CxIrStateTeardown1 manager` | `programMng` |
| 3 | local_variable | `CXProgram.AllocCXArgument manager` | `<object>Mng` |
| 3 | local_variable | `CXProgram.AllocCXTypeSignatureCopy manager` | `<object>Mng` |
| 3 | local_variable | `CXProgram.AllocCXTypeSignature manager` | `<object>Mng` |
| 3 | local_variable | `CXProgram.CXProgramFunctionAlloc manager` | `<object>Mng` |
| 3 | local_variable | `CXProgram.CXProgramFunctionLiveSlots manager` | `<object>Mng` |
| 3 | local_variable | `CXProgram.CXProgramFunctionReleaseNew manager` | `<object>Mng` |
| 3 | local_variable | `CXProgram.CXProgramFunctionSlotLive manager` | `<object>Mng` |
| 3 | local_variable | `CXProgram.GetCXArg manager` | `<object>Mng` |
| 3 | local_variable | `CXProgram.GetCXLine manager` | `<object>Mng` |
| 3 | local_variable | `CXProgram.GetCXTypeSignatureArrayFromArray manager` | `<object>Mng` |
| 3 | local_variable | `CXProgram.GetCXTypeSignatureFromArray manager` | `<object>Mng` |
| 3 | local_variable | `CXProgram.GetCXTypeSignatureStructFromArray manager` | `<object>Mng` |
| 3 | local_variable | `CXProgram.ReleaseCXArg manager` | `<object>Mng` |
| 3 | local_variable | `CXProgram.ReleaseCXTypeSignature manager` | `<object>Mng` |
| 3 | parameter | `CXTypeSignature.GetSize IsForUpdateSymbolsTable` | `forSymbolsTableUpdate` |
| 3 | field | `CXTypeSignature.IsDeref` | `Deref` |
| 3 | method | `CXTypeSignature.IsInnerReference` | `InnerReference` |
| 3 | parameter | `CXValue.Set_bool data` | `newValue` |
| 3 | parameter | `CXValue.Set_bytes data` | `newValue` |
| 3 | parameter | `CXValue.Set_f32 data` | `newValue` |
| 3 | parameter | `CXValue.Set_f64 data` | `newValue` |
| 3 | parameter | `CXValue.Set_i16 data` | `newValue` |
| 3 | parameter | `CXValue.Set_i32 data` | `newValue` |
| 3 | parameter | `CXValue.Set_i64 data` | `newValue` |
| 3 | parameter | `CXValue.Set_i8 data` | `newValue` |
| 3 | parameter | `CXValue.Set_ptr data` | `newValue` |
| 3 | parameter | `CXValue.Set_str_bytes data` | `newValue` |
| 3 | parameter | `CXValue.Set_str data` | `newValue` |
| 3 | parameter | `CXValue.Set_ui16 data` | `newValue` |
| 3 | parameter | `CXValue.Set_ui32 data` | `newValue` |
| 3 | parameter | `CXValue.Set_ui64 data` | `newValue` |
| 3 | parameter | `CXValue.Set_ui8 data` | `newValue` |
| 3 | field | `DereferenceStruct.isPointer` | `pointerTyped` |
| 3 | local_variable | `EscapedAllocate data` | `payload` |
| 3 | local_variable | `GetFormattedType arrayData` | `arraySignature` |
| 3 | function | `GetSliceData` | `GetSliceElements` |
| 3 | local_variable | `GetSlice dataLen` | `elementsLen` |
| 3 | local_variable | `GetSlice dataOffset` | `elementsOffset` |
| 3 | function | `IsArithmeticOperator` | `OpcodeOperatorArithmetic` |
| 3 | function | `IsComparisonOperator` | `OpcodeOperatorComparison` |
| 3 | function | `IsOperator` | `OpcodeOperator` |
| 3 | function | `IsTypeArrayAtomic` | `CXArgumentTypeArrayAtomic` |
| 3 | local_variable | `IsTypeArrayAtomic isThereDeclPointer` | `declPointerFound` |
| 3 | function | `IsTypeAtomic` | `CXArgumentTypeAtomic` |
| 3 | function | `IsTypePointerArrayAtomic` | `CXArgumentTypePointerArrayAtomic` |
| 3 | local_variable | `IsTypePointerArrayAtomic isThereDeclPointer` | `declPointerFound` |
| 3 | function | `IsTypePointerAtomic` | `CXArgumentTypePointerAtomic` |
| 3 | function | `IsTypePointerSliceAtomic` | `CXArgumentTypePointerSliceAtomic` |
| 3 | local_variable | `IsTypePointerSliceAtomic isThereDeclPointer` | `declPointerFound` |
| 3 | function | `IsTypePointerStruct` | `CXArgumentTypePointerStruct` |
| 3 | local_variable | `IsTypePointerStruct isThereDeclPointer` | `declPointerFound` |
| 3 | local_variable | `IsTypePointerStruct IsThereDeclSlice` | `declSliceFound` |
| 3 | function | `IsTypeSliceAtomic` | `CXArgumentTypeSliceAtomic` |
| 3 | local_variable | `IsTypeSliceAtomic isThereDeclPointer` | `declPointerFound` |
| 3 | function | `IsTypeStruct` | `CXArgumentTypeStruct` |
| 3 | local_variable | `IsTypeStruct isThereDeclPointer` | `declPointerFound` |
| 3 | local_variable | `IsTypeStruct IsThereDeclSlice` | `declSliceFound` |
| 3 | function | `IsValidSliceIndex` | `SliceIndexValid` |
| 3 | function_type | `OpcodeHandler` | `OpcodeGoFunction` |
| 3 | function | `processBuiltInOperators` | `builtInOperatorRun` |
| 3 | function | `processNonAtomicOperators` | `nonAtomicOperatorRun` |
| 3 | parameter | `ResizeMemory isExpand` | `expand` |
| 3 | struct | `SerializedDataSize` | `SerializedSizes` |
| 3 | function | `SerializeDebugInfo` | `SerializeSizes` |
| 3 | parameter | `serializeEncoder.serializeEncoderLength isNil` | `nilValue` |
| 3 | local_variable | `serializeExpressions manager` | `<object>Mng` |
| 3 | local_variable | `serializeFunctionGidValid manager` | `<object>Mng` |
| 3 | local_variable | `serializeFunctions manager` | `<object>Mng` |
| 3 | loop_variable | `serializeUnusedManagersCheck manager` | `unusedMng` |
| 3 | local_variable | `SliceAppendWriteByte outputSliceData` | `outputSliceElements` |
| 3 | local_variable | `SliceAppendWrite outputSliceData` | `outputSliceElements` |
| 3 | local_variable | `SliceCopyEx outputSliceData` | `outputSliceElements` |
| 3 | local_variable | `SliceInsert outputSliceData` | `outputSliceElements` |
| 3 | local_variable | `SliceRemove sliceData` | `sliceElements` |
| 4 | function | `AllocateSeq` | `SequenceAllocate` |
| 4 | function | `AllocateStr` | `StrAllocate` |
| 4 | function | `AllocateStrAt` | `StrAtAllocate` |
| 4 | function | `AllocateStrBytesAt` | `StrBytesAtAllocate` |
| 4 | function | `AllocCXTypeSignatureOfCXArg` | `CXTypeSignatureOfCXArgAlloc` |
| 4 | function | `AllocCXTypeSignatureOfCXArgBuilder` | `CXTypeSignatureOfCXArgBuilderAlloc` |
| 4 | function | `AllocCXTypeSignatureOfCXArgDeprecate` | `CXTypeSignatureOfCXArgDeprecateAlloc` |
| 4 | function | `AllocCXTypeSignatureOfCXArgForStructs` | `CXTypeSignatureOfCXArgForStructsAlloc` |
| 4 | function | `AppendPrintableValue` | `PrintableValueAppend` |
| 4 | function | `BuildStrPackages` | `StrPackagesBuild` |
| 4 | function | `CalculateDereferences` | `DereferencesCalculate` |
| 4 | function | `CompileErrorCodeWarning` | `ErrorCodeWarningCompile` |
| 4 | method | `CompileErrorList.CompileErrorListInit` | `ErrorListInitCompile` |
| 4 | function | `CompileErrorSyntaxText` | `ErrorSyntaxTextCompile` |
| 4 | method | `CXArgument.GetAssignmentElement` | `AssignmentElementGet` |
| 4 | method | `CXArgument.GetSize` | `SizeGet` |
| 4 | method | `CXArgument.GetType` | `TypeGet` |
| 4 | method | `CXArgument.SetPackage` | `PackageSet` |
| 4 | method | `CXArgument.SetType` | `TypeSet` |
| 4 | method | `CXAtomicOperator.AddInput` | `InputAdd` |
| 4 | method | `CXAtomicOperator.AddLabel` | `LabelAdd` |
| 4 | method | `CXAtomicOperator.AddOutput` | `OutputAdd` |
| 4 | method | `CXAtomicOperator.GetInputs` | `InputsGet` |
| 4 | method | `CXAtomicOperator.GetOperatorName` | `OperatorNameGet` |
| 4 | method | `CXAtomicOperator.GetOperatorNameBytes` | `OperatorNameBytesGet` |
| 4 | method | `CXAtomicOperator.GetOutputs` | `OutputsGet` |
| 4 | method | `CXExpression.GetLabel` | `LabelGet` |
| 4 | method | `CXFunction.AddExpression` | `ExpressionAdd` |
| 4 | method | `CXFunction.AddExpressionByLineNumber` | `ExpressionByLineNumberAdd` |
| 4 | method | `CXFunction.AddInput` | `InputAdd` |
| 4 | method | `CXFunction.AddLocalVariableName` | `LocalVariableNameAdd` |
| 4 | method | `CXFunction.AddOutput` | `OutputAdd` |
| 4 | method | `CXFunction.GetExpressionByLabel` | `ExpressionByLabelGet` |
| 4 | method | `CXFunction.GetExpressionByLine` | `ExpressionByLineGet` |
| 4 | method | `CXFunction.GetExpressions` | `ExpressionsGet` |
| 4 | method | `CXFunction.GetInputs` | `InputsGet` |
| 4 | method | `CXFunction.GetOutputs` | `OutputsGet` |
| 4 | method | `CXFunction.RemoveExpression` | `ExpressionRemove` |
| 4 | method | `CXFunction.RemoveLocalVariableFromArray` | `LocalVariableFromArrayRemove` |
| 4 | method | `CXPackage.AddFunction` | `FunctionAdd` |
| 4 | method | `CXPackage.AddGlobal` | `GlobalAdd` |
| 4 | method | `CXPackage.AddImport` | `ImportAdd` |
| 4 | method | `CXPackage.AddMethod` | `MethodAdd` |
| 4 | method | `CXPackage.AddStruct` | `StructAdd` |
| 4 | method | `CXPackage.GetCurrentFunction` | `CurrentFunctionGet` |
| 4 | method | `CXPackage.GetFunction` | `FunctionGet` |
| 4 | method | `CXPackage.GetGlobal` | `GlobalGet` |
| 4 | method | `CXPackage.GetImport` | `ImportGet` |
| 4 | method | `CXPackage.GetMethod` | `MethodGet` |
| 4 | method | `CXPackage.GetStruct` | `StructGet` |
| 4 | method | `CXPackage.RemoveFunction` | `FunctionRemove` |
| 4 | method | `CXPackage.RemoveImport` | `ImportRemove` |
| 4 | method | `CXPackage.RemoveStruct` | `StructRemove` |
| 4 | method | `CXProgram.AddCXArgInArray` | `CXArgInArrayAdd` |
| 4 | method | `CXProgram.AddCXAtomicOp` | `CXAtomicOpAdd` |
| 4 | method | `CXProgram.AddCXTypeSignatureInArray` | `CXTypeSignatureInArrayAdd` |
| 4 | method | `CXProgram.AddNativeFunctionInArray` | `NativeFunctionInArrayAdd` |
| 4 | method | `CXProgram.AddPackage` | `PackageAdd` |
| 4 | method | `CXProgram.AddPackageInArray` | `PackageInArrayAdd` |
| 4 | method | `CXProgram.AddPointerArgsToCXArgsArray` | `PointerArgsToCXArgsArrayAdd` |
| 4 | method | `CXProgram.AddPointerArgsToTypeSignaturesArray` | `PointerArgsToTypeSignaturesArrayAdd` |
| 4 | method | `CXProgram.AddStructInArray` | `StructInArrayAdd` |
| 4 | method | `CXProgram.AllocCXArgument` | `CXArgumentAlloc` |
| 4 | method | `CXProgram.AllocCXTypeSignature` | `CXTypeSignatureAlloc` |
| 4 | method | `CXProgram.AllocCXTypeSignatureCopy` | `CXTypeSignatureCopyAlloc` |
| 4 | method | `CXProgram.ConvertIndexArgsToPointerArgs` | `IndexArgsToPointerArgsConvert` |
| 4 | method | `CXProgram.ConvertIndexTypeSignaturesToPointerArgs` | `IndexTypeSignaturesToPointerArgsConvert` |
| 4 | method | `CXProgram.GetCurrentCall` | `CurrentCallGet` |
| 4 | method | `CXProgram.GetCurrentFunction` | `CurrentFunctionGet` |
| 4 | method | `CXProgram.GetCurrentPackage` | `CurrentPackageGet` |
| 4 | method | `CXProgram.GetCXArg` | `CXArgGet` |
| 4 | method | `CXProgram.GetCXArgFromArray` | `CXArgFromArrayGet` |
| 4 | method | `CXProgram.GetCXAtomicOp` | `CXAtomicOpGet` |
| 4 | method | `CXProgram.GetCXAtomicOpFromExpressions` | `CXAtomicOpFromExpressionsGet` |
| 4 | method | `CXProgram.GetCXLine` | `CXLineGet` |
| 4 | method | `CXProgram.GetCXTypeSignatureArrayFromArray` | `CXTypeSignatureArrayFromArrayGet` |
| 4 | method | `CXProgram.GetCXTypeSignatureFromArray` | `CXTypeSignatureFromArrayGet` |
| 4 | method | `CXProgram.GetCXTypeSignatureStructFromArray` | `CXTypeSignatureStructFromArrayGet` |
| 4 | method | `CXProgram.GetFunction` | `FunctionGet` |
| 4 | method | `CXProgram.GetOperation` | `OperationGet` |
| 4 | method | `CXProgram.GetPackage` | `PackageGet` |
| 4 | method | `CXProgram.GetPackageFromArray` | `PackageFromArrayGet` |
| 4 | method | `CXProgram.GetPreviousCXAtomicOpFromExpressions` | `PreviousCXAtomicOpFromExpressionsGet` |
| 4 | method | `CXProgram.GetPreviousCXLine` | `PreviousCXLineGet` |
| 4 | method | `CXProgram.GetStruct` | `StructGet` |
| 4 | method | `CXProgram.GetStructFromArray` | `StructFromArrayGet` |
| 4 | method | `CXProgram.PrintProgram` | `ProgramPrint` |
| 4 | method | `CXProgram.ReleaseCXArg` | `CXArgRelease` |
| 4 | method | `CXProgram.ReleaseCXArgBuilder` | `CXArgBuilderRelease` |
| 4 | method | `CXProgram.ReleaseCXArgUnlessNamed` | `CXArgUnlessNamedRelease` |
| 4 | method | `CXProgram.ReleaseCXArgWithDebug` | `CXArgWithDebugRelease` |
| 4 | method | `CXProgram.ReleaseCXTypeSignature` | `CXTypeSignatureRelease` |
| 4 | method | `CXProgram.RunBuffersPrepare` | `BuffersPrepareRun` |
| 4 | method | `CXProgram.RunCallbackValues` | `CallbackValuesRun` |
| 4 | method | `CXStruct.AddField` | `FieldAdd` |
| 4 | method | `CXStruct.GetField` | `FieldGet` |
| 4 | method | `CXStruct.GetStructSize` | `StructSizeGet` |
| 4 | method | `CXTypeSignature.AppendValueTypeName` | `ValueTypeNameAppend` |
| 4 | method | `CXTypeSignature.GetArrayLength` | `ArrayLengthGet` |
| 4 | method | `CXTypeSignature.GetArraySize` | `ArraySizeGet` |
| 4 | method | `CXTypeSignature.GetCXArgFormat` | `CXArgFormatGet` |
| 4 | method | `CXTypeSignature.GetOffset` | `OffsetGet` |
| 4 | method | `CXTypeSignature.GetSize` | `SizeGet` |
| 4 | method | `CXTypeSignature.GetStructType` | `StructTypeGet` |
| 4 | method | `CXTypeSignature.GetType` | `TypeGet` |
| 4 | method | `CXTypeSignature.GetTypeString` | `TypeStringGet` |
| 4 | method | `CXTypeSignature.GetValueTypeName` | `ValueTypeNameGet` |
| 4 | function | `GetArgSize` | `ArgSizeGet` |
| 4 | function | `GetCXTypeSignatureRepresentationOfCXArg` | `CXTypeSignatureRepresentationOfCXArgGet` |
| 4 | function | `GetDerefSize` | `DerefSizeGet` |
| 4 | function | `GetDerefSizeSlice` | `DerefSizeSliceGet` |
| 4 | function | `GetFinalOffset` | `FinalOffsetGet` |
| 4 | function | `GetFormattedName` | `FormattedNameGet` |
| 4 | function | `GetFormattedType` | `FormattedTypeGet` |
| 4 | function | `GetNativeSize` | `NativeSizeGet` |
| 4 | function | `GetSlice` | `SliceGet` |
| 4 | function | `GetSliceHeader` | `SliceHeaderGet` |
| 4 | function | `GetSliceLen` | `SliceLenGet` |
| 4 | function | `GetSliceOffset` | `SliceOffsetGet` |
| 4 | function | `MakeArgument` | `ArgumentMake` |
| 4 | function | `MakeAtomicOperatorExpression` | `AtomicOperatorExpressionMake` |
| 4 | function | `MakeCallStack` | `CallStackMake` |
| 4 | function | `MakeCXLineExpression` | `CXLineExpressionMake` |
| 4 | function | `MakeField` | `FieldMake` |
| 4 | function | `MakeFunction` | `FunctionMake` |
| 4 | function | `MakeGlobal` | `GlobalMake` |
| 4 | function | `MakePackage` | `PackageMake` |
| 4 | function | `MakePointer` | `PointerMake` |
| 4 | function | `MakeStruct` | `StructMake` |
| 4 | function | `MakeStructParameter` | `StructParameterMake` |
| 4 | function | `ResizeMemory` | `MemoryResize` |
| 4 | function | `SerializeAffordance` | `AffordanceSerialize` |
| 4 | function | `SerializeCXProgram` | `CXProgramSerialize` |
| 4 | function | `WriteToSlice` | `ToSliceWrite` |
| 6 | function | `CalculateDereference_Array` | `DereferenceArrayCalculate` |
| 6 | function | `CalculateDereference_Pointer` | `DereferencePointerCalculate` |
| 6 | function | `CalculateDereference_Slice` | `DereferenceSliceCalculate` |
| 6 | global_variable | `ConstCxArg_BOOL` | `ConstCXArgBool` |
| 6 | global_variable | `ConstCxArg_F32` | `ConstCXArgF32` |
| 6 | global_variable | `ConstCxArg_F64` | `ConstCXArgF64` |
| 6 | global_variable | `ConstCxArg_FUNC` | `ConstCXArgFunc` |
| 6 | global_variable | `ConstCxArg_I16` | `ConstCXArgI16` |
| 6 | global_variable | `ConstCxArg_I32` | `ConstCXArgI32` |
| 6 | global_variable | `ConstCxArg_I64` | `ConstCXArgI64` |
| 6 | global_variable | `ConstCXArg_I8` | `ConstCXArgI8` |
| 6 | global_variable | `ConstCxArg_PTR` | `ConstCXArgPtr` |
| 6 | global_variable | `ConstCxArg_PTR` | `ConstCXArgPtr` |
| 6 | global_variable | `ConstCxArg_STR` | `ConstCXArgStr` |
| 6 | global_variable | `ConstCxArg_UI16` | `ConstCXArgUI16` |
| 6 | global_variable | `ConstCxArg_UI32` | `ConstCXArgUI32` |
| 6 | global_variable | `ConstCxArg_UI64` | `ConstCXArgUI64` |
| 6 | global_variable | `ConstCxArg_UI8` | `ConstCXArgUI8` |
| 6 | global_variable | `ConstCxArg_UND_TYPE` | `ConstCXArgUndType` |
| 6 | enum_value | `CX_ARGUMENT` | `CXArgument` |
| 6 | enum_value | `CX_ATOMIC_OPERATOR` | `CXAtomicOperator` |
| 6 | enum_value | `CXEXPR_ARRAY_LITERAL` | `CxexprArrayLiteral` |
| 6 | enum_value | `CXEXPR_FUNCTION_CALL` | `CxexprFunctionCall` |
| 6 | enum_value | `CXEXPR_METHOD_CALL` | `CxexprMethodCall` |
| 6 | enum_value | `CXEXPR_SHORT_DECLARATION_SHARED` | `CxexprShortDeclarationShared` |
| 6 | enum_value | `CXEXPR_SHORT_DECLARATION_SHARED_LAST` | `CxexprShortDeclarationSharedLast` |
| 6 | enum_value | `CXEXPR_STRUCT_LITERAL` | `CxexprStructLiteral` |
| 6 | enum_value | `CXEXPR_STRUCT_LITERAL_FIELD` | `CxexprStructLiteralField` |
| 6 | enum | `CXEXPR_TYPE` | `CxexprType` |
| 6 | enum_value | `CXEXPR_UNUSED` | `CxexprUnused` |
| 6 | enum_value | `CX_LINE` | `CXLine` |
| 6 | enum | `CXOPERATION_TYPE` | `CxoperationType` |
| 6 | method | `CXPackage.AddGlobal_TypeSignature` | `GlobalTypeSignatureAdd` |
| 6 | method | `CXStruct.AddField_CXAtomicOps` | `FieldCXAtomicOpsAdd` |
| 6 | method | `CXStruct.AddField_Globals_CXAtomicOps` | `FieldGlobalsCXAtomicOpsAdd` |
| 6 | method | `CXStruct.AddField_TypeSignature` | `FieldTypeSignatureAdd` |
| 6 | enum | `CXTypeSignature_TYPE` | `CXTypeSignatureType` |
| 6 | method | `CXValue.Get_bool` | `BoolGet` |
| 6 | method | `CXValue.Get_bytes` | `BytesGet` |
| 6 | method | `CXValue.Get_f32` | `F32Get` |
| 6 | method | `CXValue.Get_f64` | `F64Get` |
| 6 | method | `CXValue.Get_i16` | `I16Get` |
| 6 | method | `CXValue.Get_i32` | `I32Get` |
| 6 | method | `CXValue.Get_i64` | `I64Get` |
| 6 | method | `CXValue.Get_i8` | `I8Get` |
| 6 | method | `CXValue.Get_ptr` | `PtrGet` |
| 6 | method | `CXValue.GetSlice_bytes` | `SliceBytesGet` |
| 6 | method | `CXValue.GetSlice_f32` | `SliceF32Get` |
| 6 | method | `CXValue.GetSlice_f64` | `SliceF64Get` |
| 6 | method | `CXValue.GetSlice_i16` | `SliceI16Get` |
| 6 | method | `CXValue.GetSlice_i32` | `SliceI32Get` |
| 6 | method | `CXValue.GetSlice_i64` | `SliceI64Get` |
| 6 | method | `CXValue.GetSlice_i8` | `SliceI8Get` |
| 6 | method | `CXValue.GetSlice_ui16` | `SliceUI16Get` |
| 6 | method | `CXValue.GetSlice_ui32` | `SliceUI32Get` |
| 6 | method | `CXValue.GetSlice_ui64` | `SliceUI64Get` |
| 6 | method | `CXValue.GetSlice_ui8` | `SliceUI8Get` |
| 6 | method | `CXValue.Get_str` | `StrGet` |
| 6 | method | `CXValue.Get_str_bytes` | `StrBytesGet` |
| 6 | method | `CXValue.Get_ui16` | `UI16Get` |
| 6 | method | `CXValue.Get_ui32` | `UI32Get` |
| 6 | method | `CXValue.Get_ui64` | `UI64Get` |
| 6 | method | `CXValue.Get_ui8` | `UI8Get` |
| 6 | method | `CXValue.Set_bool` | `BoolSet` |
| 6 | method | `CXValue.Set_bytes` | `BytesSet` |
| 6 | method | `CXValue.Set_f32` | `F32Set` |
| 6 | method | `CXValue.Set_f64` | `F64Set` |
| 6 | method | `CXValue.Set_i16` | `I16Set` |
| 6 | method | `CXValue.Set_i32` | `I32Set` |
| 6 | method | `CXValue.Set_i64` | `I64Set` |
| 6 | method | `CXValue.Set_i8` | `I8Set` |
| 6 | method | `CXValue.Set_ptr` | `PtrSet` |
| 6 | method | `CXValue.Set_str` | `StrSet` |
| 6 | method | `CXValue.Set_str_bytes` | `StrBytesSet` |
| 6 | method | `CXValue.Set_str_concat` | `StrConcatSet` |
| 6 | method | `CXValue.Set_ui16` | `UI16Set` |
| 6 | method | `CXValue.Set_ui32` | `UI32Set` |
| 6 | method | `CXValue.Set_ui64` | `UI64Set` |
| 6 | method | `CXValue.Set_ui8` | `UI8Set` |
| 6 | function | `getFormattedType_CXArg` | `formattedTypeCXArgGet` |
| 6 | enum_value | `TYPE_ARRAY_ATOMIC` | `TypeArrayAtomic` |
| 6 | enum_value | `TYPE_ARRAY_COMPLEX` | `TypeArrayComplex` |
| 6 | enum_value | `TYPE_ARRAY_POINTER_COMPLEX` | `TypeArrayPointerComplex` |
| 6 | enum_value | `TYPE_ARRAY_STRUCT` | `TypeArrayStruct` |
| 6 | enum_value | `TYPE_ATOMIC` | `TypeAtomic` |
| 6 | enum_value | `TYPE_COMPLEX` | `TypeComplex` |
| 6 | enum_value | `TYPE_CXARGUMENT_DEPRECATE` | `TypeCxargumentDeprecate` |
| 6 | enum_value | `TYPE_POINTER_ARRAY_ATOMIC` | `TypePointerArrayAtomic` |
| 6 | enum_value | `TYPE_POINTER_ARRAY_STRUCT` | `TypePointerArrayStruct` |
| 6 | enum_value | `TYPE_POINTER_ATOMIC` | `TypePointerAtomic` |
| 6 | enum_value | `TYPE_POINTER_COMPLEX` | `TypePointerComplex` |
| 6 | enum_value | `TYPE_POINTER_SLICE_ATOMIC` | `TypePointerSliceAtomic` |
| 6 | enum_value | `TYPE_POINTER_STRUCT` | `TypePointerStruct` |
| 6 | enum_value | `TYPE_SLICE_ATOMIC` | `TypeSliceAtomic` |
| 6 | enum_value | `TYPE_SLICE_COMPLEX` | `TypeSliceComplex` |
| 6 | enum_value | `TYPE_SLICE_POINTER_COMPLEX` | `TypeSlicePointerComplex` |
| 6 | enum_value | `TYPE_SLICE_POINTER_STRUCT` | `TypeSlicePointerStruct` |
| 6 | enum_value | `TYPE_SLICE_STRUCT` | `TypeSliceStruct` |
| 6 | enum_value | `TYPE_STRUCT` | `TypeStruct` |
| 6 | enum_value | `TYPE_UNUSED` | `TypeUnused` |
| 7 | field | `CXProgram.SysInitExprs` | `SysInitExpressions` |
| 7 | field | `CXValue.Expr` | `Expression` |
| 7 | field | `DereferenceStruct.idxCounter` | `indexCounter` |

### t3-cx-compiler-llvm

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | struct | `debugInfo` | `debugMetadata` |
| 3 | function | `debugInfoNew` | `debugMetadataNew` |
| 3 | local_variable | `emitter.dataRuns data` | `segment` |
| 3 | local_variable | `emitter.header data` | `constantBytes` |
| 3 | method | `emitter.newTmp` | `localNew` |
| 3 | field | `emitter.tmp` | `localCount` |
| 3 | parameter | `llvmBytes data` | `raw` |
| 7 | field | `emitter.decls` | `declarations` |
| 7 | field | `emitter.fn` | `function` |
| 7 | method | `emitter.retType` | `returnType` |

### t4-actions

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | local_variable | `AddPointer tmp` | `fieldlessSym` |
| 3 | local_variable | `AddPointer tmpIndex` | `fieldlessSymIndex` |
| 3 | loop_variable | `assignStructLiteralFields exprInfo` | `fromExpr` |
| 3 | parameter | `checkMatchParamTypes isInputs` | `inputsChecked` |
| 3 | parameter | `CheckRedeclared symbolsData` | `symbolTable` |
| 3 | parameter | `DeclarationSpecifiersStruct isExternal` | `external` |
| 3 | local_variable | `DeclareImportAt hasSlash` | `slashFound` |
| 3 | parameter | `escapeDeclare symbolsData` | `symbolTable` |
| 3 | local_variable | `FunctionDeclaration symbolsData` | `symbolTable` |
| 3 | parameter | `FunctionHeader isMethod` | `method` |
| 3 | function | `generateTempVarName` | `temporaryVariableNameGenerate` |
| 3 | parameter | `GetGlobalSymbol symbolsData` | `symbolTable` |
| 3 | parameter | `GiveOffset symbolsData` | `symbolTable` |
| 3 | function | `hasDeclSpec` | `declSpecFound` |
| 3 | function | `hasDerefOp` | `derefOpFound` |
| 3 | function | `IsAllArgsBasicTypes` | `ArgsAllBasicTypes` |
| 3 | function | `isParseOp` | `parseOperator` |
| 3 | function | `isPointerAdded` | `pointerAdded` |
| 3 | function | `IsTempVar` | `TemporaryVariable` |
| 3 | parameter | `lookupSymbol symbolsData` | `symbolTable` |
| 3 | parameter | `PostfixExpressionIncDec isInc` | `increment` |
| 3 | function | `processAssignment` | `assignmentResolve` |
| 3 | function | `ProcessExpressionArguments` | `ExpressionArgumentsResolve` |
| 3 | parameter | `ProcessExpressionArguments isInput` | `inputArguments` |
| 3 | local_variable | `ProcessExpressionArguments isLocalVar` | `localVariable` |
| 3 | parameter | `ProcessExpressionArguments symbolsData` | `symbolTable` |
| 3 | function | `ProcessFunctionParameters` | `FunctionParametersDeclare` |
| 3 | parameter | `ProcessFunctionParameters symbolsData` | `symbolTable` |
| 3 | function | `processGlobalInitializer` | `globalInitializerResolve` |
| 3 | function | `ProcessGoTos` | `GoTosResolve` |
| 3 | local_variable | `ProcessGoTos isOpGoto` | `opGoto` |
| 3 | function | `processLocalInitialization` | `localInitializationResolve` |
| 3 | function | `ProcessLoopJumpsOutsideLoop` | `LoopJumpsOutsideLoopReport` |
| 3 | function | `ProcessMethodCall` | `MethodCallResolve` |
| 3 | parameter | `ProcessMethodCall symbolsData` | `symbolTable` |
| 3 | function | `ProcessOperatorExpression` | `OperatorExpressionOutputsSize` |
| 3 | function | `ProcessPointerStructs` | `PointerStructsResolve` |
| 3 | function | `ProcessShortCircuitJumps` | `ShortCircuitJumpsResolve` |
| 3 | function | `ProcessShortDeclaration` | `ShortDeclarationComplete` |
| 3 | function | `ProcessSlice` | `SliceDereferenceSet` |
| 3 | function | `ProcessSliceAssignment` | `SliceAssignmentComplete` |
| 3 | local_variable | `ProcessSliceAssignment isOpIdent` | `opIdent` |
| 3 | function | `ProcessStringAssignment` | `StringAssignmentPassBySet` |
| 3 | local_variable | `ProcessStringAssignment isOpIdent` | `opIdent` |
| 3 | function | `ProcessSymbolFields` | `SymbolFieldsCopy` |
| 3 | function | `ProcessTempVariable` | `TemporaryVariableComplete` |
| 3 | local_variable | `ProcessTempVariable isOpIdent` | `opIdent` |
| 3 | function | `processTestExpression` | `testExpressionResolve` |
| 3 | function | `ProcessTypedOperator` | `TypedOperatorResolve` |
| 3 | parameter | `releaseFunctionSymbols symbolsData` | `symbolTable` |
| 3 | parameter | `scopeShadowCheck symbolsData` | `symbolTable` |
| 3 | local_variable | `shortCircuitOperandExpression temp` | `temporary` |
| 3 | struct | `SymbolsData` | `SymbolTable` |
| 3 | struct | `SymbolsDataKey` | `SymbolTableKey` |
| 3 | field | `SymbolsDataKey.SymbolsDataKeyName` | `SymbolTableKeyName` |
| 3 | field | `SymbolsDataKey.SymbolsDataKeyPackage` | `SymbolTableKeyPackage` |
| 3 | parameter | `symbolsScopeClose symbolsData` | `symbolTable` |
| 3 | parameter | `UpdateSymbolsTable symbolsData` | `symbolTable` |
| 3 | parameter | `WritePrimaryExprs isSlice` | `sliceTyped` |
| 4 | function | `AddConstCode` | `ConstCodeAdd` |
| 4 | function | `AddConstI32` | `ConstI32Add` |
| 4 | function | `AddExprsToFunction` | `ExpressionsToFunctionAdd` |
| 4 | function | `AddJmpToReturnExpressions` | `JmpToReturnExpressionsAdd` |
| 4 | function | `AddPointer` | `PointerAdd` |
| 4 | function | `CheckRedeclared` | `RedeclaredCheck` |
| 4 | function | `CheckTypes` | `TypesCheck` |
| 4 | function | `CheckUndValidTypes` | `UndValidTypesCheck` |
| 4 | function | `CopyArgFields` | `ArgFieldsCopy` |
| 4 | function | `CreateSelectionStatement` | `SelectionStatementCreate` |
| 4 | function | `DeclareGlobalInPackage` | `GlobalInPackageDeclare` |
| 4 | function | `DeclareGlobals` | `GlobalsDeclare` |
| 4 | function | `DeclareImport` | `ImportDeclare` |
| 4 | function | `DeclareImportAt` | `ImportAtDeclare` |
| 4 | function | `DeclareImportScoped` | `ImportScopedDeclare` |
| 4 | function | `DeclareLocal` | `LocalDeclare` |
| 4 | function | `DeclarePackage` | `PackageDeclare` |
| 4 | function | `DeclareStruct` | `StructDeclare` |
| 4 | function | `GetGlobalSymbol` | `GlobalSymbolGet` |
| 4 | function | `InitConstants` | `ConstantsInit` |
| 4 | function | `MakeCXLineExpressionCurrent` | `CXLineExpressionCurrentMake` |
| 4 | function | `UpdateSymbolsTable` | `SymbolsTableUpdate` |
| 4 | function | `WritePrimary` | `PrimaryWrite` |
| 4 | function | `WritePrimaryExprs` | `PrimaryExpressionsWrite` |
| 6 | enum_value | `SEL_ELSEIF` | `SelElseif` |
| 6 | enum_value | `SEL_ELSEIFELSE` | `SelElseifelse` |
| 6 | enum_value | `SEL_RESERVED` | `SelReserved` |
| 7 | function | `checkMatchExprNumOfInputs` | `checkMatchExpressionNumberOfInputs` |
| 7 | function | `checkMatchExprNumOfOutputs` | `checkMatchExpressionNumberOfOutputs` |
| 7 | field | `escapeState.escapeStatePrgrm` | `escapeStateProgram` |
| 7 | function | `exprLine` | `expressionLine` |
| 7 | function | `invalidPlaceholderTypeSig` | `invalidPlaceholderTypeSignature` |
| 7 | field | `shortDeclarationType.typ` | `type` |

### t4-cipher

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | parameter | `cipherBytesOutput data` | `content` |
| 3 | parameter | `cipherSliceOutput data` | `content` |
| 4 | function | `RegisterPackage` | `PackageRegister` |
| 4 | function | `RegisterPackage` | `PackageRegister` |
| 5 | struct | `CipherHashTable` | `HashTable` |
| 5 | function | `CipherSignatureVerify` | `SignatureVerify` |

### t4-cx-compiler-lower

| Step | Kind | Old | New |
|---|---|---|---|
| 7 | field | `state.statePrgrm` | `stateProgram` |

### t4-cxnet

| Step | Kind | Old | New |
|---|---|---|---|
| 4 | function | `InitConstants` | `ConstantsInit` |
| 4 | function | `RegisterPackage` | `PackageRegister` |

### t4-cx-runtime-printer

| Step | Kind | Old | New |
|---|---|---|---|
| 5 | function | `PrinterCompileErrorAppend` | `CompileErrorAppend` |
| 5 | function | `PrinterCompileErrorsPrint` | `CompileErrorsPrint` |
| 5 | function | `PrinterCompileErrorsRegister` | `CompileErrorsRegister` |
| 5 | function | `PrinterErrorHeader` | `ErrorHeader` |
| 5 | function | `PrinterOpDebugPrintStack` | `OpDebugPrintStack` |
| 5 | function | `PrinterRuntimeErrorPrint` | `RuntimeErrorPrint` |
| 5 | function | `PrinterStackPrint` | `StackPrint` |
| 5 | function | `PrinterVmErrorPrint` | `VmErrorPrint` |
| 5 | function | `PrinterVmStopMessage` | `VmStopMessage` |
| 5 | function | `PrinterVmTestFailurePrint` | `VmTestFailurePrint` |

### t4-execute

| Step | Kind | Old | New |
|---|---|---|---|
| 4 | function | `MakeCall` | `CallMake` |
| 4 | function | `RunCompiled` | `CompiledRun` |
| 4 | function | `RunCxAst` | `CXASTRun` |
| 4 | function | `RunPrepare` | `PrepareRun` |

### t4-opcodes

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | local_variable | `assert isStr` | `strTyped` |

### t4-regexp

| Step | Kind | Old | New |
|---|---|---|---|
| 4 | function | `RegisterPackage` | `PackageRegister` |
| 4 | function | `RegisterPackage` | `PackageRegister` |

### t5-cxos

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | local_variable | `OpOsReadF32Slice outputSliceData` | `outputSliceElements` |
| 3 | local_variable | `OpOsReadF64Slice outputSliceData` | `outputSliceElements` |
| 3 | local_variable | `OpOsReadI16Slice outputSliceData` | `outputSliceElements` |
| 3 | local_variable | `OpOsReadI32Slice outputSliceData` | `outputSliceElements` |
| 3 | local_variable | `OpOsReadI64Slice outputSliceData` | `outputSliceElements` |
| 3 | local_variable | `OpOsReadI8Slice outputSliceData` | `outputSliceElements` |
| 3 | local_variable | `OpOsReadUI16Slice outputSliceData` | `outputSliceElements` |
| 3 | local_variable | `OpOsReadUI32Slice outputSliceData` | `outputSliceElements` |
| 3 | local_variable | `OpOsReadUI64Slice outputSliceData` | `outputSliceElements` |
| 3 | local_variable | `OpOsReadUI8Slice outputSliceData` | `outputSliceElements` |
| 3 | local_variable | `OpOsWriteF32Slice data` | `elements` |
| 3 | local_variable | `OpOsWriteF64Slice data` | `elements` |
| 3 | local_variable | `OpOsWriteI16Slice data` | `elements` |
| 3 | local_variable | `OpOsWriteI32Slice data` | `elements` |
| 3 | local_variable | `OpOsWriteI64Slice data` | `elements` |
| 3 | local_variable | `OpOsWriteI8Slice data` | `elements` |
| 3 | local_variable | `OpOsWriteUI16Slice data` | `elements` |
| 3 | local_variable | `OpOsWriteUI32Slice data` | `elements` |
| 3 | local_variable | `OpOsWriteUI64Slice data` | `elements` |
| 3 | local_variable | `OpOsWriteUI8Slice data` | `elements` |
| 4 | function | `InitConstants` | `ConstantsInit` |
| 4 | function | `InitConstants` | `ConstantsInit` |
| 4 | function | `RegisterPackage` | `PackageRegister` |
| 4 | function | `RegisterPackage` | `PackageRegister` |
| 6 | enum_value | `OS_RUN_EMPTY_CMD` | `OsRunEmptyCmd` |
| 6 | enum_value | `OS_RUN_PANIC` | `OsRunPanic` |
| 6 | enum_value | `OS_RUN_START_FAILED` | `OsRunStartFailed` |
| 6 | enum_value | `OS_RUN_SUCCESS` | `OsRunSuccess` |
| 6 | enum_value | `OS_RUN_TIMEOUT` | `OsRunTimeout` |
| 6 | enum_value | `OS_RUN_WAIT_FAILED` | `OsRunWaitFailed` |
| 6 | enum_value | `OS_SEEK_CUR` | `OsSeekCurrent` |
| 6 | enum_value | `OS_SEEK_END` | `OsSeekEnd` |
| 6 | enum_value | `OS_SEEK_SET` | `OsSeekSet` |

### t5-cxpartialparsing

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | function | `isDecimal` | `runeDecimal` |
| 3 | function | `isHex` | `runeHex` |
| 3 | function | `isLetter` | `runeLetter` |
| 4 | function | `NewLexer` | `LexerNew` |
| 7 | field | `Lexer.b` | `b` |
| 7 | field | `Lexer.c` | `c` |
| 7 | field | `Lexer.e` | `e` |
| 7 | field | `Lexer.l` | `l` |
| 7 | field | `Lexer.r` | `r` |

### t5-cx-runtime-core

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | local_variable | `CoreTrailerRead info` | `fileStat` |
| 5 | function | `CoreLibrariesActive` | `LibrariesActive` |
| 5 | function | `CoreLibrariesEnabledResolve` | `LibrariesEnabledResolve` |
| 5 | function | `CoreLibrariesMissing` | `LibrariesMissing` |
| 5 | function | `CoreLibraryNames` | `LibraryNames` |
| 5 | function | `CoreOpcodeLinked` | `OpcodeLinked` |
| 5 | function | `CoreOpcodeName` | `OpcodeName` |
| 5 | function | `CoreOpcodesMissing` | `OpcodesMissing` |
| 5 | function | `CoreOpcodesUsed` | `OpcodesUsed` |
| 5 | function | `CoreOperationRefusalLine` | `OperationRefusalLine` |
| 5 | function | `CoreOperationRefused` | `OperationRefused` |
| 5 | function | `CoreProgramMake` | `ProgramMake` |
| 5 | function | `CoreRegister` | `Register` |
| 5 | struct | `CoreTrailer` | `Trailer` |
| 5 | function | `CoreTrailerAppend` | `TrailerAppend` |
| 5 | function | `CoreTrailerDecode` | `TrailerDecode` |
| 5 | constant | `CoreTrailerFormatVersion` | `TrailerFormatVersion` |
| 5 | constant | `CoreTrailerMagic` | `TrailerMagic` |
| 5 | function | `CoreTrailerRead` | `TrailerRead` |
| 5 | constant | `CoreTrailerSize` | `TrailerSize` |

### t5-plfileoutput

| Step | Kind | Old | New |
|---|---|---|---|
| 4 | function | `GetImportFiles` | `ImportFilesGet` |

### t6-cxfx

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | field | `CxfxTexturePixels.CxfxTexturePixelsData` | `CxfxTexturePixelsValues` |
| 3 | parameter | `cxglBufferData data` | `buffer` |
| 3 | parameter | `cxglBufferSubData data` | `buffer` |
| 3 | parameter | `decodeHdr data` | `buffer` |
| 3 | result | `decodeHdr data` | `pixelBytes` |
| 3 | result | `decodeImg data` | `pixelBytes` |
| 3 | local_variable | `OpAlBufferData data` | `samples` |
| 3 | local_variable | `OpAlLoadWav data` | `encoded` |
| 3 | local_variable | `uploadTexture data` | `pixelBytes` |
| 4 | function | `InitConstants` | `ConstantsInit` |
| 4 | function | `InitConstants` | `ConstantsInit` |
| 4 | function | `PushEvent` | `EventPush` |
| 4 | function | `PushFramebufferSizeEvent` | `FramebufferSizeEventPush` |
| 4 | function | `PushKeyboardEvent` | `KeyboardEventPush` |
| 4 | function | `PushMouseEvent` | `MouseEventPush` |
| 4 | function | `PushWindowPositionEvent` | `WindowPositionEventPush` |
| 4 | function | `PushWindowSizeEvent` | `WindowSizeEventPush` |
| 5 | struct | `CxfxCallback` | `Callback` |
| 5 | constant | `CxfxCallbackI32I32CountMax` | `CallbackI32I32CountMax` |
| 5 | constant | `CxfxCallbackInputCountMax` | `CallbackInputCountMax` |
| 5 | constant | `CxfxFontCountMax` | `FontCountMax` |
| 5 | constant | `CxfxGifCountMax` | `GifCountMax` |
| 5 | struct | `CxfxShaderSourceArgument` | `ShaderSourceArgument` |
| 5 | struct | `CxfxState` | `State` |
| 5 | constant | `CxfxTextBufferSizeInitial` | `TextBufferSizeInitial` |
| 5 | struct | `CxfxTexturePixels` | `TexturePixels` |
| 5 | constant | `CxfxTexturePixelsCountMax` | `TexturePixelsCountMax` |
| 5 | constant | `CxfxWindowCountMax` | `WindowCountMax` |
| 6 | constant | `ACTION_MOVE` | `ActionMove` |
| 6 | constant | `ACTION_PRESS` | `ActionPress` |
| 6 | enum_value | `ACTION_RELEASE` | `ActionRelease` |
| 6 | constant | `ACTION_REPEAT` | `ActionRepeat` |
| 6 | enum_value | `APP_CURSOR_POS` | `AppCursorPos` |
| 6 | enum_value | `APP_EVENT_COUNT` | `AppEventCount` |
| 6 | enum_value | `APP_FOCUS_OFF` | `AppFocusOff` |
| 6 | enum_value | `APP_FOCUS_ON` | `AppFocusOn` |
| 6 | enum_value | `APP_FRAMEBUFFER_SIZE` | `AppFramebufferSize` |
| 6 | enum_value | `APP_KEYBOARD` | `AppKeyboard` |
| 6 | enum_value | `APP_MOUSE` | `AppMouse` |
| 6 | enum_value | `APP_MOUSE_BUTTON` | `AppMouseButton` |
| 6 | enum_value | `APP_NONE` | `AppNone` |
| 6 | enum_value | `APP_PAINT` | `AppPaint` |
| 6 | enum_value | `APP_START` | `AppStart` |
| 6 | enum_value | `APP_STOP` | `AppStop` |
| 6 | enum_value | `APP_WINDOW_POSITION` | `AppWindowPosition` |
| 6 | enum_value | `APP_WINDOW_SIZE` | `AppWindowSize` |
| 6 | enum_value | `HDR_32_RLE_RGBE` | `HDR32RLERGBE` |
| 6 | enum_value | `HDR_NONE` | `HDRNone` |
| 6 | function | `Slice_f32_ToPtr` | `SliceF32ToPtr` |
| 6 | function | `Slice_i32_ToPtr` | `SliceI32ToPtr` |
| 6 | function | `Slice_ui32_ToPtr` | `SliceUI32ToPtr` |
| 6 | function | `Slice_ui8_ToPtr` | `SliceUI8ToPtr` |

### t6-declextractor

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | local_variable | `ExtractEnums tmp` | `enumDeclaration` |
| 3 | local_variable | `ExtractGlobals tmp` | `globalDeclaration` |
| 3 | local_variable | `ExtractImports tmp` | `importDeclaration` |
| 3 | parameter | `scanLinesWithLineTerminator data` | `buffer` |
| 4 | function | `ExtractAllDeclarations` | `AllDeclarationsExtract` |
| 4 | function | `ExtractEnums` | `EnumsExtract` |
| 4 | function | `ExtractFuncs` | `FuncsExtract` |
| 4 | function | `ExtractGlobals` | `GlobalsExtract` |
| 4 | function | `ExtractImports` | `ImportsExtract` |
| 4 | function | `ExtractStructs` | `StructsExtract` |
| 4 | function | `ExtractTypeDefinitions` | `TypeDefinitionsExtract` |
| 4 | function | `GetDeclarations` | `DeclarationsGet` |

### t7-cxmobile

| Step | Kind | Old | New |
|---|---|---|---|
| 4 | function | `CopyAssetsTo` | `AssetsToCopy` |
| 4 | function | `CopyAssetsToFilesDir` | `AssetsToFilesDirCopy` |
| 4 | function | `CopyAssetTo` | `AssetToCopy` |

### t7-init

| Step | Kind | Old | New |
|---|---|---|---|
| 4 | function | `MakeProgram` | `ProgramMake` |
| 4 | function | `RegisterPackages` | `PackagesRegister` |

### t8-cxparsingcompletor

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | function | `isDecimal` | `runeDecimal` |
| 3 | function | `isHex` | `runeHex` |
| 3 | function | `isLetter` | `runeLetter` |
| 3 | local_variable | `Tokenize tmpOutput` | `lineTokens` |
| 4 | function | `InitCXCore` | `CXCoreInit` |
| 4 | method | `Lexer.PrintLex` | `LexPrint` |
| 4 | function | `NewLexer` | `LexerNew` |
| 7 | field | `Lexer.b` | `b` |
| 7 | field | `Lexer.c` | `c` |
| 7 | field | `Lexer.e` | `e` |
| 7 | field | `Lexer.l` | `l` |
| 7 | field | `Lexer.r` | `r` |

### t8-typechecker

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | local_variable | `ParseDeclarationSpecifier dataType` | `typeWord` |
| 3 | local_variable | `ParseDeclarationSpecifier splitDataType` | `typeWordParts` |
| 4 | function | `GetSourceBytes` | `SourceBytesGet` |
| 4 | function | `ParseAllDeclarations` | `AllDeclarationsParse` |
| 4 | function | `ParseDeclarationSpecifier` | `DeclarationSpecifierParse` |
| 4 | function | `ParseFuncHeaders` | `FuncHeadersParse` |
| 4 | function | `ParseFuncParameters` | `FuncParametersParse` |
| 4 | function | `ParseGlobals` | `GlobalsParse` |
| 4 | function | `ParseImports` | `ImportsParse` |
| 4 | function | `ParseParameterDeclaration` | `ParameterDeclarationParse` |
| 4 | function | `ParseStructs` | `StructsParse` |

### t9-astapi

| Step | Kind | Old | New |
|---|---|---|---|
| 4 | function | `AddEmptyFunctionToPackage` | `EmptyFunctionToPackageAdd` |
| 4 | function | `AddEmptyPackage` | `EmptyPackageAdd` |
| 4 | function | `AddLiteralInputToExpression` | `LiteralInputToExpressionAdd` |
| 4 | function | `AddNativeExpressionToFunction` | `NativeExpressionToFunctionAdd` |
| 4 | function | `AddNativeExpressionToFunctionByLineNumber` | `NativeExpressionToFunctionByLineNumberAdd` |
| 4 | function | `AddNativeInputToExpression` | `NativeInputToExpressionAdd` |
| 4 | function | `AddNativeInputToFunction` | `NativeInputToFunctionAdd` |
| 4 | function | `AddNativeOutputToExpression` | `NativeOutputToExpressionAdd` |
| 4 | function | `AddNativeOutputToFunction` | `NativeOutputToFunctionAdd` |
| 4 | function | `FindFunction` | `FunctionFind` |
| 4 | function | `FindPackage` | `PackageFind` |
| 4 | function | `GetAccessibleArgsForFunctionByType` | `AccessibleArgsForFunctionByTypeGet` |
| 4 | function | `GetPackagesNameList` | `PackagesNameListGet` |
| 4 | function | `MakeInputExpressionAPointer` | `InputExpressionAPointerMake` |
| 4 | function | `MakeOutputExpressionAPointer` | `OutputExpressionAPointerMake` |
| 4 | function | `RemoveExpressionFromFunction` | `ExpressionFromFunctionRemove` |
| 4 | function | `RemoveFunctionFromPackage` | `FunctionFromPackageRemove` |

### t9-cxparsing

| Step | Kind | Old | New |
|---|---|---|---|
| 4 | function | `AddInitFunction` | `InitFunctionAdd` |
| 4 | function | `ParseSourceStrings` | `SourceStringsParse` |

### t9-token-fuzzer

| Step | Kind | Old | New |
|---|---|---|---|
| 3 | function | `isDecimal` | `runeDecimal` |
| 3 | function | `isLetter` | `runeLetter` |
