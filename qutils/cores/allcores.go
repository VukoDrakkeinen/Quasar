// +build !go1.5

package cores

import "runtime"

func UseAll() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
