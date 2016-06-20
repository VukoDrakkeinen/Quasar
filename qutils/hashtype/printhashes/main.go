package main

import (
	"fmt"
	"github.com/VukoDrakkeinen/Quasar/core"
	"github.com/VukoDrakkeinen/Quasar/qutils/hashtype"
)

func main() {
	fmt.Printf("Hash of ComicInfo: %v\n", hashtype.Struct(core.ComicInfo{}))
	fmt.Printf("Hash of ChapterScanlation: %v\n", hashtype.Struct(core.ChapterScanlation{}))
	//fmt.Printf("Hash of updateInfoBridged: %v\n", HashStructType(updateInfoBridged{}))
	//HashStructType(&updateInfoBridged{})
}
