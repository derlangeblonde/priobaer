package util

type MaybeInt struct {
	Value int 
	Valid bool
}

func JustInt(i int) MaybeInt {
	return MaybeInt{Value: i, Valid: true}
}

func NoneInt() MaybeInt {
	return MaybeInt{Valid: false}
}

