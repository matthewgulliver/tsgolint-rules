package a

// want "comment is not permitted"
func a() {}

// lintcn:severity error
func b() {}

//go:generate echo hi
func c() {}

//nolint:errcheck
func d() {}

func e() {
	// want "comment is not permitted"
	a()
	b() // want "comment is not permitted"
	/* want "comment is not permitted" */
	c()
	d()
}
