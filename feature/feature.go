package feature

type Key interface {
	ActsAsFeatureKey()
}

type Feature interface {
	Key() Key
}
