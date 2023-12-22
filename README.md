# glc

GLC is a demonstration of a method of implementing dynamically-scoped variables. In the case of GLC, there is only one variable available, which is a `context.Context`.

For the purposes of this document, the callstack grows upwards, meaning if function `A` calls function `B`, `A` is lower on the callstack, and `B` is higher.

## Dynamic variables
While this repo does not implement generic dynamic variables, it demonstrates the concept, and implements a single dynamic variable of type `context.Context`, which is a good use-case for dynamic variables in Go.

Dynamically-scoped variables are variables whose bindings are passed down through the call stack.
If function `A` declares a dynamic variable `X` and then calls function `B`, which calls function `C`, `C` is able to see `X`, even though it has not been passed to `C`.

### Bindings
In languages that deal in dynamic variables, it's common to distinguish between a name for a variable and the thing the variable holds.

For instance:
```go
	var x int
	x = 10
```

In this situation, x is a variable, and it is bound to 10. We can contrast this with this next example:
```go
	x := NewQueue()
	x.add(something)
```

In this example, x is bound to a queue, the exact value of which is not relevant to the example.
When we call `add` on x, it does not change x`s binding. x's binding has not changed, even if the value it is bound to has been mutated.


### Dynamic Variables and Mutation

For dynamic variables, mutability is usually undesirable, meaning dynamic variables should usually be read-only. If a function `C` wants to change the state of a dynamic variable `X` for some function `D` which it calls, it should create a new instance of X and bind it to X, rather than modifying X itself.

The reason mutability is undesirable is that we want functions lower on the callstack to be able to communicate to functions that they call indirectly, higher on the callstack, but we don't want those higher functions to be able to communicate back down the callstack, which could affect the behavior of the lower functions.

A great imaginary use-case is directing output. If `os.Stdout` were a dynamic variable, we could have a function `A` which bound `os.Stdout` to some file, and then all of the functions which `A` called, or the functions those functions called, etc. would see `os.Stdout` bound to that file, but no other goroutines would see this, and once `A` returned, the binding would die.

Other functions which may be indirectly called may re-bind `os.Stdout` themselves, but we would not want that to propagate up the callstack. This is why immutability is important.

If a function wants to alter the value of a dynamic variable `X`, it should re-bind `X` rather than modifying `X`.

## This Repo

This repository implements a single dynamic variable, which is of type `contex.Context`.

The variable's binding is changed with the `WithContext` function, which takes a `context.Context` (`ctx`) and a `func()` (`f`) and executes `f` with the dynamic variable bound to `ctx`. 

The variable's value can be looked up with the `GetContext` function.

Any function may call another function using `WithContext`, and functions higher on the stack can then use `GetContext` to retrieve the `ctx` value. Those functions may themselves call `WithContext`, which will change the value of `ctx` for the functions they call, without affecting the binding of `ctx` for functions lower on the stack.

For instance:
```go
func a() {
	ctxa := context.Background()
	WithContext(ctxa, b)
	// We see no modifications to the binding of ctx by functions we call.
	ctx.Value("foo") == nil // true
}

func b() {
	ctx := GetContext() // ctx is ctxa from a().
	WithContext(context.WithValue(ctx, "foo", "bar"), c)
	// ctx is unchanged. We see no changes from c.
}

func c() {
	ctx := GetContext() // ctx is ctx from b().
	v := ctx.Value("foo").(string)
	v == "bar" // true
}
```

## Implementation

This library implements dynamically-scoped variables by encoding a reference to the variable into the callstack.

Each variable the library is asked to bind is given a unique ID, of 64 bits. This is large enough that is will not be reasonably exhausted, since it would take many centuries for a program running non-stop to exhaust the 64-bit IDs.

This encoding is done by calling `WithContext`, which calls several functions in succession. Each byte of the ID is encoded by adding a function to the callstack, before finally calling the function `f` which will have access to the dynamic variable.

`f` can then call `GetContext`, which will walk back down the stack, looking for calls to these encoding functions, and associating them with byte values. Thereby it is able to determine the ID. `GetContext` then uses the ID to get a `context.Context` value from a map, and return it to `f`.

The encoding naturally allows dynamic binding since `GetContext` will only find the most recent call to `WithContext`.

## Downsides

This has been implemented in a way that should be safe across Go versions. It does not depend upon the `unsafe` package, or upon the layout of Go internals. As such, it should work and continue to work without causing panics.

However, it's not without its downsides.

The method we use to encode data onto the callstack means that we add multiple functions onto the callstack for every call to `WithContext`. These will be visible to users of this library if they experience a panic, and they will appear as useless garbage on the stack.
