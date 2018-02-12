package feature

import (
	"github.com/bilus/fencer/primitives"
	"github.com/bilus/rtreego"
)

type Key interface {
	Show() string
}

type Feature interface {
	rtreego.Spatial
	Contains(point primitives.Point) (bool, error)
	Key() Key
}
