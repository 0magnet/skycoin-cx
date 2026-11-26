package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	sourceargs "github.com/skycoin/skycoin-cx/t0-cx-compiler-sourceargs"
	"github.com/skycoin/skycoin-cx/t0-cx-tools-packagestore"   // thin: only Init and Close; no depth gain
	"github.com/skycoin/skycoin-cx/t1-cx-tools-packageloader"  // thin: only LoadCXProgram; no depth gain
	"github.com/skycoin/skycoin-cx/t2-cx-tools-packageencoder" // thin: only SavePackagesToDisk; no depth gain
	"github.com/skycoin/skycoin-cx/t2-cx-tools-packagegraph"   // thin: only GetImportGraph; no depth gain
	"github.com/skycoin/skycoin-cx/t2-cx-tools-packagetree"    // thin: only GetImportTree; no depth gain
)

func main() {
	loadFlag := flag.Bool("load", false, "OPTION: Load a program to the database, with a given name to load it as and path to the program")
	saveFlag := flag.Bool("save", false, "OPTION: Save a package to disk, with a given name to search on the database and a new directory path to save to")
	treeFlag := flag.Bool("tree", false, "OPTION: Print the import dependency tree for a given program on the database")
	graphFlag := flag.Bool("graph", false, "OPTION: Print the import dependencies for each module in a program on the database")
	nameFlag := flag.String("name", "", "The name of the program to load or save")
	pathFlag := flag.String("path", "", "The path to the program to load or save")
	dbFlag := flag.String("db", "", "The path to the database file, created if missing")
	flag.Parse()

	if *dbFlag == "" {
		log.Fatal("-db is required")
	}
	if err := packagestore.Init(*dbFlag); err != nil {
		log.Fatal(err)
	}
	defer packagestore.Close()

	if flag.NArg() > 0 {
		log.Fatalf("Unexpected arguments %q. Type -help for more information", flag.Args())
	}
	actions := 0
	for _, set := range []bool{*loadFlag, *saveFlag, *treeFlag, *graphFlag} {
		if set {
			actions++
		}
	}
	if actions != 1 {
		log.Fatal("Give exactly one of -load, -save, -tree or -packagegraph. Type -help for more information")
	}
	if *nameFlag == "" {
		log.Fatal("-name is required")
	}

	programName := *nameFlag

	if *treeFlag {
		output, err := packagetree.ImportTreeGet(programName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(output)
		return
	}

	if *graphFlag {
		output, err := packagegraph.ImportGraphGet(programName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(output)
		return
	}

	if *pathFlag == "" {
		log.Fatal("-path is required with -load and -save")
	}
	path := filepath.Clean(*pathFlag)

	if *loadFlag {
		_, sourceCodes, _, err := sourceargs.Parse([]string{path}, true)
		if err != nil {
			log.Fatal(err)
		}
		err = packageloader.CXProgramLoad(programName, sourceCodes)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	if *saveFlag {
		err := packageencoder.PackagesToDiskSave(programName, path)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Fatal("Wrong arguments provided. Type -help for more information")
}
