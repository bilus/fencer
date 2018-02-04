package feature

import (
	"github.com/bilus/rtreego"
)

type Key interface {
	ActsAsFeatureKey()
}

type Feature interface {
	rtreego.Spatial
	Contains(point rtreego.Point) (bool, error)
	Key() Key
}
