package main

// generate:reset
type st struct {
	i      int
	s      string
	b      bool
	ptrI   *int
	ptrS   string
	ptrB   bool
	st2    st2
	ptrSt2 *st2
	sl1    []int
	ptrSl1 *[]int
	mp1    map[string]string
	ptrMp1 *map[string]string
}

var a int

// generate:reset
type st2 struct {
	i2  int
	st2 string
}
