package glc

import (
	"fmt"
	"runtime"
)

//go:generate bash -c "./generate.sh >encoder.go && gofmt -w encoder.go"

func lastID() (uint64, bool) {
	//return slowlastID()
	//return fastlastID()
	//return fasterlastID()
	return fastestlastID()
}

func fastestlastID() (uint64, bool) {
	var pcs [10000]uintptr
	count := runtime.Callers(0, pcs[:])
	stack := pcs[:count]

	for i, pc := range stack {
		// This if statement below is an optimization very closely tied to the
		// final binary size of the encend function. Therefore, it depends heavily
		// on the content of the function and the architecture for which it has
		// been compiled.
		//
		// We're doing a quick check to see if the pc is plausibly within the range
		// of the encend function before calling FuncForPC, which is quite expensive.
		if pc >= encendpc && pc < encendpc+35 {
			e := runtime.FuncForPC(pc).Entry()
			if e == encendpc {
				var value uint64
				for _, pc := range stack[i:] {
					e := runtime.FuncForPC(pc).Entry()
					if e == encstartpc {
						return value, true
					}
					v, ok := valForPC(e)
					if !ok {
						// Non-encoding interim program counter
						continue
					}
					value <<= 8
					value |= uint64(v)
				}
			} else {
				e := runtime.FuncForPC(pc).Entry()
				panic(fmt.Sprintf("EXPENSIVE! Expected encendpc(%d) but got %d\n", encendpc, e))
			}
		}
	}
	return 0, false
}

func fasterlastID() (uint64, bool) {
	var pcs [10000]uintptr
	count := runtime.Callers(0, pcs[:])
	stack := pcs[:count]

	for i, pc := range stack {
		e := runtime.FuncForPC(pc).Entry()
		if e == encendpc {
			var value uint64
			for _, pc := range stack[i:] {
				e := runtime.FuncForPC(pc).Entry()
				if e == encstartpc {
					return value, true
				}
				v, ok := valForPC(e)
				if !ok {
					// Non-encoding interim program counter
					continue
				}
				value <<= 8
				value |= uint64(v)
			}
		}
	}
	return 0, false
}

func fastlastID() (uint64, bool) {
	var pcs [10000]uintptr
	count := runtime.Callers(0, pcs[:])
	stack := pcs[:count]

	for i, pc := range stack {
		e := runtime.FuncForPC(pc).Entry()
		if e == encendpc {
			var value uint64
			for _, pc := range stack[i:] {
				e := runtime.FuncForPC(pc).Entry()
				if e == encstartpc {
					return value, true
				}
				// A switch will be more efficient, but not sure how much, or
				// if it's worth it. We should benchmark this.
				v, ok := encmap[e]
				if !ok {
					// Non-encoding interim program counter
					continue
				}
				value <<= 8
				value |= uint64(v)
			}
		}
	}
	return 0, false
}

func slowlastID() (uint64, bool) {
	var pcs [10000]uintptr
	count := runtime.Callers(0, pcs[:])
	stack := pcs[:count]
	frames := runtime.CallersFrames(stack)
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		if frame.Entry == encendpc {
			var value uint64
			for frame, more := frames.Next(); more; frame, more = frames.Next() {
				if frame.Entry == encstartpc {
					return value, true
				}
				// A switch will be more efficient, but not sure how much, or
				// if it's worth it. We should benchmark this.
				v, ok := encmap[frame.Entry]
				if !ok {
					return 0, false
				}
				value <<= 8
				value |= uint64(v)
			}
		}
	}
	return 0, false
}
