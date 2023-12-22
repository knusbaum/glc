# glc

GLC is a demonstration of a method of implementing dynamically-scoped variables in Go. In the case of GLC, there is only one variable available, which is of type `context.Context`.

For the purposes of this document, the callstack grows upwards, meaning if function `A` calls function `B`, `A` is lower on the callstack, and `B` is higher.

## Dynamic variables
While this repo does not implement generic dynamic variables, it demonstrates the concept, and implements a single dynamic variable of type `context.Context`, which is a good use-case for dynamic variables in Go.

Dynamically-scoped variables are variables whose bindings are passed up through the call stack.
If function `A` declares a dynamic variable `X` and then calls function `B`, which calls function `C`, `C` is able to see the binding of `X`, even though it has not been passed to `C`.

### Bindings
In languages that deal in first-class dynamic variables, it's common to distinguish between a name for a variable and the thing the variable refers to. A name 'referring' to a value is called a 'binding'. So if a variable `x` refers to a value `y`, then `x` is bound to `y`.

For instance:
```go
	var x int
	x = 10
```

In this situation, `x` is a variable, and it is bound to the value `10`. We can contrast this with this next example:
```go
	x := NewQueue()
	x.add(something)
```

In this example, `x` is bound to a `Queue`, the exact value of which is not relevant to the example.
When we call `add` on `x`, it does not change `x`'s binding, rather we're calling `add` on the thing `x` is bound to, which is the `Queue`. `x` is still bound to the same `Queue`. `x`'s binding has not changed, even if the value `x` is bound to has been mutated.


### Dynamic Variables and Mutation

For dynamic variables, mutability is usually undesirable, meaning dynamic variables should usually be read-only. If a function `C` wants to change the state of a dynamic variable `X` for some function it wants to call, `D`, it should create a new value and bind it to `X`, rather than modifying the thing `X` is bound to. Then `D` will see the new binding, but anything else with a reference to what `X` is bound to won't see a change.

The reason mutability is undesirable is that we want functions lower on the callstack to be able to communicate to functions higher on the callstack, which they call indirectly, but we don't want those higher functions to be able to communicate back down the callstack, which could affect the behavior of the lower functions.

The reasons we want this may not be obvious, but should be intuitive if we consider the general statement that from the point of a function, we want another function we call to behave how we want it to, but we don't want that to indirectly affect how we behave, or how our caller behaves.

In short, we want to be able to indirectly affect our callees, but not for our callees to indirectly affect us.

#### Example

A great imaginary use-case is directing output. If `os.Stdout` were a dynamic variable, we could have a function `A` which bound `os.Stdout` to some file, and then all of the functions which `A` called, or the functions those functions called, etc. would see `os.Stdout` bound to that file, but no other goroutines would see this, and once `A` returned, the binding would die.

Other functions which may be indirectly called by `A` may re-bind `os.Stdout` themselves, but we would not want that to propagate back down the callstack to `A`. This is why immutability is important.

If `os.Stdout` can be re-bound, but not mutated, then if `A` calls `B` and then calls `C`, any of `B`s modifications to `os.Stdout` cannot be seen by `C`, which is what we want. We want `A` to be able to tell `B` and `C` where to send their output, but not to let `B` or `C` interfere with each other.

So, if a function wants to alter the value of a dynamic variable `X`, it should re-bind `X` rather than modifying `X`. For the purposes of this library, this is not an issue, since `context.Context` variables are immutable, but in general the type of dynamic variables needs to be considered, and immutable types should be preferred.

## This Repo

This repository implements a single unnamed dynamic variable, which is of type `contex.Context`.

The variable's binding is changed with the `WithContext` function, which takes a `context.Context` (`ctx`) and a `func()` (`f`) and executes `f` with the dynamic variable bound to `ctx`. 

The variable's bound value can be retrieved with the `GetContext` function.

Any function may call another function using `WithContext`, and functions higher on the stack can then use `GetContext` to retrieve the `context.Context` value. Those functions may themselves call `WithContext`, which will change the binding of the variable for the functions they call, without affecting the binding for functions lower on the stack.

For instance:
```go
func main() {
	a()
}

func a() {
	ctxa := context.Background()
	WithContext(ctxa, b)
	
    // We see no modifications to the value of ctx by functions we call.
	ctx.Value("foo") == nil // true
	
	// ctx is nil, since we we were not called with WithContext.
	// There was no binding present.
	ctx = GetContext()
}

func b() {
	ctx := GetContext() // ctx is ctxa from a().
	WithContext(context.WithValue(ctx, "foo", "bar"), c)
	// ctx is unchanged. We see no changes from c.
	ctx = GetContext()
	ctx.Value("foo") == nil
	ctx.Value("bar") == nil
}

func c() {
	ctx := GetContext() // ctx is ctx from b().
	v := ctx.Value("foo").(string)
	v == "bar" // true
	WithContext(context.WithValue(ctx, "bar", "baz"), d)
}

func d() {
	...
}
```

## Implementation

This library implements dynamically-scoped variables by encoding a reference to the variable into the callstack.

Each variable the library is asked to bind is given a unique ID of 64 bits. This is large enough that it will not be reasonably exhausted, since it would take many centuries for a program running non-stop to exhaust the 64-bit IDs.

This encoding is done by calling `WithContext`, which calls several functions in succession. Each byte of the ID is encoded by adding a function to the callstack, before finally calling the function `f` which will have access to the dynamic variable.

`f` can then call `GetContext`, which will walk back down the stack, looking for calls to these encoding functions, and associating them with byte values. Thereby it is able to determine the ID. `GetContext` then uses the ID to get a `context.Context` value from a `sync.Map`, and return it to `f`.

The encoding naturally allows dynamic binding since `GetContext` will only find the most recent call to `WithContext`. Subsequent calls to `WithContext` will take precedence over previous calls since `GetContext` only looks for the most recent calls on the stack.

## Downsides

This has been implemented in a way that should be safe across Go versions. It does not depend upon the `unsafe` package, or upon the layout of Go internals. As such, it should work and continue to work without causing panics.

However, it's not without its downsides.

The method we use to encode data onto the callstack means that we add multiple functions onto the callstack for every call to `WithContext`. These will be visible to users of this library if they experience a panic, and they will appear as useless garbage on the stack.

Also, the runtime of `GetContext` is O(n) in the size of the stack. There are benchmarks in this repo which show this not to be a big issue, as the cost is relatively small, but it's still something to consider.

## Irrelevant Notes

`sync.Map` was rewritten to use generics, but during its implementation there weren't any performance gain found, although we were able to eliminate much of the use of unsafe pointers. As such, the main branch uses the standard library's sync.Map, wrapped in a type-safe type rather than the rewritten one.

The type-safe map can still be found in commit `bd1cfe2` for anyone who is interested.
