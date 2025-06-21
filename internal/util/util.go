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

type IDer interface {
	Id() int
}

func IDs[T IDer](items []T) []int {
    out := make([]int, len(items))
    for i, v := range items {
        out[i] = v.Id()
    }
    return out
}
