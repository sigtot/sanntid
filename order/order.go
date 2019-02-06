package order

// TODO: Rename package or move to somewhere else

type Dir int

const (
	Down Dir = iota
	Cab      // Should Cab be a dir?
	Up
)

type Order struct {
	Floor int
	Dir   Dir
}
