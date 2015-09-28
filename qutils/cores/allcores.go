// +build !go1.5, !go1.6, !go1.7, !go1.8, !go1.9
// +build go1.4, go1.3, go1.2, go1.1

package cores

import "runtime"

func UseAll() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
